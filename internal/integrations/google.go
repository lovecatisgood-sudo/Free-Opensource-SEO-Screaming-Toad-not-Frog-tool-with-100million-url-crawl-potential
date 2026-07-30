// Package integrations contains bounded adapters for optional external SEO
// data. Callers must explicitly select an integration and credential reference.
package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/secretstore"
)

type Endpoints struct{ PageSpeed, CrUX, SearchConsole, GA4, OAuthToken string }

func DefaultEndpoints() Endpoints {
	return Endpoints{PageSpeed: "https://www.googleapis.com/pagespeedonline/v5/runPagespeed", CrUX: "https://chromeuxreport.googleapis.com/v1/records:queryRecord", SearchConsole: "https://www.googleapis.com/webmasters/v3/sites/", GA4: "https://analyticsdata.googleapis.com/v1beta/properties/", OAuthToken: "https://oauth2.googleapis.com/token"}
}

type Client struct {
	HTTP      *http.Client
	Secrets   secretstore.Store
	Endpoints Endpoints
	Now       func() time.Time
}

func (c Client) prepare() Client {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 60 * time.Second}
	}
	defaults := DefaultEndpoints()
	if c.Endpoints.PageSpeed == "" {
		c.Endpoints.PageSpeed = defaults.PageSpeed
	}
	if c.Endpoints.CrUX == "" {
		c.Endpoints.CrUX = defaults.CrUX
	}
	if c.Endpoints.SearchConsole == "" {
		c.Endpoints.SearchConsole = defaults.SearchConsole
	}
	if c.Endpoints.GA4 == "" {
		c.Endpoints.GA4 = defaults.GA4
	}
	if c.Endpoints.OAuthToken == "" {
		c.Endpoints.OAuthToken = defaults.OAuthToken
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	return c
}

type Observation[T any] struct {
	Provider       string `json:"provider"`
	EvidenceSource string `json:"evidence_source"`
	ProfileVersion string `json:"profile_version"`
	Scope          string `json:"scope"`
	ObservedAt     string `json:"observed_at"`
	Freshness      string `json:"freshness,omitempty"`
	Data           T      `json:"data"`
}
type PageSpeedResult struct {
	RequestedURL      string             `json:"requested_url"`
	FinalURL          string             `json:"final_url"`
	Strategy          string             `json:"strategy"`
	LighthouseVersion string             `json:"lighthouse_version"`
	FetchTime         string             `json:"fetch_time"`
	Scores            map[string]float64 `json:"scores"`
	Metrics           map[string]float64 `json:"metrics"`
}

func (c Client) PageSpeed(ctx context.Context, target, strategy, keyRef string) (Observation[PageSpeedResult], error) {
	c = c.prepare()
	if strategy == "" {
		strategy = "mobile"
	}
	if strategy != "mobile" && strategy != "desktop" {
		return Observation[PageSpeedResult]{}, errors.New("PageSpeed strategy must be mobile or desktop")
	}
	if err := validatePublicTarget(target); err != nil {
		return Observation[PageSpeedResult]{}, err
	}
	endpoint, err := url.Parse(c.Endpoints.PageSpeed)
	if err != nil {
		return Observation[PageSpeedResult]{}, err
	}
	query := endpoint.Query()
	query.Set("url", target)
	query.Set("strategy", strategy)
	for _, category := range []string{"performance", "accessibility", "best-practices", "seo"} {
		query.Add("category", category)
	}
	if keyRef != "" {
		key, err := c.secret(ctx, keyRef)
		if err != nil {
			return Observation[PageSpeedResult]{}, err
		}
		query.Set("key", string(key))
	}
	endpoint.RawQuery = query.Encode()
	var response struct {
		ID         string `json:"id"`
		Lighthouse struct {
			Version    string `json:"lighthouseVersion"`
			FetchTime  string `json:"fetchTime"`
			FinalURL   string `json:"finalDisplayedUrl"`
			Categories map[string]struct {
				Score float64 `json:"score"`
			} `json:"categories"`
			Audits map[string]struct {
				NumericValue float64 `json:"numericValue"`
			} `json:"audits"`
		} `json:"lighthouseResult"`
	}
	if err := c.call(ctx, http.MethodGet, endpoint.String(), "", nil, &response); err != nil {
		return Observation[PageSpeedResult]{}, err
	}
	result := PageSpeedResult{RequestedURL: target, FinalURL: response.Lighthouse.FinalURL, Strategy: strategy, LighthouseVersion: response.Lighthouse.Version, FetchTime: response.Lighthouse.FetchTime, Scores: map[string]float64{}, Metrics: map[string]float64{}}
	for name, value := range response.Lighthouse.Categories {
		result.Scores[name] = value.Score
	}
	for _, name := range []string{"first-contentful-paint", "largest-contentful-paint", "cumulative-layout-shift", "interaction-to-next-paint", "speed-index", "total-blocking-time"} {
		if value, ok := response.Lighthouse.Audits[name]; ok {
			result.Metrics[name] = value.NumericValue
		}
	}
	return Observation[PageSpeedResult]{Provider: "pagespeed-insights", EvidenceSource: "lab", ProfileVersion: "v5", Scope: target, ObservedAt: c.Now().UTC().Format(time.RFC3339Nano), Data: result}, nil
}

