package services

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/local/dicom-disc-suite/shared/models"
)

func TestRepeatedStudyKeepsQueuedPackage(t *testing.T) {
	b := StudyPackageBuilder{Repository: testStudyRetriever{}, TempRoot: t.TempDir(), ViewerBuilds: testViewerBuilds(), Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	study := models.Study{StudyInstanceUID: "same-study"}
	first, err := b.Build(context.Background(), study)
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(first.TempPath, "queued-marker")
	if err := os.WriteFile(marker, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := b.Build(context.Background(), study)
	if err != nil {
		t.Fatal(err)
	}
	if first.TempPath == second.TempPath || first.ID == second.ID {
		t.Fatal("jobs share their package")
	}
	if len(second.ID) > 40 {
		t.Fatal("job ID exceeds TD Bridge limit")
	}
	raw, err := os.ReadFile(marker)
	if err != nil || string(raw) != "original" {
		t.Fatal("queued package was modified")
	}
}

type blockingDownload struct{ extracted <-chan struct{} }

func (r blockingDownload) RetrieveStudy(ctx context.Context, _ string, destination string) error {
	select {
	case <-r.extracted:
		return os.WriteFile(filepath.Join(destination, "instance.dcm"), []byte("DICOM"), 0600)
	case <-ctx.Done():
		return ctx.Err()
	}
}

type observedViewerFS struct {
	fs.FS
	extracted chan struct{}
	once      sync.Once
}

func (f *observedViewerFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if err == nil && name == windowsViewer {
		f.once.Do(func() { close(f.extracted) })
	}
	return file, err
}
func TestViewerPreparationOverlapsDownload(t *testing.T) {
	extracted := make(chan struct{})
	b := StudyPackageBuilder{
		Repository:   blockingDownload{extracted},
		ViewerBuilds: &observedViewerFS{FS: testViewerBuilds(), extracted: extracted},
		TempRoot:     t.TempDir(), Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := b.Build(ctx, models.Study{StudyInstanceUID: "parallel"}); err != nil {
		t.Fatal(err)
	}
}
func TestViewerFailureCancelsAndJoinsDownload(t *testing.T) {
	b := StudyPackageBuilder{
		Repository: blockingDownload{make(chan struct{})},
		TempRoot:   t.TempDir(), Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := b.Build(ctx, models.Study{StudyInstanceUID: "invalid-viewer"})
	if err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected extraction error, got %v", err)
	}
	if ctx.Err() != nil {
		t.Fatal("download was not cancelled promptly")
	}
}
