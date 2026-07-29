package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/seo-auditor/seo-auditor/internal/config"
	"github.com/seo-auditor/seo-auditor/internal/version"
)

func main() {
	paths, err := config.ResolvePaths()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := paths.Ensure(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result := map[string]any{"name": "seo-auditor", "version": version.Version, "status": "ready", "data_dir": paths.Data}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
