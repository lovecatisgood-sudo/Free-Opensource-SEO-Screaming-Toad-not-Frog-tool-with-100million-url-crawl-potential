package localclient

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
	"strings"
	"sync"
	"time"
)

type Client struct {
	base   *url.URL
	http   *http.Client
	mu     sync.Mutex
	cookie *http.Cookie
	csrf   string
}

func New(rawBase string) (*Client, error) {
	parsed, err := url.Parse(strings.TrimSuffix(rawBase, "/"))
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" || parsed.Hostname() == "" || parsed.Path != "" {
		return nil, errors.New("local API URL must be an HTTP origin")
	}
	address := net.ParseIP(parsed.Hostname())
	if address == nil || !address.IsLoopback() {
		return nil, errors.New("local API must use a numeric loopback address")
	}
	return &Client{base: parsed, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) Bootstrap(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cookie != nil {
		return nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base.String()+"/api/v1/session", nil)
	if err != nil {
		return err
	}
	request.Header.Set("Origin", c.base.String())
	response, err := c.http.Do(request)
	if err != nil {
		return fmt.Errorf("connect to local API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return decodeError(response)
	}
	var body struct {
		CSRF string `json:"csrf_token"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&body); err != nil {
		return err
	}
	cookies := response.Cookies()
	if len(cookies) == 0 || body.CSRF == "" {
		return errors.New("local API session response is incomplete")
	}
	c.cookie = cookies[0]
	c.csrf = body.CSRF
	return nil
}

func (c *Client) Call(ctx context.Context, method, path string, input, output any) error {
	if !strings.HasPrefix(path, "/api/v1/") {
		return errors.New("client path is outside API v1")
	}
	if err := c.Bootstrap(ctx); err != nil {
		return err
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		if len(encoded) > 1<<20 {
			return errors.New("request exceeds local client limit")
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.base.String()+path, body)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Origin", c.base.String())
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.AddCookie(c.cookie)
	if method != http.MethodGet && method != http.MethodHead {
		request.Header.Set("X-CSRF-Token", c.csrf)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return fmt.Errorf("local API request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return decodeError(response)
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(io.LimitReader(response.Body, 16<<20)).Decode(output)
}

func decodeError(response *http.Response) error {
	var body struct {
		Error struct{ Code, Message string } `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&body)
	if body.Error.Message == "" {
		body.Error.Message = response.Status
	}
	return fmt.Errorf("%s: %s", body.Error.Code, body.Error.Message)
}
