package fetchpolicy

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type FetchLimits struct {
	TotalTimeout            time.Duration
	MaximumRedirects        int
	MaximumResponseHeaders  int
	MaximumCompressedBytes  int64
	MaximumDecodedBytes     int64
	MaximumCompressionRatio float64
}

func DefaultFetchLimits() FetchLimits {
	return FetchLimits{
		TotalTimeout:            45 * time.Second,
		MaximumRedirects:        10,
		MaximumResponseHeaders:  200,
		MaximumCompressedBytes:  10 << 20,
		MaximumDecodedBytes:     25 << 20,
		MaximumCompressionRatio: 100,
	}
}

func (l FetchLimits) validate() error {
	if l.TotalTimeout <= 0 || l.TotalTimeout > 10*time.Minute {
		return errors.New("total timeout is outside supported range")
	}
	if l.MaximumRedirects < 0 || l.MaximumRedirects > 50 {
		return errors.New("redirect limit is outside supported range")
	}
	if l.MaximumResponseHeaders < 1 || l.MaximumResponseHeaders > 1000 {
		return errors.New("header limit is outside supported range")
	}
	if l.MaximumCompressedBytes < 1 || l.MaximumDecodedBytes < 1 {
		return errors.New("response byte limits are invalid")
	}
	if l.MaximumCompressionRatio < 1 || l.MaximumCompressionRatio > 1000 {
		return errors.New("compression ratio limit is invalid")
	}
	return nil
}

type RedirectEvidence struct {
	Hop        int
	SourceURL  string
	StatusCode int
	TargetURL  string
}

type FetchResult struct {
	RequestedURL    string
	FinalURL        string
	StatusCode      int
	Header          http.Header
	Body            []byte
	ContentType     string
	CompressedBytes int64
	DecodedBytes    int64
	Redirects       []RedirectEvidence
	StartedAt       time.Time
	FinishedAt      time.Time
}

type Fetcher struct {
	guard     TargetValidator
	client    *http.Client
	limits    FetchLimits
	userAgent string
}

func NewFetcher(guard TargetValidator, transport http.RoundTripper, limits FetchLimits, userAgent string) (*Fetcher, error) {
	if guard == nil || transport == nil {
		return nil, errors.New("target guard and transport are required")
	}
	if err := limits.validate(); err != nil {
		return nil, err
	}
	if userAgent == "" || len(userAgent) > 256 || strings.ContainsAny(userAgent, "\r\n") {
		return nil, errors.New("user agent is invalid")
	}
	return &Fetcher{
		guard: guard,
		client: &http.Client{
			Transport: transport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		limits: limits, userAgent: userAgent,
	}, nil
}

func (f *Fetcher) Fetch(ctx context.Context, raw string) (FetchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, f.limits.TotalTimeout)
	defer cancel()
	result := FetchResult{RequestedURL: raw, StartedAt: time.Now().UTC()}
	current := raw
	for hop := 0; ; hop++ {
		target, err := f.guard.Validate(ctx, current)
		if err != nil {
			return FetchResult{}, fmt.Errorf("target validation: %w", err)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL.String(), nil)
		if err != nil {
			return FetchResult{}, fmt.Errorf("create request: %w", err)
		}
		request.Header.Set("User-Agent", f.userAgent)
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,text/plain;q=0.8,*/*;q=0.1")
		request.Header.Set("Accept-Encoding", "gzip")
		response, err := f.client.Do(request)
		if err != nil {
			return FetchResult{}, fmt.Errorf("perform request: %w", err)
		}
		if headerCount(response.Header) > f.limits.MaximumResponseHeaders {
			_ = response.Body.Close()
			return FetchResult{}, errors.New("response header count exceeds limit")
		}
		if isRedirect(response.StatusCode) {
			location := response.Header.Get("Location")
			_ = response.Body.Close()
			if location == "" {
				return FetchResult{}, errors.New("redirect response has no location")
			}
			if hop >= f.limits.MaximumRedirects {
				return FetchResult{}, errors.New("redirect limit reached")
			}
			next, err := resolveRedirect(target.URL, location)
			if err != nil {
				return FetchResult{}, err
			}
			result.Redirects = append(result.Redirects, RedirectEvidence{
				Hop: hop, SourceURL: target.RequestKey, StatusCode: response.StatusCode, TargetURL: next,
			})
			current = next
			continue
		}
		body, compressed, decoded, err := f.readBody(response)
		if err != nil {
			return FetchResult{}, err
		}
		result.FinalURL = target.RequestKey
		result.StatusCode = response.StatusCode
		result.Header = response.Header.Clone()
		result.Body = body
		result.ContentType = response.Header.Get("Content-Type")
		result.CompressedBytes = compressed
		result.DecodedBytes = decoded
		result.FinishedAt = time.Now().UTC()
		return result, nil
	}
}

func (f *Fetcher) readBody(response *http.Response) ([]byte, int64, int64, error) {
	defer response.Body.Close()
	compressed := &countingReader{reader: io.LimitReader(response.Body, f.limits.MaximumCompressedBytes+1)}
	var decodedReader io.Reader = compressed
	encoding := strings.ToLower(strings.TrimSpace(response.Header.Get("Content-Encoding")))
	switch encoding {
	case "", "identity":
	case "gzip":
		gzipReader, err := gzip.NewReader(compressed)
		if err != nil {
			return nil, compressed.count, 0, fmt.Errorf("open gzip response: %w", err)
		}
		defer gzipReader.Close()
		decodedReader = gzipReader
	default:
		return nil, 0, 0, fmt.Errorf("unsupported content encoding %q", encoding)
	}
	body, err := io.ReadAll(io.LimitReader(decodedReader, f.limits.MaximumDecodedBytes+1))
	if err != nil {
		return nil, compressed.count, int64(len(body)), fmt.Errorf("read response: %w", err)
	}
	decoded := int64(len(body))
	if compressed.count > f.limits.MaximumCompressedBytes {
		return nil, compressed.count, decoded, errors.New("compressed response exceeds byte limit")
	}
	if decoded > f.limits.MaximumDecodedBytes {
		return nil, compressed.count, decoded, errors.New("decoded response exceeds byte limit")
	}
	if encoding == "gzip" && compressed.count > 0 && float64(decoded)/float64(compressed.count) > f.limits.MaximumCompressionRatio {
		return nil, compressed.count, decoded, errors.New("response compression ratio exceeds limit")
	}
	return body, compressed.count, decoded, nil
}

type countingReader struct {
	reader io.Reader
	count  int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.count += int64(n)
	return n, err
}

func headerCount(header http.Header) int {
	count := 0
	for _, values := range header {
		count += len(values)
	}
	return count
}

func isRedirect(status int) bool {
	return status == http.StatusMovedPermanently || status == http.StatusFound || status == http.StatusSeeOther || status == http.StatusTemporaryRedirect || status == http.StatusPermanentRedirect
}

func resolveRedirect(base *url.URL, location string) (string, error) {
	if len(location) > 8192 {
		return "", errors.New("redirect location is too long")
	}
	reference, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parse redirect location: %w", err)
	}
	if reference.Scheme != "" && reference.Scheme != "http" && reference.Scheme != "https" {
		return "", fmt.Errorf("redirect scheme is not allowed")
	}
	resolved := base.ResolveReference(reference)
	if resolved.Port() != "" {
		if _, err := strconv.ParseUint(resolved.Port(), 10, 16); err != nil {
			return "", fmt.Errorf("redirect port is invalid")
		}
	}
	return resolved.String(), nil
}
