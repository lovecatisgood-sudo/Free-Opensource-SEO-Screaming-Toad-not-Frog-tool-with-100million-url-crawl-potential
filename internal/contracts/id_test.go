package contracts

import (
	"strings"
	"testing"
)

func TestNewID(t *testing.T) {
	t.Parallel()

	id, err := NewID("crawl")
	if err != nil {
		t.Fatalf("NewID: %v", err)
	}
	if !strings.HasPrefix(string(id), "crawl_") || len(id) != len("crawl_")+32 {
		t.Fatalf("unexpected ID format: %q", id)
	}
}

func TestNewIDRequiresPrefix(t *testing.T) {
	t.Parallel()

	if _, err := NewID(""); err == nil {
		t.Fatal("expected an error")
	}
}
