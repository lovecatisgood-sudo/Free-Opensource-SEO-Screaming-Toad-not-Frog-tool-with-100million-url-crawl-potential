// seo-auditor-sbom emits a deterministic CycloneDX inventory from the pinned
// Go module graph and production pnpm workspace graph.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type component struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	PURL    string `json:"purl,omitempty"`
}

type document struct {
	BOMFormat   string      `json:"bomFormat"`
	SpecVersion string      `json:"specVersion"`
	Version     int         `json:"version"`
	Metadata    metadata    `json:"metadata"`
	Components  []component `json:"components"`
}

type metadata struct {
	Component component `json:"component"`
}

func main() {
	components := make(map[string]component)
	if err := goComponents(components); err != nil {
		fail(err)
	}
	if err := nodeComponents(components); err != nil {
		fail(err)
	}
	items := make([]component, 0, len(components))
	for _, item := range components {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].PURL < items[j].PURL })
	result := document{BOMFormat: "CycloneDX", SpecVersion: "1.5", Version: 1, Metadata: metadata{Component: component{Type: "application", Name: "seo-auditor", Version: buildVersion()}}, Components: items}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fail(err)
	}
}

func goComponents(destination map[string]component) error {
	output, err := exec.Command("go", "list", "-m", "-json", "all").Output()
	if err != nil {
		return fmt.Errorf("list Go modules: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var item struct{ Path, Version string }
		if err := decoder.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if item.Path == "" || item.Version == "" {
			continue
		}
		purl := "pkg:golang/" + item.Path + "@" + item.Version
		destination[purl] = component{Type: "library", Name: item.Path, Version: item.Version, PURL: purl}
	}
	return nil
}

func nodeComponents(destination map[string]component) error {
	output, err := exec.Command("pnpm", "list", "-r", "--prod", "--json", "--depth", "Infinity").Output()
	if err != nil {
		return fmt.Errorf("list pnpm packages: %w", err)
	}
	var roots []map[string]any
	if err := json.Unmarshal(output, &roots); err != nil {
		return err
	}
	for _, root := range roots {
		walkNode(root, destination)
	}
	return nil
}

func walkNode(value map[string]any, destination map[string]component) {
	for _, field := range []string{"dependencies", "optionalDependencies"} {
		entries, _ := value[field].(map[string]any)
		for name, raw := range entries {
			item, _ := raw.(map[string]any)
			version, _ := item["version"].(string)
			version = strings.TrimPrefix(version, "npm:")
			if version != "" && !strings.HasPrefix(version, "link:") {
				purl := "pkg:npm/" + name + "@" + version
				destination[purl] = component{Type: "library", Name: name, Version: version, PURL: purl}
			}
			walkNode(item, destination)
		}
	}
}

func buildVersion() string {
	if value := strings.TrimSpace(os.Getenv("SEO_AUDITOR_VERSION")); value != "" {
		return value
	}
	return "development"
}

func fail(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
