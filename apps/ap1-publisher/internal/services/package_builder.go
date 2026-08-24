package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/local/dicom-disc-suite/shared/models"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type StudyPackageBuilder struct {
	Repository   StudyRetriever
	TempRoot     string
	ViewerSource string
	Logger       *slog.Logger
}

type StudyRetriever interface {
	RetrieveStudy(context.Context, string, string) error
}

func (b *StudyPackageBuilder) Build(ctx context.Context, study models.Study) (models.DiscJob, error) {
	created := time.Now()
	now := models.DiscJob{ID: fmt.Sprintf("job-%d", created.UnixNano()), StudyInstanceUID: study.StudyInstanceUID, PatientID: study.PatientID, PatientName: study.PatientName, Status: models.Downloading, CreatedAt: created, UpdatedAt: created}
	root := filepath.Join(b.TempRoot, study.StudyInstanceUID)
	now.TempPath = root
	now.DataPath = filepath.Join(root, "data")
	now.ViewerPath = filepath.Join(root, "AP2")
	now.ManifestPath = filepath.Join(root, "study.json")
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
	if b.ViewerSource != "" {
		if _, err := os.Stat(b.ViewerSource); err == nil {
			if err = copyFile(b.ViewerSource, filepath.Join(now.ViewerPath, filepath.Base(b.ViewerSource))); err != nil {
				return now, err
			}
		} else {
			os.WriteFile(filepath.Join(now.ViewerPath, "README.txt"), []byte("AP2 executable will be copied here after its production build.\n"), 0644)
		}
	}
	manifest := models.StudyManifest{
		StudyID:          study.StudyID,
		ESTUID:           study.StudyInstanceUID,
		StudyInstanceUID: study.StudyInstanceUID,
		PatientID:        study.PatientID,
		PatientName:      study.PatientName,
		Patient: models.ManifestPatient{
			ID:   study.PatientID,
			Name: study.PatientName,
		},
		StudyDescription: study.StudyDescription,
		Modality:         study.Modality,
		StudyDate:        study.StudyDate,
		TotalImages:      study.InstanceCount,
		Series:           study.Series,
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return now, err
	}
	if err = os.WriteFile(now.ManifestPath, raw, 0644); err != nil {
		return now, err
	}
	b.Logger.Info("Study manifest created", "job_id", now.ID)
	if err = GenerateDiscLabel(now.LabelPath, study); err != nil {
		return now, fmt.Errorf("generate disc label: %w", err)
	}
	now.Status = models.Ready
	return now, nil
}
func copyFile(src, dst string) error {
	in, e := os.ReadFile(src)
	if e != nil {
		return e
	}
	return os.WriteFile(dst, in, 0755)
}
