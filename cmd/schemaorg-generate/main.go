package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/ruledata/schemaorg"
)

func main() {
	input := flag.String("input", "", "official Schema.org release JSON-LD")
	output := flag.String("output", "", "generated compact registry JSON")
	version := flag.String("version", "", "Schema.org release version")
	source := flag.String("source", "", "immutable upstream source URL")
	published := flag.String("published", "", "upstream publication date (YYYY-MM-DD)")
	flag.Parse()
	if *input == "" || *output == "" || *version == "" || *source == "" || *published == "" {
		fatal(errors.New("input, output, version, source and published are required"))
	}
	body, err := os.ReadFile(*input)
	if err != nil {
		fatal(err)
	}
	var document struct {
		Graph []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal(body, &document); err != nil {
		fatal(fmt.Errorf("decode source JSON-LD: %w", err))
	}
	digest := sha256.Sum256(body)
	bundle := schemaorg.Bundle{
		SchemaVersion: 1,
		Metadata: schemaorg.Metadata{
			Vocabulary: "Schema.org", Version: *version, Published: *published, SourceURL: *source,
			SourceSHA256: hex.EncodeToString(digest[:]), GeneratedBy: "cmd/schemaorg-generate.v1",
		},
		Types: make(map[string]schemaorg.TypeDefinition), Properties: make(map[string]schemaorg.PropertyDefinition),
	}
	for _, node := range document.Graph {
		id, _ := node["@id"].(string)
		name, ok := schemaName(id)
		if !ok {
			continue
		}
		kinds := stringValues(node["@type"])
		pending := containsID(node["schema:isPartOf"], "https://pending.schema.org")
		superseded := firstSchemaName(node["schema:supersededBy"])
		if contains(kinds, "rdfs:Class") {
			bundle.Types[name] = schemaorg.TypeDefinition{Parents: schemaNames(node["rdfs:subClassOf"]), SupersededBy: superseded, Pending: pending}
		}
		if contains(kinds, "rdf:Property") {
			bundle.Properties[name] = schemaorg.PropertyDefinition{Domains: schemaNames(node["schema:domainIncludes"]), Ranges: schemaNames(node["schema:rangeIncludes"]), SupersededBy: superseded, Pending: pending}
		}
	}
	bundle.Metadata.TypeCount = len(bundle.Types)
	bundle.Metadata.PropertyCount = len(bundle.Properties)
	if bundle.Metadata.TypeCount < 500 || bundle.Metadata.PropertyCount < 500 {
		fatal(fmt.Errorf("generated registry is unexpectedly small: %d types, %d properties", bundle.Metadata.TypeCount, bundle.Metadata.PropertyCount))
	}
	encoded, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
}

func stringValues(value any) []string {
	switch item := value.(type) {
	case string:
		return []string{item}
	case []any:
		result := make([]string, 0, len(item))
		for _, child := range item {
			result = append(result, stringValues(child)...)
		}
		return result
	case map[string]any:
		if id, ok := item["@id"].(string); ok {
			return []string{id}
		}
	}
	return nil
}

func schemaNames(value any) []string {
	seen := make(map[string]struct{})
	for _, id := range stringValues(value) {
		if name, ok := schemaName(id); ok {
			seen[name] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func firstSchemaName(value any) string {
	values := schemaNames(value)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func schemaName(value string) (string, bool) {
	for _, prefix := range []string{"schema:", "https://schema.org/", "http://schema.org/"} {
		if strings.HasPrefix(value, prefix) {
			name := strings.TrimSpace(strings.TrimPrefix(value, prefix))
			return name, name != "" && !strings.ContainsAny(name, "/#")
		}
	}
	return "", false
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func containsID(value any, wanted string) bool {
	for _, id := range stringValues(value) {
		if id == wanted {
			return true
		}
	}
	return false
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "schemaorg-generate:", err)
	os.Exit(1)
}
