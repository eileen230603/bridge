package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveStudyPathUsesExplicitContentDirectory(t *testing.T) {
	content := filepath.Join(t.TempDir(), "portable-study")
	manifest, data := resolveStudyPath(content, filepath.Join("ignored", "viewer.exe"), "ignored")
	assertPaths(t, content, manifest, data)
}

func TestResolveStudyPathUsesStudyBesideExecutable(t *testing.T) {
	root := t.TempDir()
	createStudyDir(t, root, time.Now())
	manifest, data := resolveStudyPath("", filepath.Join(root, "Portable DICOM Viewer.exe"), filepath.Join(root, "elsewhere"))
	assertPaths(t, root, manifest, data)
}

func TestResolveStudyPathFindsNewestDevelopmentStudy(t *testing.T) {
	repo := t.TempDir()
	older := filepath.Join(repo, "runtime", "temp", "older")
	newer := filepath.Join(repo, "runtime", "temp", "newer")
	createStudyDir(t, older, time.Now().Add(-time.Hour))
	createStudyDir(t, newer, time.Now())

	workDir := filepath.Join(repo, "apps", "ap2-viewer", "frontend")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest, data := resolveStudyPath("", "", workDir)
	assertPaths(t, newer, manifest, data)
}

func TestVisibleImageSeriesExcludesStructuredReports(t *testing.T) {
	series := []ViewerSeries{
		{ID: "mr", Modality: "MR", Files: []ViewerFile{{InstanceUID: "mr-1"}, {InstanceUID: "mr-2"}}},
		{ID: "sr", Modality: "SR", Files: []ViewerFile{{InstanceUID: "sr-1"}}},
		{ID: "ct", Modality: "CT", Files: []ViewerFile{{InstanceUID: "ct-1"}}},
	}

	visible, imageCount := visibleImageSeries(series)
	if len(visible) != 2 || visible[0].ID != "mr" || visible[1].ID != "ct" {
		t.Fatalf("unexpected visible series: %#v", visible)
	}
	if imageCount != 3 {
		t.Fatalf("expected 3 visible images, got %d", imageCount)
	}
}

func TestVisibleImageSeriesRecognizesNormalizedSRModality(t *testing.T) {
	visible, imageCount := visibleImageSeries([]ViewerSeries{{ID: "sr", Modality: " sr ", Files: []ViewerFile{{InstanceUID: "sr-1"}}}})
	if len(visible) != 0 || imageCount != 0 {
		t.Fatalf("SR must be completely hidden: series=%#v imageCount=%d", visible, imageCount)
	}
}

func createStudyDir(t *testing.T, dir string, modified time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "study.dat")
	if err := os.WriteFile(manifest, []byte(`{"id":"test","series":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(manifest, modified, modified); err != nil {
		t.Fatal(err)
	}
}

func assertPaths(t *testing.T, root, manifest, data string) {
	t.Helper()
	if manifest != filepath.Join(root, "study.dat") || data != filepath.Join(root, "data") {
		t.Fatalf("unexpected paths: manifest=%q data=%q", manifest, data)
	}
}
