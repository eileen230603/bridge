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
	if _, _, _, alpha := img.At(0, 0).RGBA(); alpha != 0 {
		t.Fatal("outside of the physical disc must be transparent")
	}
	if _, _, _, alpha := img.At(labelSize/2, labelSize/2).RGBA(); alpha != 0 {
		t.Fatal("center hole must be transparent")
	}
	if _, _, _, alpha := img.At(labelSize/2, 100).RGBA(); alpha == 0 {
		t.Fatal("printable disc surface must be opaque")
	}
}
