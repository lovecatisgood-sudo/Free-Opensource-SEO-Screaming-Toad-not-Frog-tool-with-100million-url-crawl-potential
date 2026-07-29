package fetchpolicy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"sort"
	"strings"

	"golang.org/x/net/idna"
)

type NormalizedURL struct {
	URL        *url.URL
	RequestKey string
}

// NormalizeURL creates a deterministic fetch identity. It never resolves DNS;
// callers must separately validate all resolved addresses immediately before use.
func NormalizeURL(raw string) (NormalizedURL, error) {
	if len(raw) == 0 || len(raw) > 8192 {
		return NormalizedURL{}, errors.New("URL length is invalid")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return NormalizedURL{}, fmt.Errorf("parse URL: %w", err)
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme != "http" && u.Scheme != "https" {
		return NormalizedURL{}, errors.New("only HTTP and HTTPS URLs are allowed")
	}
	if u.User != nil {
		return NormalizedURL{}, errors.New("URL userinfo is not allowed")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "" {
		return NormalizedURL{}, errors.New("URL host is required")
	}
	if net.ParseIP(host) == nil {
		host, err = idna.Lookup.ToASCII(host)
		if err != nil {
			return NormalizedURL{}, fmt.Errorf("normalize international host: %w", err)
		}
	}
	port := u.Port()
	if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		u.Host = "[" + host + "]"
	} else {
		u.Host = host
	}
	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	} else {
		cleaned := path.Clean(u.EscapedPath())
		if strings.HasSuffix(u.Path, "/") && cleaned != "/" {
			cleaned += "/"
		}
		decoded, decodeErr := url.PathUnescape(cleaned)
		if decodeErr != nil {
			return NormalizedURL{}, fmt.Errorf("normalize path: %w", decodeErr)
		}
		u.Path = decoded
		u.RawPath = ""
	}
	query := u.Query()
	for key, values := range query {
		sort.Strings(values)
		query[key] = values
	}
	u.RawQuery = query.Encode()
	return NormalizedURL{URL: u, RequestKey: u.String()}, nil
}
