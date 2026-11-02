package services

import (
	"context"
	"errors"
	"github.com/local/dicom-disc-suite/shared/models"
	"testing"
)

func TestHistoryFailureStopsPackageBeforeDownload(t *testing.T) {
	sentinel := errors.New("cannot persist")
	builder := StudyPackageBuilder{TempRoot: t.TempDir(), OnJobCreated: func(job models.DiscJob) error {
		if job.ID == "" || job.TempPath == "" {
			t.Fatal("incomplete initial snapshot")
		}
		return sentinel
	}}
	// Repository and logger are intentionally nil: neither should be used before persistence succeeds.
	_, err := builder.Build(context.Background(), models.Study{StudyInstanceUID: "test"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected persistence error, got %v", err)
	}
}
