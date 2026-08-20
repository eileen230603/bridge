package adapters

import (
	"context"

	"github.com/local/dicom-disc-suite/shared/models"
)

type EpsonJobMonitor interface {
	GetStatus(context.Context, models.DiscJob) (models.JobStatus, error)
}

// NoopEpsonJobMonitor preserves the last known state until STF/extension
// monitoring is implemented.
type NoopEpsonJobMonitor struct{}

func (NoopEpsonJobMonitor) GetStatus(_ context.Context, job models.DiscJob) (models.JobStatus, error) {
	return job.Status, nil
}