type CrUXRequest struct {
	URL        string `json:"url,omitempty"`
	Origin     string `json:"origin,omitempty"`
	FormFactor string `json:"formFactor,omitempty"`
}
type CrUXResult struct {
	Key              map[string]string          `json:"key"`
	CollectionPeriod map[string]any             `json:"collection_period"`
	Metrics          map[string]json.RawMessage `json:"metrics"`
}

func (c Client) CrUX(ctx context.Context, input CrUXRequest, keyRef string) (Observation[CrUXResult], error) {
	c = c.prepare()
	if (input.URL == "") == (input.Origin == "") {
		return Observation[CrUXResult]{}, errors.New("CrUX requires exactly one URL or origin")
	}
	scope := input.URL
	if scope == "" {
		scope = input.Origin
	}
	if err := validatePublicTarget(scope); err != nil {
		return Observation[CrUXResult]{}, err
	}
	if input.FormFactor != "" && !strings.Contains(" PHONE DESKTOP TABLET ", " "+input.FormFactor+" ") {
		return Observation[CrUXResult]{}, errors.New("invalid CrUX form factor")
	}
	key, err := c.secret(ctx, keyRef)
	if err != nil {
		return Observation[CrUXResult]{}, err
	}
	endpoint, err := url.Parse(c.Endpoints.CrUX)
	if err != nil {
		return Observation[CrUXResult]{}, err
	}
	query := endpoint.Query()
	query.Set("key", string(key))
	endpoint.RawQuery = query.Encode()
	var response struct {
		Record struct {
			Key              map[string]string          `json:"key"`
			CollectionPeriod map[string]any             `json:"collectionPeriod"`
			Metrics          map[string]json.RawMessage `json:"metrics"`
		} `json:"record"`
	}
	if err := c.call(ctx, http.MethodPost, endpoint.String(), "", input, &response); err != nil {
		return Observation[CrUXResult]{}, err
	}
	period, _ := json.Marshal(response.Record.CollectionPeriod)
	freshness := string(period)
	return Observation[CrUXResult]{Provider: "chrome-ux-report", EvidenceSource: "field", ProfileVersion: "v1", Scope: scope, ObservedAt: c.Now().UTC().Format(time.RFC3339Nano), Freshness: freshness, Data: CrUXResult{Key: response.Record.Key, CollectionPeriod: response.Record.CollectionPeriod, Metrics: response.Record.Metrics}}, nil
}

type SearchConsoleRequest struct {
	SiteURL    string   `json:"site_url"`
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	Dimensions []string `json:"dimensions,omitempty"`
	RowLimit   int      `json:"row_limit,omitempty"`
	DataState  string   `json:"data_state,omitempty"`
}
type SearchConsoleRow struct {
	Keys        []string `json:"keys"`
	Clicks      float64  `json:"clicks"`
	Impressions float64  `json:"impressions"`
	CTR         float64  `json:"ctr"`
	Position    float64  `json:"position"`
}
type SearchConsoleResult struct {
	Rows            []SearchConsoleRow `json:"rows"`
	AggregationType string             `json:"response_aggregation_type"`
	DataState       string             `json:"data_state"`
	StartDate       string             `json:"start_date"`
	EndDate         string             `json:"end_date"`
}

