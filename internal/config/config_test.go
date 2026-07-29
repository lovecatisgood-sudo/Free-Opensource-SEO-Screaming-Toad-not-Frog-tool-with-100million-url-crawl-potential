package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveAndEnsurePaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("SEO_AUDITOR_DATA_DIR", filepath.Join(root, "data"))
	t.Setenv("SEO_AUDITOR_CACHE_DIR", filepath.Join(root, "cache"))
	t.Setenv("SEO_AUDITOR_ARTIFACT_DIR", filepath.Join(root, "artifacts"))

	paths, err := ResolvePaths()
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	if err := paths.Ensure(); err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	for _, path := range []string{paths.Data, paths.Cache, paths.Artifacts} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != directoryMode {
			t.Fatalf("%s mode = %o, want %o", path, info.Mode().Perm(), directoryMode)
		}
	}
}

func TestResolveServerRejectsPublicBinding(t *testing.T) {
	t.Setenv("SEO_AUDITOR_BIND_HOST", "0.0.0.0")
	if _, err := ResolveServer(); err == nil {
		t.Fatal("expected public bind rejection")
	}
}
