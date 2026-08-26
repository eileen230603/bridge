package main

import (
	"context"
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
