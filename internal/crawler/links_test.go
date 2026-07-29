package crawler

import (
	"reflect"
	"testing"
)

func TestDiscoverLinksResolvesBaseAndDeduplicates(t *testing.T) {
	t.Parallel()

	body := []byte(`<html><head><base href="https://example.com/docs/"></head><body>
        <a href="start#one">one</a><a href="start#two">two</a>
        <a href="mailto:test@example.com">mail</a><iframe src="/frame"></iframe></body></html>`)
	got, err := DiscoverLinks("https://example.com/original", body, 100)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://example.com/docs/start", "https://example.com/frame"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("links = %#v, want %#v", got, want)
	}
}

func TestDiscoverLinksIsBounded(t *testing.T) {
	t.Parallel()

	body := []byte(`<a href="/1"></a><a href="/2"></a><a href="/3"></a>`)
	got, err := DiscoverLinks("https://example.com/", body, 2)
	if err != nil || len(got) != 2 {
		t.Fatalf("links = %#v, err = %v", got, err)
	}
}
