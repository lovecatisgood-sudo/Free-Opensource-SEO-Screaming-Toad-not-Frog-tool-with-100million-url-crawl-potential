package segmentstore

import (
	"context"
	"strings"
	"testing"
)

func TestImmutableManagedSegmentCommitAndVerify(t *testing.T) {
	t.Parallel()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.Commit(context.Background(), "crawl_test", 0, 1024, strings.NewReader("one\ntwo\n"))
	if err != nil || !first.Inserted {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	replay, err := store.Commit(context.Background(), "crawl_test", 0, 1024, strings.NewReader("one\ntwo\n"))
	if err != nil || replay.Inserted || replay.Checksum != first.Checksum {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	if _, err := store.Commit(context.Background(), "crawl_test", 0, 1024, strings.NewReader("different")); err == nil {
		t.Fatal("conflicting immutable commit succeeded")
	}
	if err := store.Verify(first.ObjectKey, first.Checksum); err != nil {
		t.Fatal(err)
	}
}

func TestSegmentStoreRejectsTraversalAndOversize(t *testing.T) {
	t.Parallel()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Commit(context.Background(), "../escape", 0, 10, strings.NewReader("x")); err == nil {
		t.Fatal("invalid crawl identity accepted")
	}
	if _, err := store.Commit(context.Background(), "crawl_test", 0, 2, strings.NewReader("three")); err == nil {
		t.Fatal("oversized segment accepted")
	}
	if err := store.Verify("../escape", strings.Repeat("0", 64)); err == nil {
		t.Fatal("traversal object key accepted")
	}
}
