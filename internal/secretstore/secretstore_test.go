package secretstore

import (
	"context"
	"testing"
)

func TestMemoryCopiesValuesAndNeverListsSecrets(t *testing.T) {
	store := NewMemory()
	value := []byte("token-value")
	if err := store.Put(context.Background(), "secret_google", value); err != nil {
		t.Fatal(err)
	}
	value[0] = 'X'
	loaded, err := store.Get(context.Background(), "secret_google")
	if err != nil {
		t.Fatal(err)
	}
	if string(loaded) != "token-value" {
		t.Fatalf("loaded=%q", loaded)
	}
	loaded[0] = 'Y'
	again, _ := store.Get(context.Background(), "secret_google")
	if string(again) != "token-value" {
		t.Fatal("store returned mutable backing memory")
	}
}
func TestReferencesAndBoundsAreStrict(t *testing.T) {
	store := NewMemory()
	for _, ref := range []string{"google", "secret_", "secret_bad space", "../secret_x"} {
		if err := store.Put(context.Background(), ref, []byte("x")); err == nil {
			t.Errorf("accepted %q", ref)
		}
	}
	if err := store.Put(context.Background(), "secret_ok", make([]byte, 64<<10+1)); err == nil {
		t.Fatal("oversized secret accepted")
	}
}
