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

func TestResolveRendererUsesTrustedStartupPaths(t *testing.T) {
	t.Setenv("SEO_AUDITOR_RENDERER_SCRIPT", filepath.Join(t.TempDir(), "worker.js"))
	t.Setenv("PLAYWRIGHT_BROWSERS_PATH", filepath.Join(t.TempDir(), "browsers"))
	t.Setenv("SEO_AUDITOR_NODE_BINARY", "node-pinned")
	t.Setenv("SEO_AUDITOR_CONTAINER_SANDBOX", "1")
	configuration, err := ResolveRenderer()
	if err != nil {
		t.Fatal(err)
	}
	if !configuration.Enabled || !filepath.IsAbs(configuration.ScriptPath) || !filepath.IsAbs(configuration.BrowserPath) || configuration.NodeBinary != "node-pinned" || !configuration.ContainerSandbox {
		t.Fatalf("configuration=%+v", configuration)
	}
}

func TestResolveRendererDisabledAndRejectsAmbiguousSandboxFlag(t *testing.T) {
	t.Setenv("SEO_AUDITOR_RENDERER_SCRIPT", "")
	configuration, err := ResolveRenderer()
	if err != nil || configuration.Enabled {
		t.Fatalf("configuration=%+v err=%v", configuration, err)
	}
	t.Setenv("SEO_AUDITOR_RENDERER_SCRIPT", filepath.Join(t.TempDir(), "worker.js"))
	t.Setenv("SEO_AUDITOR_CONTAINER_SANDBOX", "yes")
	if _, err := ResolveRenderer(); err == nil {
		t.Fatal("expected ambiguous sandbox flag to be rejected")
	}
}
