package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
)

func TestSaveEpsonPerformancePersistsAndKeepsActiveSnapshot(t *testing.T) {
	root := t.TempDir()
	original := &adapters.TdBridgePublisher{MonitoringFolder: root, StagingDirectory: root, DiscType: "DVD", Format: "UDF102", DefaultCopies: 2}
	app := &App{
		configPath: filepath.Join(root, "config.json"),
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		publisher:  original,
		cfg:        config.Config{Epson: config.EpsonConfig{MonitoringFolder: root, StagingDirectory: root, Enabled: true, DefaultCopies: 2}},
	}
	if err := app.SaveEpsonConfig(config.EpsonConfig{DiscType: "DVD", Format: "UDF102", PrintMode: 2, LabelType: 1}); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(app.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Epson.PrintMode != 2 || loaded.Epson.LabelType != 1 || loaded.Epson.DefaultCopies != 2 || !loaded.Epson.Enabled || loaded.Epson.MonitoringFolder != root {
		t.Fatalf("settings lost: %+v", loaded.Epson)
	}
	next := app.publisher.(*adapters.TdBridgePublisher)
	if next == original || original.PrintMode != 0 || next.PrintMode != 2 || next.DefaultCopies != 2 {
		t.Fatal("settings must apply to future jobs while preserving active snapshots")
	}
	before, err := os.ReadFile(app.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.SaveEpsonConfig(config.EpsonConfig{PrintMode: 2, LabelType: 3}); err == nil {
		t.Fatal("invalid settings accepted")
	}
	after, err := os.ReadFile(app.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) || app.publisher != next {
		t.Fatal("invalid settings changed persisted or active state")
	}
}
