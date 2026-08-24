// app.go
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/local/dicom-disc-suite/apps/ap2-viewer/internal/environment"
)

// Definición de estructuras locales para el manifiesto del visor
type ViewerFile struct {
	Position    int    `json:"position"`
	InstanceUID string `json:"instanceUid"`
}

type ViewerSeries struct {
	ID                string       `json:"id"`
	SeriesInstanceUID string       `json:"seriesInstanceUid"`
	Modality          string       `json:"modality"`
	Name              string       `json:"name"`
	Files             []ViewerFile `json:"files"`
}

type ViewerStudy struct {
	ID               string         `json:"id"`
	StudyInstanceUID string         `json:"studyInstanceUid"`
	PatientName      string         `json:"patientName"`
	PatientID        string         `json:"patientId"`
	PatientBirthDate *string        `json:"patientBirthDate"`
	PatientAge       *string        `json:"patientAge"`
	PatientSex       *string        `json:"patientSex"`
	StudyDate        *string        `json:"studyDate"`
	StudyTime        *string        `json:"studyTime"`
	StudyDescription *string        `json:"studyDescription"`
	Series           []ViewerSeries `json:"series"`
}
type ViewerState struct {
	DataFound  bool         `json:"dataFound"`
	DataPath   string       `json:"dataPath"`
	Manifest   *ViewerStudy `json:"manifest,omitempty"`
	ImageCount int          `json:"imageCount"`
	Error      string       `json:"error,omitempty"`
}
type App struct {
	ctx           context.Context
	validator     environment.ExecutionEnvironmentValidator
	activeDataDir string
}

func NewApp(validator environment.ExecutionEnvironmentValidator) *App {
	return &App{validator: validator}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

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

func (a *App) resolveStudyPath() (string, string) {
	// 1. Si enviamos la ruta manual por variable de entorno
	if dev := os.Getenv("DICOM_VIEWER_CONTENT_DIR"); dev != "" {
		return filepath.Join(dev, "study.json"), filepath.Join(dev, "data")
	}

	
	devTempDir := "/Users/medicaresoft/Desktop/Practicas/dicomdisc/bridge/runtime/temp"

	if info, err := os.Stat(devTempDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(devTempDir)
		if err == nil && len(entries) > 0 {
			var newestDir string
			var newestTime int64 = 0

			// Buscar la carpeta de sesión más reciente creada por AP1
			for _, entry := range entries {
				if entry.IsDir() {
					sessionPath := filepath.Join(devTempDir, entry.Name())
					candidateJson := filepath.Join(sessionPath, "study.json")

					if fileInfo, err := os.Stat(candidateJson); err == nil {
						if fileInfo.ModTime().UnixNano() > newestTime {
							newestTime = fileInfo.ModTime().UnixNano()
							newestDir = sessionPath
						}
					}
				}
			}

			if newestDir != "" {
				
				return filepath.Join(newestDir, "study.json"), filepath.Join(newestDir, "data")
			}
		}
	}

	// 3. Fallback en Producción (cuando AP2.exe corre grabado en el CD junto a study.json)
	base := a.getBaseDir()
	return filepath.Join(base, "study.json"), filepath.Join(base, "data")
}

func (a *App) LoadStudy() ViewerState {
   

	
	if e := a.validator.Validate(); e != nil {
		return ViewerState{Error: e.Error()}
	}

	jsonPath, dataPath := a.resolveStudyPath()
	a.activeDataDir = dataPath

	info, e := os.Stat(dataPath)
	if e != nil || !info.IsDir() {
		return ViewerState{DataPath: dataPath, DataFound: false}
	}

	state := ViewerState{DataFound: true, DataPath: dataPath}
	entries, _ := os.ReadDir(dataPath)
	for _, x := range entries {
		if !x.IsDir() {
			state.ImageCount++
		}
	}

	raw, e := os.ReadFile(jsonPath)
	if e == nil {
		
		var m ViewerStudy
		if json.Unmarshal(raw, &m) == nil {
			state.Manifest = &m
		}
	}

	return state
}

func (a *App) GetDicomFile(uidOrFilename string) (string, error) {
	_, dataDir := a.resolveStudyPath()

	cleanFilename := filepath.Base(uidOrFilename)

	// Buscar primero el UUID exacto tal cual viene en instanceUid (sin agregar .dcm)
	filePath := filepath.Join(dataDir, cleanFilename)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		// Si no lo encuentra, intentar agregándole .dcm por respaldo
		filePathWithExt := filePath + ".dcm"
		bytes, err = os.ReadFile(filePathWithExt)
		if err != nil {
			return "", fmt.Errorf("no se encontró el archivo DICOM: %s en %s", cleanFilename, dataDir)
		}
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}