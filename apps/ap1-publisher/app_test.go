package main

import (
	"context"
	"embed"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/shared/models"
)

func TestNewAppDoesNotValidateViewerBuildsDuringStartup(t *testing.T) {
	originalViewerBuilds := viewerBuilds
	viewerBuilds = embed.FS{}
	defer func() { viewerBuilds = originalViewerBuilds }()

	configPath := filepath.Join(t.TempDir(), "config.json")
	configJSON := `{
  "temporaryDirectory": "temp",
  "completedDirectory": "completed",
  "studyApi": {"protocol": "http", "host": "127.0.0.1", "port": 4000, "timeoutSeconds": 1},
  "epson": {"enabled": true, "monitoringFolder": "epson", "stagingDirectory": "staging", "defaultCopies": 1}
}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AP1_CONFIG", configPath)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("AP1 startup must not depend on embedded viewers: %v", err)
	}
	if app.builder.ViewerBuilds == nil {
		t.Fatal("viewer filesystem must be passed through for package-time validation")
	}
}

func TestResolveAP1ConfigPathFindsConfigRelativeToExecutable(t *testing.T) {
	appRoot := t.TempDir()
	configPath := filepath.Join(appRoot, "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(appRoot, "build", "bin", "DICOM Disc Publisher.exe")

	got, found := resolveAP1ConfigPath("", executable, filepath.Join(t.TempDir(), "unrelated-working-directory"))
	if !found {
		t.Fatal("expected executable-relative config to be found")
	}
	if got != configPath {
		t.Fatalf("expected executable-relative config %q, got %q", configPath, got)
	}
}

func TestResolveAP1ConfigPathPrefersExplicitPath(t *testing.T) {
	explicit := filepath.Join(t.TempDir(), "custom.json")
	if got, found := resolveAP1ConfigPath(explicit, "ignored.exe", "ignored"); !found || got != explicit {
		t.Fatalf("expected explicit config %q, got %q", explicit, got)
	}
}

func TestNewAppUsesEmbeddedConfigWhenNoExternalFileExists(t *testing.T) {
	t.Setenv("AP1_CONFIG", "")
	workDir := t.TempDir()
	executable := filepath.Join(workDir, "DICOM Disc Publisher.exe")
	if path, found := resolveAP1ConfigPath("", executable, workDir); found || path != "" {
		t.Fatalf("expected no external config, got path=%q found=%v", path, found)
	}

	cfg, err := config.LoadBytes(defaultConfig, filepath.Join(workDir, "apps", "ap1-publisher"))
	if err != nil {
		t.Fatalf("embedded config must be valid: %v", err)
	}
	if cfg.TemporaryDirectory != filepath.Join(workDir, "runtime", "temp") {
		t.Fatalf("unexpected portable runtime path: %q", cfg.TemporaryDirectory)
	}
}

func TestListJobsUpdatesEachJobFromTdBridge(t *testing.T) {
	folder := t.TempDir()
	for name, content := range map[string]string{"job-a.DON": "", "job-b.ERR": "Bridge error", "unrelated.ERR": "ignore"} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(folder, "TDBStatus.txt"), []byte("[job-b]\nSTATUS=6\nERROR=JDF0203\nDETAIL_STATUS=14\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &App{
		ctx: context.Background(), monitor: adapters.TdBridgeJobMonitor{MonitoringFolder: folder},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		jobs:   []models.DiscJob{{ID: "job-a", Status: models.QueuedForEpson}, {ID: "job-b", Status: models.QueuedForEpson}, {ID: "job-c", Status: models.QueuedForEpson}},
	}
	jobs := app.ListJobs()
	if jobs[0].Status != models.Completed || jobs[1].Status != models.Failed || jobs[1].ErrorCode != "JDF0203" || jobs[1].ErrorMessage != "No hay una Epson Discproducer configurada." || jobs[2].Status != models.QueuedForEpson {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
}

func TestListJobsConcurrentAccess(t *testing.T) {
	app := &App{ctx: context.Background(), logger: slog.New(slog.NewTextHandler(io.Discard, nil)), jobs: []models.DiscJob{{ID: "job-a", Status: models.QueuedForEpson}}}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = app.ListJobs() }()
	}
	wg.Wait()
}

func TestServerConnectionSuccessHTTPErrorAndTimeout(t *testing.T) {
	tests := []struct {
		name                    string
		handler                 http.HandlerFunc
		wantStatus, wantMessage string
	}{
		{"success", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }, "Conectado", ""},
		{"http error", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadGateway) }, "Error", "No se pudo conectar al servidor."},
		{"timeout", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}, "Error", "Tiempo de espera agotado."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()
			u, _ := url.Parse(server.URL)
			port, _ := strconv.Atoi(u.Port())
			app := &App{}
			result := app.TestServerConnection(config.ServerConfig{Protocol: "http", Host: u.Hostname(), Port: port, TimeoutSeconds: 1})
			if result.Status != tt.wantStatus || result.Message != tt.wantMessage {
				t.Fatalf("got %+v", result)
			}
		})
	}
}
