package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/local/dicom-disc-suite/shared/models"
)

type StudyPackageBuilder struct {
	Repository   StudyRetriever
	TempRoot     string
	ViewerBuilds fs.FS
	Logger       *slog.Logger
}

type StudyRetriever interface {
	RetrieveStudy(context.Context, string, string) error
}

func (b *StudyPackageBuilder) Build(ctx context.Context, study models.Study) (models.DiscJob, error) {
	created := time.Now()
	now := models.DiscJob{
		ID:               fmt.Sprintf("job-%d", created.UnixNano()),
		StudyInstanceUID: study.StudyInstanceUID,
		PatientID:        study.PatientID,
		PatientName:      study.PatientName,
		StudyDescription: study.StudyDescription,
		Status:           models.Downloading,
		CreatedAt:        created,
		UpdatedAt:        created,
	}

	root := filepath.Join(b.TempRoot, study.StudyInstanceUID)
	now.TempPath = root
	now.DataPath = filepath.Join(root, "data")
	now.ViewerPath = filepath.Join(root, "AP2")
	now.ManifestPath = filepath.Join(root, "study.dat")
	now.LabelPath = filepath.Join(root, "label", "label.png")

	b.Logger.Info("Preparing study package", "study_uid", study.StudyInstanceUID, "job_id", now.ID)

	if err := os.RemoveAll(root); err != nil {
		return now, err
	}

	for _, p := range []string{now.DataPath, now.ViewerPath, filepath.Dir(now.LabelPath)} {
		if err := os.MkdirAll(p, 0755); err != nil {
			return now, err
		}
	}

	b.Logger.Info("Starting HTTP study download", "study_uid", study.StudyInstanceUID)
	if err := b.Repository.RetrieveStudy(ctx, study.StudyInstanceUID, now.DataPath); err != nil {
		return now, fmt.Errorf("download study: %w", err)
	}

	now.Status = models.Preparing

	if err := extractViewerBuilds(b.ViewerBuilds, now.ViewerPath); err != nil {
		return now, fmt.Errorf("extract embedded AP2 viewer builds: %w", err)
	}

	// 1. Mapear la estructura del estudio
	manifest := MapSymphonyStudyToViewerStudy(study)

	// 2. Convertir a JSON
	jsonBytes, err := json.Marshal(manifest)
	if err != nil {
		return now, fmt.Errorf("marshal study manifest: %w", err)
	}

	// 3. Cifrar usando la función de crypto_utils.go
	encryptedBytes, err := EncryptData(jsonBytes, SecretKey)
	if err != nil {
		return now, fmt.Errorf("encrypt study manifest: %w", err)
	}

	// 4. Guardar como binario cifrado
	if err = os.WriteFile(now.ManifestPath, encryptedBytes, 0644); err != nil {
		return now, fmt.Errorf("write study.dat: %w", err)
	}
	b.Logger.Info("Encrypted study manifest created", "job_id", now.ID)

	if err = GenerateDiscLabel(now.LabelPath, study); err != nil {
		return now, fmt.Errorf("generate disc label: %w", err)
	}

	now.Status = models.Ready
	b.Logger.Info("Study package prepared", "job_id", now.ID)
	return now, nil
}

const (
	windowsViewer = "viewer-builds/windows/Symphony Viewer.exe"
	macOSViewer   = "viewer-builds/macos/Symphony Viewer.app"
)

func extractViewerBuilds(builds fs.FS, destination string) error {
	if builds == nil {
		return fmt.Errorf("embedded viewer filesystem is not configured")
	}
	if info, err := fs.Stat(builds, windowsViewer); err != nil || !info.Mode().IsRegular() {
		if err == nil {
			err = fmt.Errorf("not a regular file")
		}
		return fmt.Errorf("embedded Windows build %q is missing: %w", windowsViewer, err)
	}
	if info, err := fs.Stat(builds, macOSViewer); err != nil || !info.IsDir() {
		if err == nil {
			err = fmt.Errorf("not a directory")
		}
		return fmt.Errorf("embedded macOS build %q is missing: %w", macOSViewer, err)
	}

	return fs.WalkDir(builds, "viewer-builds", func(source string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel("viewer-builds", filepath.FromSlash(source))
		if err != nil || relative == "." {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if source == "viewer-builds/windows/README.txt" || source == "viewer-builds/macos/README.txt" {
			return nil
		}
		content, err := fs.ReadFile(builds, source)
		if err != nil {
			return fmt.Errorf("read embedded viewer file %q: %w", source, err)
		}
		mode := fs.FileMode(0o644)
		if source == windowsViewer || strings.Contains(source, "/Contents/MacOS/") {
			mode = 0o755
		}
		if err := os.WriteFile(target, content, mode); err != nil {
			return fmt.Errorf("write viewer file %q: %w", target, err)
		}
		return nil
	})
}
