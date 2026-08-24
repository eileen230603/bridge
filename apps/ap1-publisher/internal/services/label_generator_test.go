package services

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/local/dicom-disc-suite/shared/models"
)

func TestGenerateDiscLabelCreatesRecommendedSizePNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "label.png")
	study := models.Study{PatientName: "Ana Torres", PatientID: "P-1003", StudyDescription: "Abdomen", Modality: "CT", StudyDate: "2026-08-19"}
	if err := GenerateDiscLabel(path, study); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != labelSize || img.Bounds().Dy() != labelSize {
		t.Fatalf("unexpected label size: %v", img.Bounds())
	}
}
