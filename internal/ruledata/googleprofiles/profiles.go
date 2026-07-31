package googleprofiles

import (
	"embed"
	"encoding/json"
	"errors"
	"sync"
)

type Metadata struct {
	Provider    string   `json:"provider"`
	Version     string   `json:"version"`
	ReviewedAt  string   `json:"reviewed_at"`
	Sources     []string `json:"sources"`
	Limitations string   `json:"limitations"`
}

type NestedRequirement struct {
	Types        []string `json:"types"`
	MinimumCount int      `json:"minimum_count"`
	Required     []string `json:"required"`
}

type Profile struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	AppliesTo   []string            `json:"applies_to"`
	Required    []string            `json:"required"`
	Recommended []string            `json:"recommended"`
	Nested      []NestedRequirement `json:"nested"`
}

type Bundle struct {
	SchemaVersion int       `json:"schema_version"`
	Metadata      Metadata  `json:"metadata"`
	Profiles      []Profile `json:"profiles"`
}

//go:embed data/v2026-07-30.json
var data embed.FS

var cached struct {
	sync.Once
	value Bundle
	err   error
}

func Default() (Bundle, error) {
	cached.Do(func() {
		body, err := data.ReadFile("data/v2026-07-30.json")
		if err != nil {
			cached.err = err
			return
		}
		if err := json.Unmarshal(body, &cached.value); err != nil {
			cached.err = err
			return
		}
		if cached.value.SchemaVersion != 1 || cached.value.Metadata.Version == "" || len(cached.value.Profiles) != 3 {
			cached.err = errors.New("Google search-feature profile bundle is invalid")
		}
	})
	return cached.value, cached.err
}
