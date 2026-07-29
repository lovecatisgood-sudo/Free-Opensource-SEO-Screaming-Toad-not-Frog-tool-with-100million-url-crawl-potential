package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const directoryMode = 0o700

type Paths struct {
	Data      string
	Cache     string
	Artifacts string
}

// ResolvePaths returns application-owned directories. Environment overrides
// are intended for trusted local startup configuration, never crawl or MCP input.
func ResolvePaths() (Paths, error) {
	data, err := resolve("SEO_AUDITOR_DATA_DIR", defaultDataDir)
	if err != nil {
		return Paths{}, err
	}
	cache, err := resolve("SEO_AUDITOR_CACHE_DIR", defaultCacheDir)
	if err != nil {
		return Paths{}, err
	}
	artifacts := os.Getenv("SEO_AUDITOR_ARTIFACT_DIR")
	if artifacts == "" {
		artifacts = filepath.Join(data, "artifacts")
	}
	artifacts, err = cleanAbsolute(artifacts)
	if err != nil {
		return Paths{}, fmt.Errorf("artifact directory: %w", err)
	}
	return Paths{Data: data, Cache: cache, Artifacts: artifacts}, nil
}

func (p Paths) Ensure() error {
	for _, dir := range []string{p.Data, p.Cache, p.Artifacts} {
		if err := os.MkdirAll(dir, directoryMode); err != nil {
			return fmt.Errorf("create application directory: %w", err)
		}
		if runtime.GOOS != "windows" {
			if err := os.Chmod(dir, directoryMode); err != nil {
				return fmt.Errorf("secure application directory: %w", err)
			}
		}
	}
	return nil
}

func resolve(env string, fallback func() (string, error)) (string, error) {
	value := os.Getenv(env)
	if value == "" {
		var err error
		value, err = fallback()
		if err != nil {
			return "", err
		}
	}
	return cleanAbsolute(value)
}

func cleanAbsolute(path string) (string, error) {
	if path == "" {
		return "", errors.New("path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func defaultDataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(base, "seo-auditor"), nil
}

func defaultCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache directory: %w", err)
	}
	return filepath.Join(base, "seo-auditor"), nil
}
