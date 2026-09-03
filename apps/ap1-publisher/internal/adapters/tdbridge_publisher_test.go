package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/local/dicom-disc-suite/shared/models"
)

func TestCreateJobStagesRealJDF(t *testing.T) {
    root := t.TempDir()
    job := validDiscJob(t, root)
    staging := filepath.Join(root, "staging")
    p := TdBridgePublisher{
        StagingDirectory: staging, 
        DefaultCopies: 2,
    DiscType: "DVD", Format: "UDF102",}
    path, err := p.CreateJob(context.Background(), job)
    if err != nil {
        t.Fatal(err)
    }
    raw, err := os.ReadFile(path)
    if err != nil {
        t.Fatal(err)
    }
    text := string(raw)
    for _, want := range []string{
        "JOB_ID=job-123\r\n", 
        "COPIES=2\r\n", 
        "DISC_TYPE=DVD\r\n", 
        "FORMAT=UDF102\r\n", 
        "\tdata\r\n", 
        "\tstudy.dat\r\n", 
        "\tautorun.inf\r\n", 
        "LABEL=" + job.LabelPath + "\r\n",
    } {
        if !strings.Contains(text, want) {
            t.Fatalf("JDF missing %q:\n%s", want, text)
        }
    }
    if strings.Contains(text, "\tlabel\r\n") {
        t.Fatalf("label directory included as disc content:\n%s", text)
    }
    if strings.Contains(text, "\tAP2\r\n") {
        t.Fatalf("AP2 directory should not be created on disc root:\n%s", text)
    }
}

func TestCreateJobRejectsMissingPackage(t *testing.T) {
    p := TdBridgePublisher{StagingDirectory: t.TempDir()}
    _, err := p.CreateJob(context.Background(), models.DiscJob{ID: "job-1"})
    if err == nil || err.Error() != "study package does not exist" {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestSubmitJobRejectsMissingMonitoringFolder(t *testing.T) {
    file := filepath.Join(t.TempDir(), "job.jdf")
    if err := os.WriteFile(file, []byte("finished"), 0o644); err != nil {
        t.Fatal(err)
    }
    p := TdBridgePublisher{MonitoringFolder: filepath.Join(t.TempDir(), "missing")}
    if err := p.SubmitJob(context.Background(), file); err == nil || !strings.Contains(err.Error(), "does not exist") {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestSubmitJobRejectsMissingJDF(t *testing.T) {
    p := TdBridgePublisher{MonitoringFolder: t.TempDir()}
    if err := p.SubmitJob(context.Background(), filepath.Join(t.TempDir(), "missing.jdf")); err == nil || err.Error() != "JDF file does not exist" {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestSubmitJobPublishesFinishedFileAtomically(t *testing.T) {
    root := t.TempDir()
    monitoring := filepath.Join(root, "test-tdbridge-orders")
    if err := os.Mkdir(monitoring, 0o755); err != nil {
        t.Fatal(err)
    }
    staged := filepath.Join(root, "job-123.jdf")
    want := []byte("already-built-real-jdf")
    if err := os.WriteFile(staged, want, 0o644); err != nil {
        t.Fatal(err)
    }
    p := TdBridgePublisher{MonitoringFolder: monitoring}
    if err := p.SubmitJob(context.Background(), staged); err != nil {
        t.Fatal(err)
    }
    got, err := os.ReadFile(filepath.Join(monitoring, "job-123.jdf"))
    if err != nil || string(got) != string(want) {
        t.Fatalf("published file mismatch: %q, %v", got, err)
    }
    if _, err := os.Stat(filepath.Join(monitoring, "job-123.jdf.tmp")); !os.IsNotExist(err) {
        t.Fatalf("partial file remains: %v", err)
    }
}

func validDiscJob(t *testing.T, root string) models.DiscJob {
    t.Helper()
    packageRoot := filepath.Join(root, "study")
    job := models.DiscJob{
        ID:           "job-123",
        TempPath:     packageRoot,
        DataPath:     filepath.Join(packageRoot, "data"),
        ViewerPath:   packageRoot,
        ManifestPath: filepath.Join(packageRoot, "study.dat"),
        LabelPath:    filepath.Join(packageRoot, "label", "label.png"),
    }
    for _, dir := range []string{job.DataPath, job.ViewerPath, filepath.Dir(job.LabelPath)} {
        if err := os.MkdirAll(dir, 0o755); err != nil {
            t.Fatal(err)
        }
    }
    if err := os.WriteFile(job.ManifestPath, []byte("{}"), 0o644); err != nil {
        t.Fatal(err)
    }
    if err := os.WriteFile(job.LabelPath, []byte("png"), 0o644); err != nil {
        t.Fatal(err)
    }
    autorunPath := filepath.Join(packageRoot, "autorun.inf")
    if err := os.WriteFile(autorunPath, []byte("[autorun]\r\n"), 0o644); err != nil {
        t.Fatal(err)
    }
    return job
}