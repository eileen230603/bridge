package main

import (
	"context"
	"encoding/json"
	"github.com/local/dicom-disc-suite/apps/ap2-viewer/internal/environment"
	"github.com/local/dicom-disc-suite/shared/models"
	"os"
	"path/filepath"
)

type ViewerState struct {
	DataFound  bool                  `json:"dataFound"`
	DataPath   string                `json:"dataPath"`
	Manifest   *models.StudyManifest `json:"manifest,omitempty"`
	ImageCount int                   `json:"imageCount"`
	Error      string                `json:"error,omitempty"`
}
type App struct {
	ctx       context.Context
	validator environment.ExecutionEnvironmentValidator
}

func NewApp() *App                         { return &App{validator: environment.DevelopmentEnvironmentValidator{}} }
func (a *App) startup(ctx context.Context) { a.ctx = ctx }
func (a *App) LoadStudy() ViewerState {
	if e := a.validator.Validate(); e != nil {
		return ViewerState{Error: e.Error()}
	}
	exe, e := os.Executable()
	if e != nil {
		return ViewerState{Error: e.Error()}
	}
	base := filepath.Dir(exe)
	if dev := os.Getenv("DICOM_VIEWER_CONTENT_DIR"); dev != "" {
		base = dev
	}
	data := filepath.Join(base, "data")
	info, e := os.Stat(data)
	if e != nil || !info.IsDir() {
		return ViewerState{DataPath: data, DataFound: false}
	}
	state := ViewerState{DataFound: true, DataPath: data}
	entries, _ := os.ReadDir(data)
	for _, x := range entries {
		if !x.IsDir() {
			state.ImageCount++
		}
	}
	raw, e := os.ReadFile(filepath.Join(base, "study.json"))
	if e == nil {
		var m models.StudyManifest
		if json.Unmarshal(raw, &m) == nil {
			state.Manifest = &m
		}
	}
	return state
}
