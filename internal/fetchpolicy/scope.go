package fetchpolicy

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

const (
	maximumScopePatterns = 100
	maximumPatternLength = 1024
)

type ScopeConfig struct {
	AllowedHosts      []string
	AllowSubdomains   bool
	IncludePathRegex  []string
	ExcludePathRegex  []string
	IncludeQueryRegex []string
	ExcludeQueryRegex []string
}

type Scope struct {
	hosts           []string
	allowSubdomains bool
	includePath     []*regexp.Regexp
	excludePath     []*regexp.Regexp
	includeQuery    []*regexp.Regexp
	excludeQuery    []*regexp.Regexp
}

func CompileScope(config ScopeConfig) (*Scope, error) {
	if len(config.AllowedHosts) == 0 || len(config.AllowedHosts) > maximumScopePatterns {
		return nil, fmt.Errorf("allowed hosts must contain between 1 and %d entries", maximumScopePatterns)
	}
	hosts := make([]string, 0, len(config.AllowedHosts))
	seen := make(map[string]struct{}, len(config.AllowedHosts))
	for _, raw := range config.AllowedHosts {
		normalized, err := normalizeScopeHost(raw)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized]; !exists {
			hosts = append(hosts, normalized)
			seen[normalized] = struct{}{}
		}
	}
	includePath, err := compilePatterns("include path", config.IncludePathRegex)
	if err != nil {
		return nil, err
	}
	excludePath, err := compilePatterns("exclude path", config.ExcludePathRegex)
	if err != nil {
		return nil, err
	}
	includeQuery, err := compilePatterns("include query", config.IncludeQueryRegex)
	if err != nil {
		return nil, err
	}
	excludeQuery, err := compilePatterns("exclude query", config.ExcludeQueryRegex)
	if err != nil {
		return nil, err
	}
	return &Scope{
		hosts: hosts, allowSubdomains: config.AllowSubdomains,
		includePath: includePath, excludePath: excludePath,
		includeQuery: includeQuery, excludeQuery: excludeQuery,
	}, nil
}

func (s *Scope) Evaluate(target NormalizedURL) error {
	host := strings.ToLower(target.URL.Hostname())
	hostAllowed := false
	for _, allowed := range s.hosts {
		if host == allowed || (s.allowSubdomains && strings.HasSuffix(host, "."+allowed)) {
			hostAllowed = true
			break
		}
	}
	if !hostAllowed {
		return fmt.Errorf("host is outside crawl scope")
	}
	if !matchesIncludes(s.includePath, target.URL.EscapedPath()) {
		return fmt.Errorf("path does not match include scope")
	}
	if matchesAny(s.excludePath, target.URL.EscapedPath()) {
		return fmt.Errorf("path matches exclude scope")
	}
	if !matchesIncludes(s.includeQuery, target.URL.RawQuery) {
		return fmt.Errorf("query does not match include scope")
	}
	if matchesAny(s.excludeQuery, target.URL.RawQuery) {
		return fmt.Errorf("query matches exclude scope")
	}
	return nil
}

func normalizeScopeHost(raw string) (string, error) {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, "."))
	if raw == "" || strings.ContainsAny(raw, "/?#@") {
		return "", fmt.Errorf("invalid allowed host %q", raw)
	}
	if net.ParseIP(strings.Trim(raw, "[]")) == nil && strings.Contains(raw, ":") {
		return "", fmt.Errorf("allowed host must not contain a port")
	}
	// Reuse URL IDNA normalization without accepting ports or paths.
	normalized, err := NormalizeURL("https://" + raw + "/")
	if err != nil {
		return "", fmt.Errorf("invalid allowed host %q: %w", raw, err)
	}
	return strings.ToLower(normalized.URL.Hostname()), nil
}

func compilePatterns(kind string, values []string) ([]*regexp.Regexp, error) {
	if len(values) > maximumScopePatterns {
		return nil, fmt.Errorf("%s patterns exceed maximum of %d", kind, maximumScopePatterns)
	}
	result := make([]*regexp.Regexp, 0, len(values))
	for _, value := range values {
		if value == "" || len(value) > maximumPatternLength {
			return nil, fmt.Errorf("%s pattern length is invalid", kind)
		}
		compiled, err := regexp.Compile(value)
		if err != nil {
			return nil, fmt.Errorf("compile %s pattern: %w", kind, err)
		}
		result = append(result, compiled)
	}
	return result, nil
}

func matchesIncludes(patterns []*regexp.Regexp, value string) bool {
	return len(patterns) == 0 || matchesAny(patterns, value)
}

func matchesAny(patterns []*regexp.Regexp, value string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}
