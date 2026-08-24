package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/local/dicom-disc-suite/apps/ap2-viewer/internal/environment"
	"github.com/local/dicom-disc-suite/shared/models"
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

func NewApp(validator environment.ExecutionEnvironmentValidator) *App {
    return &App{validator: validator}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// getBaseDir obtiene el directorio donde reside el ejecutable o la variable de entorno de dev
func (a *App) getBaseDir() string {
    if dev := os.Getenv("DICOM_VIEWER_CONTENT_DIR"); dev != "" {
        return dev
    }
    exe, err := os.Executable()
    if err != nil {
        return "."
    }
    return filepath.Dir(exe)
}

func (a *App) LoadStudy() ViewerState {
    if e := a.validator.Validate(); e != nil {
        return ViewerState{Error: e.Error()}
    }

    base := a.getBaseDir()
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

    raw, e := os.ReadFile(filepath.Join(base, "/Users/medicaresoft/Desktop/Practicas/dicomdisc/bridge/runtime/temp/12374e42-3ffcfa9d-a4ba9ed7-713338cf-91608256/study.json"))
    if e == nil {
        var m models.StudyManifest
        if json.Unmarshal(raw, &m) == nil {
            state.Manifest = &m
        }
    }

    return state
}

// GetDicomFile lee un archivo .dcm de la carpeta data y lo retorna en Base64
func (a *App) GetDicomFile(filename string) (string, error) {
    base := a.getBaseDir()
    filePath := filepath.Join(base, "data", filename)

    bytes, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    return base64.StdEncoding.EncodeToString(bytes), nil
}