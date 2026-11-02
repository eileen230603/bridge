package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRelativePathsSurviveSaveAndRelocation(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	raw := []byte(`{"temporaryDirectory":"runtime/temp","completedDirectory":"runtime/completed","logFile":"runtime/logs/ap1.log","historyFile":"runtime/history.sqlite","epson":{"monitoringFolder":"runtime/custom-orders","stagingDirectory":"runtime/staging"},"discLabel":{"logoPath":"assets/logo.png"}}`)
	cfg, err := LoadBytes(raw, first)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Epson.MonitoringFolder != filepath.Join(first, "runtime", "custom-orders") {
		t.Fatal("explicit monitoring path was overwritten")
	}
	cfg.StudyAPI.Host = "pacs.local"
	file := filepath.Join(first, "config.json")
	if err := Save(file, cfg); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var persisted Config
	if err := json.Unmarshal(saved, &persisted); err != nil {
		t.Fatal(err)
	}
	for key, value := range pathFields(&persisted) {
		if filepath.IsAbs(*value) {
			t.Fatalf("%s became absolute: %s", key, *value)
		}
	}
	if strings.Contains(string(saved), first) {
		t.Fatal("installation path leaked into config")
	}
	moved, err := LoadBytes(saved, second)
	if err != nil {
		t.Fatal(err)
	}
	if moved.TemporaryDirectory != filepath.Join(second, "runtime", "temp") || moved.HistoryFile != filepath.Join(second, "runtime", "history.sqlite") || moved.DiscLabel.LogoPath != filepath.Join(second, "assets", "logo.png") {
		t.Fatal("paths did not follow installation")
	}
}

func TestExplicitExternalMonitoringPathSurvivesSave(t *testing.T) {
	base, orders := t.TempDir(), t.TempDir()
	raw, _ := json.Marshal(Config{TemporaryDirectory: "temp", CompletedDirectory: "done", Epson: EpsonConfig{MonitoringFolder: orders}})
	cfg, err := LoadBytes(raw, base)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(base, "config.json")
	if err := Save(file, cfg); err != nil {
		t.Fatal(err)
	}
	moved, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Epson.MonitoringFolder != orders {
		t.Fatal("external Epson folder changed")
	}
}

func TestNewHospitalLogoIsSavedRelativeToConfiguration(t *testing.T) {
	base := t.TempDir()
	cfg, err := LoadBytes([]byte(`{"temporaryDirectory":"temp","completedDirectory":"done"}`), base)
	if err != nil {
		t.Fatal(err)
	}
	cfg.DiscLabel.LogoPath = filepath.Join(base, "assets", "hospital-logo.png")
	file := filepath.Join(base, "config.json")
	if err := Save(file, cfg); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var persisted Config
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.DiscLabel.LogoPath != filepath.Join("assets", "hospital-logo.png") {
		t.Fatal("new logo path was not portable")
	}
}
