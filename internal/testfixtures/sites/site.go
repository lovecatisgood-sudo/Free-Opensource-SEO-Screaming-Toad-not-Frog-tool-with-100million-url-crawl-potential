package sites

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/conformance"
)

//go:embed core-indexability/manifest.json core-rules/manifest.json
var fixtureFiles embed.FS

func CoreIndexability() (conformance.Manifest, error) {
	return load("core-indexability/manifest.json")
}

func CoreRules() (conformance.Manifest, error) {
	return load("core-rules/manifest.json")
}

func load(name string) (conformance.Manifest, error) {
	body, err := fixtureFiles.Open(name)
	if err != nil {
		return conformance.Manifest{}, err
	}
	defer body.Close()
	return conformance.ParseManifest(body)
}

func Files() fs.FS { return fixtureFiles }

func Handler(manifest conformance.Manifest) (http.Handler, error) {
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	cases := make(map[string]conformance.FixtureCase, len(manifest.Cases))
	for _, fixture := range manifest.Cases {
		if _, exists := cases[fixture.Path]; exists {
			return nil, fmt.Errorf("duplicate fixture path %q", fixture.Path)
		}
		cases[fixture.Path] = fixture
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fixture, exists := cases[request.URL.Path]
		if !exists {
			http.NotFound(writer, request)
			return
		}
		baseURL := "http://" + request.Host
		for key, value := range fixture.Headers {
			writer.Header().Set(key, strings.ReplaceAll(value, "{{BASE_URL}}", baseURL))
		}
		writer.WriteHeader(fixture.StatusCode)
		_, _ = writer.Write([]byte(strings.ReplaceAll(fixture.Body, "{{BASE_URL}}", baseURL)))
	}), nil
}
