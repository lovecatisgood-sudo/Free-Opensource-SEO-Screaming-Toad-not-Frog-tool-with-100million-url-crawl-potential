package crawler

import (
	"testing"
	"time"
)

func TestRobotsUsesMostSpecificAgentAndRule(t *testing.T) {
	t.Parallel()

	policy, err := ParseRobots([]byte(`
User-agent: *
Disallow: /private/
Allow: /private/public$
Crawl-delay: 1.5
Sitemap: https://example.com/sitemap.xml

User-agent: SEOAuditor
Disallow: /tmp*
Allow: /tmp/public
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(policy.Sitemaps) != 1 {
		t.Fatalf("sitemaps = %#v", policy.Sitemaps)
	}
	if decision := policy.Evaluate("SEOAuditor/1.0", "/tmp/file"); decision.Allowed {
		t.Fatalf("expected disallow: %+v", decision)
	}
	if decision := policy.Evaluate("SEOAuditor/1.0", "/tmp/public"); !decision.Allowed {
		t.Fatalf("expected longest allow: %+v", decision)
	}
	generic := policy.Evaluate("OtherBot", "/private/file")
	if generic.Allowed || generic.CrawlDelay != 1500*time.Millisecond {
		t.Fatalf("unexpected generic decision: %+v", generic)
	}
	if decision := policy.Evaluate("OtherBot", "/private/public"); !decision.Allowed {
		t.Fatalf("expected exact allow: %+v", decision)
	}
}

func TestRobotsEmptyDisallowAllows(t *testing.T) {
	t.Parallel()
	policy, err := ParseRobots([]byte("User-agent: *\nDisallow:\n"))
	if err != nil {
		t.Fatal(err)
	}
	if decision := policy.Evaluate("SEOAuditor", "/anything"); !decision.Allowed {
		t.Fatalf("unexpected disallow: %+v", decision)
	}
}
