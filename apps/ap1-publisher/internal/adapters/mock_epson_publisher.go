package adapters

import (
	"context"
	"encoding/json"
	"github.com/local/dicom-disc-suite/shared/filesystem"
	"github.com/local/dicom-disc-suite/shared/models"
	"log/slog"
	"os"
	"path/filepath"
)

type EpsonPublisher interface {
	CreateJob(context.Context, models.DiscJob) (string, error)
	SubmitJob(context.Context, string) error
}
type MockEpsonPublisher struct {
	StagingDirectory, HotFolder string
	Logger                      *slog.Logger
}

func (m *MockEpsonPublisher) CreateJob(_ context.Context, j models.DiscJob) (string, error) {
	m.Logger.Info("Preparing Epson job", "job_id", j.ID)
	if err := os.MkdirAll(m.StagingDirectory, 0755); err != nil {
		return "", err
	}
	p := filepath.Join(m.StagingDirectory, j.ID+".mock-job")
	b, _ := json.MarshalIndent(map[string]string{"notice": "MOCK ONLY - format is not an Epson specification", "jobId": j.ID, "packagePath": j.TempPath}, "", "  ")
	if err := os.WriteFile(p, b, 0644); err != nil {
		return "", err
	}
	m.Logger.Info("Epson job created", "job_id", j.ID)
	return p, nil
}
func (m *MockEpsonPublisher) SubmitJob(_ context.Context, p string) error {
	m.Logger.Info("Moving job to Epson hot folder")
	dest := filepath.Join(m.HotFolder, filepath.Base(p))
	if err := filesystem.MoveFileSafely(p, dest); err != nil {
		return err
	}
	m.Logger.Info("Job submitted successfully")
	return nil
}
