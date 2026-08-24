package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/local/dicom-disc-suite/shared/models"
)

func TestTdBridgeJobMonitorMapsDocumentedExtensions(t *testing.T) {
	for _, tt := range []struct {
		extension string
		want      models.JobStatus
	}{
		{"JDF", models.QueuedForEpson},
		{"RJD", models.QueuedForEpson},
		{"INP", models.Publishing},
		{"STP", models.Publishing},
		{"DON", models.Completed},
		{"ERR", models.Failed},
	} {
		t.Run(tt.extension, func(t *testing.T) {
			folder := t.TempDir()
			if err := os.WriteFile(filepath.Join(folder, "job-123."+tt.extension), []byte("status"), 0o644); err != nil {
				t.Fatal(err)
			}
			monitor := TdBridgeJobMonitor{MonitoringFolder: folder}
			got, err := monitor.GetStatus(context.Background(), models.DiscJob{ID: "job-123", EpsonJobPath: "job-123.jdf", Status: models.QueuedForEpson})
			if err != nil || got != tt.want {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}
