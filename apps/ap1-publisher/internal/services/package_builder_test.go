package services

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/local/dicom-disc-suite/shared/models"
)

type testStudyRetriever struct{}

func (testStudyRetriever) RetrieveStudy(_ context.Context, _ string, destination string) error {
	data := make([]byte, 132)
	copy(data[128:], "DICM")
	return os.WriteFile(filepath.Join(destination, "instance.dcm"), data, 0o644)
}

func TestStudyPackageBuilderCreatesDataAndRealManifest(t *testing.T) {
    study := models.Study{
        StudyID: 15491, StudyInstanceUID: "study-uuid", PatientName: "PEREZ MARIA",
        StudyDate: "2025-12-31", Modality: "MR", StudyDescription: "PELVIS", InstanceCount: 1,
        Series: []models.Series{{SeriesID: "15527", SeriesUID: "series-uuid", SeriesInstanceUID: "series-uuid", Modality: "MR", SeriesDescription: "T2", InstanceCount: 1, Files: []models.StudyFile{{Position: 7, InstanceUID: "instance-uuid", AE: "DEV1"}}}},
    }
    builder := StudyPackageBuilder{Repository: testStudyRetriever{}, TempRoot: t.TempDir(), Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
    job, err := builder.Build(context.Background(), study)
    if err != nil {
        t.Fatal(err)
    }
    if _, err := os.Stat(filepath.Join(job.DataPath, "instance.dcm")); err != nil {
        t.Fatal(err)
    }

    // 1. Leer el archivo binario cifrado (study.dat)
    rawEncrypted, err := os.ReadFile(job.ManifestPath)
    if err != nil {
        t.Fatal(err)
    }

    // 2. Descifrar en memoria antes de hacer el json.Unmarshal
    rawDecrypted, err := DecryptData(rawEncrypted, SecretKey)
    if err != nil {
        t.Fatalf("error al descifrar el manifest durante el test: %v", err)
    }

    var manifest ViewerStudy
    if err := json.Unmarshal(rawDecrypted, &manifest); err != nil {
        t.Fatal(err)
    }
    if manifest.ID != "15491" || manifest.StudyInstanceUID != "study-uuid" || len(manifest.Series) != 1 || len(manifest.Series[0].Files) != 1 || manifest.Series[0].Files[0].Position != 7 {
        t.Fatalf("unexpected manifest: %+v", manifest)
    }
}
