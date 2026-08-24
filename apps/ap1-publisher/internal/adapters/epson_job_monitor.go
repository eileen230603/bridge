package adapters

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

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

// TdBridgeJobMonitor implements the minimum status mechanism documented by
// Epson: TD Bridge renames the submitted job as it advances.
type TdBridgeJobMonitor struct {
	MonitoringFolder string
}

func (m TdBridgeJobMonitor) GetStatus(ctx context.Context, job models.DiscJob) (models.JobStatus, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return job.Status, err
		}
	}
	base := strings.TrimSuffix(filepath.Base(job.EpsonJobPath), filepath.Ext(job.EpsonJobPath))
	if base == "" || base == "." {
		base = job.ID
	}
	for _, candidate := range []struct {
		extension string
		status    models.JobStatus
	}{
		{".ERR", models.Failed},
		{".DON", models.Completed},
		{".INP", models.Publishing},
		{".STP", models.Publishing},
		{".RJD", models.QueuedForEpson},
		{".JDF", models.QueuedForEpson},
	} {
		_, err := os.Stat(filepath.Join(m.MonitoringFolder, base+candidate.extension))
		if err == nil {
			return candidate.status, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return job.Status, err
		}
	}
	return job.Status, nil
}
