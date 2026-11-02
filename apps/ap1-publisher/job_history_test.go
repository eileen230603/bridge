package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/services"
	"github.com/local/dicom-disc-suite/shared/models"
)

func TestRestartRestoresHistoryAndReconcilesEpson(t *testing.T) {
	root := t.TempDir()
	orders := filepath.Join(root, "orders")
	if err := os.MkdirAll(orders, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{TemporaryDirectory: "temp", CompletedDirectory: "done", HistoryFile: "history.sqlite", Epson: config.EpsonConfig{Enabled: true, MonitoringFolder: orders, StagingDirectory: "staging"}}
	raw, _ := json.Marshal(cfg)
	file := filepath.Join(root, "config.json")
	if err := os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AP1_CONFIG", file)
	history, err := services.OpenJobHistory(filepath.Join(root, "history.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	for _, job := range []models.DiscJob{
		{ID: "queued", Status: models.QueuedForEpson, CreatedAt: time.Now()},
		{ID: "interrupted", Status: models.Downloading, CreatedAt: time.Now()},
		{ID: "finished", Status: models.Completed, CreatedAt: time.Now()},
	} {
		if err := history.Save(job); err != nil {
			t.Fatal(err)
		}
	}
	history.Close()
	if err := os.WriteFile(filepath.Join(orders, "queued.DON"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	// A stale marker must not regress terminal history.
	if err := os.WriteFile(filepath.Join(orders, "finished.JDF"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	app, err := NewApp()
	if err != nil {
		t.Fatal(err)
	}
	app.ctx = context.Background()
	jobs := app.ListJobs()
	states := map[string]models.JobStatus{}
	for _, job := range jobs {
		states[job.ID] = job.Status
	}
	if states["queued"] != models.Completed || states["interrupted"] != models.Failed || states["finished"] != models.Completed {
		t.Fatalf("unexpected states: %v", states)
	}
	app.closeHistory()
	reopened, err := NewApp()
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.closeHistory()
	if len(reopened.jobs) != 3 {
		t.Fatalf("unexpected restored count %d", len(reopened.jobs))
	}
	for _, job := range reopened.jobs {
		if job.Status != states[job.ID] {
			t.Fatal("updated status not persisted")
		}
	}
}

func TestHistoryWriteFailureDoesNotPretendJobWasSaved(t *testing.T) {
	store, err := services.OpenJobHistory(filepath.Join(t.TempDir(), "history.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	store.Close()
	app := &App{history: store}
	if err := app.recordJob(models.DiscJob{ID: "test"}); err == nil {
		t.Fatal("write failure ignored")
	}
	if app.GetHistoryError() == "" || len(app.jobs) != 0 {
		t.Fatal("history failure hidden")
	}
}

func TestEmbeddedDefaultsDoNotContainInstallationOrActivation(t *testing.T) {
	var cfg config.Config
	if err := json.Unmarshal(defaultConfig, &cfg); err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(cfg.TemporaryDirectory) || filepath.IsAbs(cfg.Epson.StagingDirectory) || cfg.LicenseToken != "" || cfg.LicenseKey != "" || cfg.DiscLabel.LogoPath != "" {
		t.Fatal("machine-specific defaults embedded")
	}
}

func TestConfigBesideExecutableWinsOverWorkingDirectory(t *testing.T) {
	installation, working := t.TempDir(), t.TempDir()
	for _, dir := range []string{installation, working} {
		if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	got, found := resolveAP1ConfigPath("", filepath.Join(installation, "AP1.exe"), working)
	if !found || got != filepath.Join(installation, "config.json") {
		t.Fatalf("unexpected config %s", got)
	}
}
