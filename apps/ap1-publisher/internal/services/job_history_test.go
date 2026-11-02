package services

import (
	"github.com/local/dicom-disc-suite/shared/models"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJobHistorySurvivesReopenAndUpdatesWithoutDuplicates(t *testing.T) {
	file := filepath.Join(t.TempDir(), "history.sqlite")
	store, err := OpenJobHistory(file)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	job := models.DiscJob{ID: "job-1", Status: models.QueuedForEpson, CreatedAt: now, UpdatedAt: now, PatientName: "Test", EpsonJobPath: "job-1.jdf"}
	if err := store.Save(job); err != nil {
		t.Fatal(err)
	}
	job.Status = models.Completed
	if err := store.Save(job); err != nil {
		t.Fatal(err)
	}
	failed := models.DiscJob{ID: "job-2", Status: models.Failed, CreatedAt: now.Add(time.Second), ErrorCode: "JDF0203", ErrorMessage: "Test failure"}
	if err := store.Save(failed); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenJobHistory(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	jobs, err := reopened.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].Status != models.Completed || jobs[0].PatientName != "Test" || jobs[1].ErrorCode != "JDF0203" {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
}

func TestJobHistoryRejectsCorruptDatabase(t *testing.T) {
	file := filepath.Join(t.TempDir(), "history.sqlite")
	if err := os.WriteFile(file, []byte("not a database"), 0600); err != nil {
		t.Fatal(err)
	}
	if store, err := OpenJobHistory(file); err == nil {
		store.Close()
		t.Fatal("corrupt database silently replaced")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "not a database" {
		t.Fatal("corrupt database was modified")
	}
}
