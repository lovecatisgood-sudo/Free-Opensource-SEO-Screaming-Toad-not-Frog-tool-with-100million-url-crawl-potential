package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const directoryMode = 0o700

type Paths struct {
	Data      string
	Cache     string
	Artifacts string
}

type Server struct {
	Host string
	Port int
}

type Renderer struct {
	Enabled          bool
	NodeBinary       string
	ScriptPath       string
	BrowserPath      string
	ContainerSandbox bool
}

// ResolveRenderer reads only trusted process-start configuration. None of
// these paths can be supplied by a crawl profile, API request, or MCP tool.
func ResolveRenderer() (Renderer, error) {
	rawScript := os.Getenv("SEO_AUDITOR_RENDERER_SCRIPT")
	if rawScript == "" {
		return Renderer{}, nil
	}
	script, err := cleanAbsolute(rawScript)
	if err != nil {
		return Renderer{}, fmt.Errorf("renderer script: %w", err)
	}
	browserPath := os.Getenv("PLAYWRIGHT_BROWSERS_PATH")
	if browserPath != "" {
		browserPath, err = cleanAbsolute(browserPath)
		if err != nil {
			return Renderer{}, fmt.Errorf("renderer browser path: %w", err)
		}
	}
	node := os.Getenv("SEO_AUDITOR_NODE_BINARY")
	if node == "" {
		node = "node"
	}
	containerValue := os.Getenv("SEO_AUDITOR_CONTAINER_SANDBOX")
	if containerValue != "" && containerValue != "0" && containerValue != "1" {
		return Renderer{}, errors.New("SEO_AUDITOR_CONTAINER_SANDBOX must be 0 or 1")
	}
	return Renderer{Enabled: true, NodeBinary: node, ScriptPath: script, BrowserPath: browserPath, ContainerSandbox: containerValue == "1"}, nil
}

func ResolveServer() (Server, error) {
	host := os.Getenv("SEO_AUDITOR_BIND_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	address := net.ParseIP(host)
	if address == nil || !address.IsLoopback() {
		return Server{}, errors.New("API bind host must be a numeric loopback address")
	}
	port := 7331
	if raw := os.Getenv("SEO_AUDITOR_BIND_PORT"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 65535 {
			return Server{}, errors.New("API bind port is invalid")
		}
		port = value
	}
	return Server{Host: host, Port: port}, nil
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
