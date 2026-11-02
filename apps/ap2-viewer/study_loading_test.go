package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap2-viewer/internal/environment"
)

func writeViewerFixture(t *testing.T, root string, study ViewerStudy) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(study)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := aes.NewCipher(secretKey)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	encrypted := gcm.Seal(nonce, nonce, payload, nil)
	if err := os.WriteFile(filepath.Join(root, "study.dat"), encrypted, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadedStudyRemainsPinnedAndFilesAreOrdered(t *testing.T) {
	root := t.TempDir()
	study := ViewerStudy{Series: []ViewerSeries{{ID: "ct", Modality: "CT", Files: []ViewerFile{{Position: 2, InstanceUID: "second"}, {Position: 1, InstanceUID: "first"}}}}}
	writeViewerFixture(t, root, study)
	if err := os.WriteFile(filepath.Join(root, "data", "first.dcm"), []byte("first-study"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DICOM_VIEWER_CONTENT_DIR", root)
	app := NewApp(environment.DevelopmentEnvironmentValidator{})
	state := app.LoadStudy()
	if state.Error != "" || state.ImageCount != 2 || state.Manifest.Series[0].Files[0].InstanceUID != "first" {
		t.Fatalf("unexpected state: %+v", state)
	}
	// Changing the discovery source must never mix another study's images into the open study.
	other := t.TempDir()
	writeViewerFixture(t, other, study)
	if err := os.WriteFile(filepath.Join(other, "data", "first.dcm"), []byte("other-study"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DICOM_VIEWER_CONTENT_DIR", other)
	got, err := app.GetDicomFile("first")
	if err != nil || got != base64.StdEncoding.EncodeToString([]byte("first-study")) {
		t.Fatalf("study source changed: %q %v", got, err)
	}
	for _, name := range []string{"", ".", "..", "../study.dat", `..\study.dat`, "C:study.dat"} {
		if _, err := app.GetDicomFile(name); err == nil {
			t.Fatalf("accepted invalid name %q", name)
		}
	}
	if _, err := app.GetDicomFile("missing"); err == nil {
		t.Fatal("missing image accepted")
	}
}

func TestInvalidOrEmptyStudyCannotServeImages(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DICOM_VIEWER_CONTENT_DIR", root)
	app := NewApp(environment.DevelopmentEnvironmentValidator{})
	if _, err := app.GetDicomFile("image"); err == nil {
		t.Fatal("read before loading")
	}
	writeViewerFixture(t, root, ViewerStudy{Series: []ViewerSeries{{Modality: "SR", Files: []ViewerFile{{InstanceUID: "report"}}}}})
	if state := app.LoadStudy(); state.Error == "" {
		t.Fatal("empty image study accepted")
	}
	if err := os.WriteFile(filepath.Join(root, "study.dat"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if state := app.LoadStudy(); state.Error == "" {
		t.Fatal("corrupt manifest accepted")
	}
	if _, err := app.GetDicomFile("image"); err == nil {
		t.Fatal("read after invalid load")
	}
}

func TestReadOnlyPortableStudyBesideExecutable(t *testing.T) {
	root := t.TempDir()
	createStudyDir(t, root, time.Now())
	if err := os.Chmod(filepath.Join(root, "study.dat"), 0444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "study.dat"), 0644) })
	manifest, data := resolveStudyPath("", filepath.Join(root, "viewer.exe"), t.TempDir())
	assertPaths(t, root, manifest, data)
}
