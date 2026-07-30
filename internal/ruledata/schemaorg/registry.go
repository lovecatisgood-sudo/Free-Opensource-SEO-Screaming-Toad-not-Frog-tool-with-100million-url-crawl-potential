package schemaorg

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Metadata struct {
	Vocabulary    string `json:"vocabulary"`
	Version       string `json:"version"`
	Published     string `json:"published"`
	SourceURL     string `json:"source_url"`
	SourceSHA256  string `json:"source_sha256"`
	GeneratedBy   string `json:"generated_by"`
	TypeCount     int    `json:"type_count"`
	PropertyCount int    `json:"property_count"`
}

type TypeDefinition struct {
	Parents      []string `json:"parents,omitempty"`
	SupersededBy string   `json:"superseded_by,omitempty"`
	Pending      bool     `json:"pending,omitempty"`
}

type PropertyDefinition struct {
	Domains      []string `json:"domains,omitempty"`
	Ranges       []string `json:"ranges,omitempty"`
	SupersededBy string   `json:"superseded_by,omitempty"`
	Pending      bool     `json:"pending,omitempty"`
}

type Bundle struct {
	SchemaVersion int                           `json:"schema_version"`
	Metadata      Metadata                      `json:"metadata"`
	Types         map[string]TypeDefinition     `json:"types"`
	Properties    map[string]PropertyDefinition `json:"properties"`
}

type Registry struct{ Bundle }

//go:embed data/v30.0.json
var data embed.FS

var defaultRegistry struct {
	sync.Once
	value *Registry
	err   error
}

func Default() (*Registry, error) {
	defaultRegistry.Do(func() {
		body, err := data.ReadFile("data/v30.0.json")
		if err != nil {
			defaultRegistry.err = err
			return
		}
		var bundle Bundle
		if err := json.Unmarshal(body, &bundle); err != nil {
			defaultRegistry.err = fmt.Errorf("decode bundled Schema.org registry: %w", err)
			return
		}
		if err := bundle.Validate(); err != nil {
			defaultRegistry.err = err
			return
		}
		defaultRegistry.value = &Registry{Bundle: bundle}
	})
	return defaultRegistry.value, defaultRegistry.err
}

func (b Bundle) Validate() error {
	if b.SchemaVersion != 1 || b.Metadata.Vocabulary != "Schema.org" || b.Metadata.Version == "" {
		return errors.New("Schema.org registry metadata is invalid")
	}
	if len(b.Metadata.SourceSHA256) != 64 || !strings.HasPrefix(b.Metadata.SourceURL, "https://") {
		return errors.New("Schema.org registry provenance is invalid")
	}
	if b.Metadata.TypeCount != len(b.Types) || b.Metadata.PropertyCount != len(b.Properties) || len(b.Types) < 500 || len(b.Properties) < 500 {
		return errors.New("Schema.org registry counts are invalid")
	}
	return nil
}

func NormalizeTerm(value string) (string, bool) {
	value = strings.TrimSpace(value)
	for _, prefix := range []string{"schema:", "https://schema.org/", "http://schema.org/"} {
		if strings.HasPrefix(value, prefix) {
			name := strings.TrimPrefix(value, prefix)
			return name, name != "" && !strings.ContainsAny(name, "/#")
		}
	}
	if value != "" && !strings.ContainsAny(value, ":/#") {
		return value, true
	}
	return "", false
}

func (r *Registry) Type(value string) (string, TypeDefinition, bool) {
	name, ok := NormalizeTerm(value)
	if !ok {
		return "", TypeDefinition{}, false
	}
	definition, exists := r.Types[name]
	return name, definition, exists
}

func (r *Registry) Property(value string) (string, PropertyDefinition, bool) {
	name, ok := NormalizeTerm(value)
	if !ok {
		return "", PropertyDefinition{}, false
	}
	definition, exists := r.Properties[name]
	return name, definition, exists
}

func (r *Registry) IsA(value, ancestor string) bool {
	valueName, valueOK := NormalizeTerm(value)
	ancestorName, ancestorOK := NormalizeTerm(ancestor)
	if !valueOK || !ancestorOK {
		return false
	}
	seen := map[string]bool{}
	var visit func(string, int) bool
	visit = func(current string, depth int) bool {
		if current == ancestorName {
			return true
		}
		if depth > 64 || seen[current] {
			return false
		}
		seen[current] = true
		definition, ok := r.Types[current]
		if !ok {
			return false
		}
		for _, parent := range definition.Parents {
			if visit(parent, depth+1) {
				return true
			}
		}
		return false
	}
	return visit(valueName, 0)
}
