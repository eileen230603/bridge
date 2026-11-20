package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

var semver = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)

func TestVersionIsSemVer(t *testing.T) {
	if !semver.MatchString(Version) {
		t.Fatalf("Version %q no es SemVer (MAYOR.MENOR.PARCHE[-prerelease])", Version)
	}
}

func TestVersionMatchesVersionFile(t *testing.T) {
	raw, err := os.ReadFile("../../VERSION")
	if err != nil {
		t.Fatalf("leer VERSION: %v", err)
	}
	if got := strings.TrimSpace(string(raw)); got != Version {
		t.Fatalf("VERSION=%q pero version.go=%q; actualice ambos (ver docs/versioning.md)", got, Version)
	}
}

func TestVersionMatchesWailsJSON(t *testing.T) {
	raw, err := os.ReadFile("wails.json")
	if err != nil {
		t.Fatalf("leer wails.json: %v", err)
	}
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("wails.json inválido: %v", err)
	}
	if cfg.Info.ProductVersion != Version {
		t.Fatalf("wails.json productVersion=%q pero version.go=%q", cfg.Info.ProductVersion, Version)
	}
}
