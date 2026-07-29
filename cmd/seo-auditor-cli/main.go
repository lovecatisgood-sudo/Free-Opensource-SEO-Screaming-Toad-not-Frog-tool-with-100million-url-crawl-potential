package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/seo-auditor/seo-auditor/internal/version"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "version" {
		fmt.Fprintln(os.Stderr, "usage: seo-auditor-cli version")
		os.Exit(2)
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]string{"version": version.Version, "commit": version.Commit})
}