func (c Client) SearchConsole(ctx context.Context, input SearchConsoleRequest, tokenRef string) (Observation[SearchConsoleResult], error) {
	c = c.prepare()
	if input.SiteURL == "" || !validDateRange(input.StartDate, input.EndDate) {
		return Observation[SearchConsoleResult]{}, errors.New("Search Console property and valid date range are required")
	}
	if input.RowLimit == 0 {
		input.RowLimit = 1000
	}
	if input.RowLimit < 1 || input.RowLimit > 5000 {
		return Observation[SearchConsoleResult]{}, errors.New("Search Console row limit must be 1 to 5000")
	}
	if len(input.Dimensions) > 5 {
		return Observation[SearchConsoleResult]{}, errors.New("Search Console dimensions are limited to five")
	}
	allowedDimensions := map[string]bool{"country": true, "device": true, "page": true, "query": true, "searchAppearance": true, "date": true, "hour": true}
	for _, dimension := range input.Dimensions {
		if !allowedDimensions[dimension] {
			return Observation[SearchConsoleResult]{}, errors.New("Search Console dimension is unsupported")
		}
	}
	if input.DataState == "" {
		input.DataState = "final"
	}
	if input.DataState != "final" && input.DataState != "all" && input.DataState != "hourly_all" {
		return Observation[SearchConsoleResult]{}, errors.New("Search Console data state is unsupported")
	}
	token, err := c.bearerToken(ctx, tokenRef)
	if err != nil {
		return Observation[SearchConsoleResult]{}, err
	}
	endpoint := c.Endpoints.SearchConsole + url.PathEscape(input.SiteURL) + "/searchAnalytics/query"
	body := map[string]any{"startDate": input.StartDate, "endDate": input.EndDate, "dimensions": input.Dimensions, "rowLimit": input.RowLimit, "dataState": input.DataState}
	var response struct {
		Rows []struct {
			Keys        []string `json:"keys"`
			Clicks      float64  `json:"clicks"`
			Impressions float64  `json:"impressions"`
			CTR         float64  `json:"ctr"`
			Position    float64  `json:"position"`
		} `json:"rows"`
		AggregationType string `json:"responseAggregationType"`
	}
	if err := c.call(ctx, http.MethodPost, endpoint, "Bearer "+string(token), body, &response); err != nil {
		return Observation[SearchConsoleResult]{}, err
	}
	rows := make([]SearchConsoleRow, 0, len(response.Rows))
	for _, row := range response.Rows {
		rows = append(rows, SearchConsoleRow{Keys: row.Keys, Clicks: row.Clicks, Impressions: row.Impressions, CTR: row.CTR, Position: row.Position})
	}
	result := SearchConsoleResult{Rows: rows, AggregationType: response.AggregationType, DataState: input.DataState, StartDate: input.StartDate, EndDate: input.EndDate}
	return Observation[SearchConsoleResult]{Provider: "google-search-console", EvidenceSource: "external_api", ProfileVersion: "webmasters-v3", Scope: input.SiteURL, ObservedAt: c.Now().UTC().Format(time.RFC3339Nano), Freshness: input.DataState, Data: result}, nil
}

