package crawler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

const maximumRobotsBytes = 512 << 10

type RobotsRule struct {
	Allow       bool
	Pattern     string
	expression  *regexp.Regexp
	specificity int
}

type RobotsGroup struct {
	Agents     []string
	Rules      []RobotsRule
	CrawlDelay time.Duration
}

type RobotsPolicy struct {
	Groups   []RobotsGroup
	Sitemaps []string
}

type RobotsDecision struct {
	Allowed     bool
	MatchedRule string
	CrawlDelay  time.Duration
	StatusCode  int
}

func ParseRobots(body []byte) (RobotsPolicy, error) {
	if len(body) > maximumRobotsBytes {
		return RobotsPolicy{}, errors.New("robots.txt exceeds byte limit")
	}
	policy := RobotsPolicy{}
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	scanner.Buffer(make([]byte, 4096), maximumRobotsBytes)
	var group *RobotsGroup
	directivesSeen := false
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "user-agent":
			if value == "" {
				continue
			}
			if group == nil || directivesSeen {
				policy.Groups = append(policy.Groups, RobotsGroup{})
				group = &policy.Groups[len(policy.Groups)-1]
				directivesSeen = false
			}
			group.Agents = append(group.Agents, strings.ToLower(value))
		case "allow", "disallow":
			if group == nil || value == "" {
				continue
			}
			rule, err := compileRobotsRule(key == "allow", value)
			if err != nil {
				return RobotsPolicy{}, err
			}
			group.Rules = append(group.Rules, rule)
			directivesSeen = true
		case "crawl-delay":
			if group == nil {
				continue
			}
			seconds, err := strconv.ParseFloat(value, 64)
			if err == nil && seconds >= 0 && seconds <= 3600 {
				group.CrawlDelay = time.Duration(seconds * float64(time.Second))
			}
			directivesSeen = true
		case "sitemap":
			if parsed, err := url.Parse(value); err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
				policy.Sitemaps = append(policy.Sitemaps, parsed.String())
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return RobotsPolicy{}, fmt.Errorf("scan robots.txt: %w", err)
	}
	return policy, nil
}

func (p RobotsPolicy) Evaluate(userAgent, requestPath string) RobotsDecision {
	userAgent = strings.ToLower(userAgent)
	bestAgent := -1
	groups := make([]RobotsGroup, 0)
	for _, group := range p.Groups {
		groupSpecificity := -1
		for _, agent := range group.Agents {
			specificity := -1
			if agent == "*" {
				specificity = 0
			} else if strings.Contains(userAgent, agent) {
				specificity = len(agent)
			}
			groupSpecificity = max(groupSpecificity, specificity)
		}
		if groupSpecificity > bestAgent {
			bestAgent = groupSpecificity
			groups = groups[:0]
		}
		if groupSpecificity == bestAgent && groupSpecificity >= 0 {
			groups = append(groups, group)
		}
	}
	decision := RobotsDecision{Allowed: true}
	bestRule := -1
	for _, group := range groups {
		decision.CrawlDelay = max(decision.CrawlDelay, group.CrawlDelay)
		for _, rule := range group.Rules {
			if rule.expression.MatchString(requestPath) && (rule.specificity > bestRule || (rule.specificity == bestRule && rule.Allow)) {
				bestRule = rule.specificity
				decision.Allowed = rule.Allow
				decision.MatchedRule = rule.Pattern
			}
		}
	}
	return decision
}

func compileRobotsRule(allow bool, pattern string) (RobotsRule, error) {
	if len(pattern) > 4096 {
		return RobotsRule{}, errors.New("robots rule exceeds length limit")
	}
	anchored := strings.HasSuffix(pattern, "$")
	if anchored {
		pattern = strings.TrimSuffix(pattern, "$")
	}
	parts := strings.Split(pattern, "*")
	var expression strings.Builder
	expression.WriteString("^")
	for index, part := range parts {
		if index > 0 {
			expression.WriteString(".*")
		}
		expression.WriteString(regexp.QuoteMeta(part))
	}
	if anchored {
		expression.WriteString("$")
	}
	compiled, err := regexp.Compile(expression.String())
	if err != nil {
		return RobotsRule{}, err
	}
	return RobotsRule{Allow: allow, Pattern: pattern, expression: compiled, specificity: len(strings.ReplaceAll(pattern, "*", ""))}, nil
}

type RobotsChecker interface {
	Check(context.Context, string) (RobotsDecision, error)
}

type RobotsService struct {
	fetcher Fetcher
	agent   string
	ttl     time.Duration
	mu      sync.Mutex
	cache   map[string]cachedRobots
}

type cachedRobots struct {
	policy  RobotsPolicy
	status  int
	expires time.Time
}

func NewRobotsService(fetcher Fetcher, userAgent string, ttl time.Duration) (*RobotsService, error) {
	if fetcher == nil || userAgent == "" || ttl <= 0 {
		return nil, errors.New("robots service configuration is invalid")
	}
	return &RobotsService{fetcher: fetcher, agent: userAgent, ttl: ttl, cache: make(map[string]cachedRobots)}, nil
}

func (s *RobotsService) Check(ctx context.Context, raw string) (RobotsDecision, error) {
	target, err := url.Parse(raw)
	if err != nil {
		return RobotsDecision{}, err
	}
	key := target.Scheme + "://" + target.Host
	s.mu.Lock()
	cached, exists := s.cache[key]
	s.mu.Unlock()
	if !exists || time.Now().After(cached.expires) {
		result, err := s.fetcher.Fetch(ctx, key+"/robots.txt")
		if err != nil {
			return RobotsDecision{}, fmt.Errorf("fetch robots.txt: %w", err)
		}
		cached = cachedRobots{status: result.StatusCode, expires: time.Now().Add(s.ttl)}
		switch {
		case result.StatusCode >= 200 && result.StatusCode < 300:
			cached.policy, err = ParseRobots(result.Body)
			if err != nil {
				return RobotsDecision{}, err
			}
		case result.StatusCode == http.StatusUnauthorized || result.StatusCode == http.StatusForbidden:
			cached.policy, _ = ParseRobots([]byte("User-agent: *\nDisallow: /"))
		case result.StatusCode == http.StatusNotFound || result.StatusCode == http.StatusGone:
			// Missing robots files permit crawling.
		default:
			return RobotsDecision{}, fmt.Errorf("robots.txt returned retryable status %d", result.StatusCode)
		}
		s.mu.Lock()
		s.cache[key] = cached
		s.mu.Unlock()
	}
	requestPath := target.EscapedPath()
	if target.RawQuery != "" {
		requestPath += "?" + target.RawQuery
	}
	decision := cached.policy.Evaluate(s.agent, requestPath)
	decision.StatusCode = cached.status
	return decision, nil
}

type RobotsDeniedError struct{ Decision RobotsDecision }

func (e *RobotsDeniedError) Error() string { return "robots.txt disallows this URL" }

type RobotsEnforcingFetcher struct {
	Base   Fetcher
	Robots RobotsChecker
}

func (f RobotsEnforcingFetcher) Fetch(ctx context.Context, raw string) (fetchpolicy.FetchResult, error) {
	decision, err := f.Robots.Check(ctx, raw)
	if err != nil {
		return fetchpolicy.FetchResult{}, err
	}
	if !decision.Allowed {
		return fetchpolicy.FetchResult{}, &RobotsDeniedError{Decision: decision}
	}
	return f.Base.Fetch(ctx, raw)
}
