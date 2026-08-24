package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
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
