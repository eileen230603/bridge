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
		{"RJD", models.Processing},
		{"INP", models.Processing},
		{"STP", models.Processing},
		{"DON", models.Completed},
		{"ERR", models.Failed},
	} {
		t.Run(tt.extension, func(t *testing.T) {
			folder := t.TempDir()
			if err := os.WriteFile(filepath.Join(folder, "job-123."+tt.extension), []byte("status"), 0o644); err != nil {
				t.Fatal(err)
			}
			state, err := (TdBridgeJobMonitor{MonitoringFolder: folder}).GetStatus(context.Background(), models.DiscJob{ID: "job-123", Status: models.QueuedForEpson})
			if err != nil || state.Status != tt.want || state.Technical != tt.extension {
				t.Fatalf("got %+v, %v; want %q/%q", state, err, tt.want, tt.extension)
			}
		})
	}
}

func TestReadTDBStatusFindsExactJobSection(t *testing.T) {
	folder := t.TempDir()
	path := filepath.Join(folder, "TDBStatus")
	writeTDBStatus(t, path, "[ACTIVE_JOB]\n[COMPLETE_JOB]\nJOB1=job-other\n\n[job-other]\nSTATUS=5\nERROR=JDF9999\nDETAIL_STATUS=9\n\n[job-123]\nSTATUS=6\nERROR=JDF0203\nDETAIL_STATUS=14\n")
	sections, err := readTDBStatus(path)
	if err != nil {
		t.Fatal(err)
	}
	got := sections["job-123"]
	if got["STATUS"] != "6" || got["ERROR"] != "JDF0203" || got["DETAIL_STATUS"] != "14" {
		t.Fatalf("unexpected section: %+v", got)
	}
}

func TestTdBridgeJobMonitorMapsJDF0203WithoutExposingERRContent(t *testing.T) {
	folder := t.TempDir()
	const jobID = "job-1787585509311188200"
	technicalJDF := "JOB_ID=" + jobID + "\nCOPIES=1\nDISC_TYPE=CD\nDATA=C:\\patient\\study\nLABEL=C:\\patient\\label.png"
	if err := os.WriteFile(filepath.Join(folder, jobID+".ERR"), []byte(technicalJDF), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTDBStatus(t, filepath.Join(folder, "TDBStatus.txt"), "[ACTIVE_JOB]\n[COMPLETE_JOB]\nJOB1=job-1787585509311188200\n\n[job-1787585509311188200]\nSTATUS=6\nERROR=JDF0203\nDETAIL_STATUS=14\n\n[TDB_INFO]\nVERSION=10.0.1.0\n")
	state, err := (TdBridgeJobMonitor{MonitoringFolder: folder}).GetStatus(context.Background(), models.DiscJob{ID: jobID, Status: models.QueuedForEpson})
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != models.Failed || state.ErrorCode != "JDF0203" || state.TechnicalStatus != "6" || state.DetailStatus != "14" || state.ErrorMessage != "No hay una Epson Discproducer configurada." {
		t.Fatalf("unexpected state: %+v", state)
	}
	if state.ErrorMessage == technicalJDF {
		t.Fatal("ERR/JDF content leaked into ErrorMessage")
	}
}

func TestTdBridgeErrorMessageUnknownCode(t *testing.T) {
	if got := TdBridgeErrorMessage("JDF9999", ""); got != "TD Bridge reportó un error (código: JDF9999)." {
		t.Fatalf("unexpected message: %q", got)
	}
}

func TestTdBridgeJobMonitorJobWithoutErrorCode(t *testing.T) {
	folder := t.TempDir()
	if err := os.WriteFile(filepath.Join(folder, "job-123.ERR"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	writeTDBStatus(t, filepath.Join(folder, "TDBStatus"), "[job-123]\nSTATUS=6\nDETAIL_STATUS=14\n")
	state, err := (TdBridgeJobMonitor{MonitoringFolder: folder}).GetStatus(context.Background(), models.DiscJob{ID: "job-123", Status: models.QueuedForEpson})
	if err != nil || state.ErrorCode != "" || state.ErrorMessage != "TD Bridge reportó un error." {
		t.Fatalf("state=%+v err=%v", state, err)
	}
}

func TestTdBridgeJobMonitorIgnoresOtherJobs(t *testing.T) {
	folder := t.TempDir()
	if err := os.WriteFile(filepath.Join(folder, "job-old.ERR"), []byte("old failure"), 0o644); err != nil {
		t.Fatal(err)
	}
	job := models.DiscJob{ID: "job-current", Status: models.QueuedForEpson}
	state, err := (TdBridgeJobMonitor{MonitoringFolder: folder}).GetStatus(context.Background(), job)
	if err != nil || state.Status != job.Status || state.ErrorMessage != "" {
		t.Fatalf("unexpected state: %+v, %v", state, err)
	}
}

func TestTdBridgeJobMonitorKeepsJobsIndependent(t *testing.T) {
	folder := t.TempDir()
	for name, content := range map[string]string{"job-a.DON": "", "job-b.ERR": "failure", "job-c.JDF": ""} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	monitor := TdBridgeJobMonitor{MonitoringFolder: folder}
	wants := map[string]models.JobStatus{"job-a": models.Completed, "job-b": models.Failed, "job-c": models.QueuedForEpson}
	for id, want := range wants {
		state, err := monitor.GetStatus(context.Background(), models.DiscJob{ID: id, Status: models.QueuedForEpson})
		if err != nil || state.Status != want {
			t.Fatalf("job %s: state=%+v err=%v", id, state, err)
		}
	}
}

func TestTdBridgeJobMonitorErrorTakesPrecedence(t *testing.T) {
	folder := t.TempDir()
	for _, extension := range []string{"JDF", "INP", "ERR"} {
		if err := os.WriteFile(filepath.Join(folder, "job-123."+extension), []byte("failure"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	state, err := (TdBridgeJobMonitor{MonitoringFolder: folder}).GetStatus(context.Background(), models.DiscJob{ID: "job-123", Status: models.QueuedForEpson})
	if err != nil || state.Status != models.Failed {
		t.Fatalf("state=%+v err=%v", state, err)
	}
}

func writeTDBStatus(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