type GA4Request struct {
	PropertyID string `json:"property_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Limit      int    `json:"limit,omitempty"`
}
type GA4Result struct {
	Rows       []map[string]string `json:"rows"`
	PropertyID string              `json:"property_id"`
	StartDate  string              `json:"start_date"`
	EndDate    string              `json:"end_date"`
}

func (c Client) GA4(ctx context.Context, input GA4Request, tokenRef string) (Observation[GA4Result], error) {
	c = c.prepare()
	if _, err := strconv.ParseInt(input.PropertyID, 10, 64); err != nil || !validDateRange(input.StartDate, input.EndDate) {
		return Observation[GA4Result]{}, errors.New("GA4 property and valid date range are required")
	}
	if input.Limit == 0 {
		input.Limit = 1000
	}
	if input.Limit < 1 || input.Limit > 10000 {
		return Observation[GA4Result]{}, errors.New("GA4 row limit must be 1 to 10000")
	}
	token, err := c.bearerToken(ctx, tokenRef)
	if err != nil {
		return Observation[GA4Result]{}, err
	}
	endpoint := c.Endpoints.GA4 + url.PathEscape(input.PropertyID) + ":runReport"
	body := map[string]any{"dateRanges": []map[string]string{{"startDate": input.StartDate, "endDate": input.EndDate}}, "dimensions": []map[string]string{{"name": "landingPagePlusQueryString"}}, "metrics": []map[string]string{{"name": "sessions"}, {"name": "activeUsers"}, {"name": "keyEvents"}}, "limit": input.Limit}
	var response struct {
		DimensionHeaders []struct {
			Name string `json:"name"`
		} `json:"dimensionHeaders"`
		MetricHeaders []struct {
			Name string `json:"name"`
		} `json:"metricHeaders"`
		Rows []struct {
			DimensionValues []struct {
				Value string `json:"value"`
			} `json:"dimensionValues"`
			MetricValues []struct {
				Value string `json:"value"`
			} `json:"metricValues"`
		} `json:"rows"`
	}
	if err := c.call(ctx, http.MethodPost, endpoint, "Bearer "+string(token), body, &response); err != nil {
		return Observation[GA4Result]{}, err
	}
	rows := make([]map[string]string, 0, len(response.Rows))
	for _, row := range response.Rows {
		item := map[string]string{}
		for i, value := range row.DimensionValues {
			if i < len(response.DimensionHeaders) {
				item[response.DimensionHeaders[i].Name] = value.Value
			}
		}
		for i, value := range row.MetricValues {
			if i < len(response.MetricHeaders) {
				item[response.MetricHeaders[i].Name] = value.Value
			}
		}
		rows = append(rows, item)
	}
	result := GA4Result{Rows: rows, PropertyID: input.PropertyID, StartDate: input.StartDate, EndDate: input.EndDate}
	return Observation[GA4Result]{Provider: "google-analytics-4", EvidenceSource: "external_api", ProfileVersion: "data-v1beta", Scope: "properties/" + input.PropertyID, ObservedAt: c.Now().UTC().Format(time.RFC3339Nano), Data: result}, nil
}

func (c Client) call(ctx context.Context, method, endpoint, authorization string, input, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		if len(encoded) > 1<<20 {
			return errors.New("integration request exceeds 1 MiB")
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return errors.New("integration request could not be created")
	}
	request.Header.Set("Accept", "application/json")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response, err := c.HTTP.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("integration request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("integration returned HTTP %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 8<<20))
	if err := decoder.Decode(output); err != nil {
		return fmt.Errorf("decode integration response: %w", err)
	}
	return nil
}

type OAuthCredential struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	Expiry       string `json:"expiry,omitempty"`
}

func (c Client) secret(ctx context.Context, reference string) ([]byte, error) {
	if c.Secrets == nil {
		return nil, errors.New("secure credential store is unavailable")
	}
	return c.Secrets.Get(ctx, reference)
}

func (c Client) bearerToken(ctx context.Context, reference string) ([]byte, error) {
	stored, err := c.secret(ctx, reference)
	if err != nil {
		return nil, err
	}
	var credential OAuthCredential
	if json.Unmarshal(stored, &credential) != nil || credential.AccessToken == "" {
		return stored, nil
	}
	expiry, expiryErr := time.Parse(time.RFC3339, credential.Expiry)
	if credential.Expiry == "" || expiryErr != nil || expiry.After(c.Now().Add(time.Minute)) {
		return []byte(credential.AccessToken), nil
	}
	if credential.RefreshToken == "" || credential.ClientID == "" || credential.ClientSecret == "" {
		return nil, errors.New("OAuth credential expired and has no refresh configuration")
	}
	form := url.Values{"client_id": {credential.ClientID}, "client_secret": {credential.ClientSecret}, "refresh_token": {credential.RefreshToken}, "grant_type": {"refresh_token"}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoints.OAuthToken, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.New("OAuth refresh request could not be created")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	response, err := c.HTTP.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("OAuth refresh request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("OAuth refresh returned HTTP %d", response.StatusCode)
	}
	var refreshed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&refreshed); err != nil || refreshed.AccessToken == "" || refreshed.ExpiresIn < 1 {
		return nil, errors.New("OAuth refresh response is invalid")
	}
	credential.AccessToken = refreshed.AccessToken
	credential.Expiry = c.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
	encoded, err := json.Marshal(credential)
	if err != nil {
		return nil, errors.New("OAuth credential could not be encoded")
	}
	if err := c.Secrets.Put(ctx, reference, encoded); err != nil {
		return nil, errors.New("refreshed OAuth credential could not be stored")
	}
	return []byte(refreshed.AccessToken), nil
}

func validatePublicTarget(value string) error {
	parsed, err := url.Parse(value)
	host := strings.ToLower(parsed.Hostname())
	address := net.ParseIP(host)
	if err != nil || !strings.Contains(" http https ", " "+parsed.Scheme+" ") || host == "" || parsed.User != nil || len(value) > 8192 || host == "localhost" || strings.HasSuffix(host, ".local") || (address != nil && (address.IsPrivate() || address.IsLoopback() || address.IsUnspecified() || address.IsLinkLocalUnicast())) {
		return errors.New("integration target must be a public HTTP(S) URL without credentials")
	}
	return nil
}
func validDateRange(start, end string) bool {
	first, e1 := time.Parse("2006-01-02", start)
	last, e2 := time.Parse("2006-01-02", end)
	return e1 == nil && e2 == nil && !last.Before(first) && last.Sub(first) <= 16*31*24*time.Hour
}
