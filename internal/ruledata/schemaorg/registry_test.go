package schemaorg

import "testing"

func TestBundledRegistryHasFrozenProvenanceAndKnownTerms(t *testing.T) {
	registry, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if registry.Metadata.Version != "30.0" || registry.Metadata.SourceSHA256 != "4467fa19edcb1d7fb3c46c0adf3591b7f870c4a60b7838bdb61694fd02864cf6" {
		t.Fatalf("metadata=%+v", registry.Metadata)
	}
	if name, _, ok := registry.Type("https://schema.org/Article"); !ok || name != "Article" {
		t.Fatalf("Article lookup name=%q ok=%v", name, ok)
	}
	if name, definition, ok := registry.Property("schema:episodes"); !ok || name != "episodes" || definition.SupersededBy != "episode" {
		t.Fatalf("episodes lookup name=%q definition=%+v ok=%v", name, definition, ok)
	}
	if _, _, ok := registry.Type("DefinitelyNotARealSchemaType"); ok {
		t.Fatal("unknown type was accepted")
	}
	if !registry.IsA("NewsArticle", "Thing") || !registry.IsA("NewsArticle", "Article") || registry.IsA("Person", "Product") {
		t.Fatal("Schema.org type hierarchy resolution is incorrect")
	}
}
