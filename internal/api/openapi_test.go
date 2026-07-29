package api

import "testing"

func TestOpenAPIDocumentIncludesCoreBoundedWorkflows(t *testing.T) {
	t.Parallel()
	paths, ok := openAPIDocument["paths"].(map[string]any)
	if !ok {
		t.Fatal("paths missing")
	}
	for _, path := range []string{"/crawls", "/crawls/{crawlId}/events", "/crawls/{crawlId}/pages/{pageId}", "/comparisons", "/exports", "/artifacts/{artifactId}"} {
		if _, ok := paths[path]; !ok {
			t.Errorf("OpenAPI path missing: %s", path)
		}
	}
}
