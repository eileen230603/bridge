// app.go
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

func (a *App) resolveStudyPath() (string, string) {
	exe, _ := os.Executable()
	workDir, _ := os.Getwd()
	return resolveStudyPath(os.Getenv("DICOM_VIEWER_CONTENT_DIR"), exe, workDir)
}

// resolveStudyPath keeps the viewer portable: it never depends on a user- or
// operating-system-specific absolute path.
func resolveStudyPath(contentDir, executable, workDir string) (string, string) {
	if contentDir != "" {
		return studyPaths(contentDir)
	}

	// On a published disc, study.json and data live beside the viewer executable.
	if executable != "" {
		exeDir := filepath.Dir(executable)
		if isStudyDir(exeDir) {
			return studyPaths(exeDir)
		}
	}

	// During development, locate runtime/temp by walking up from both the
	// working directory and executable directory. This works on Windows,
	// macOS and Linux, regardless of where the repository was cloned.
	starts := []string{workDir}
	if executable != "" {
		starts = append(starts, filepath.Dir(executable))
	}
	for _, start := range starts {
		for dir := filepath.Clean(start); dir != ""; dir = filepath.Dir(dir) {
			if newest := newestStudyDir(filepath.Join(dir, "runtime", "temp")); newest != "" {
				return studyPaths(newest)
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}

	// Preserve a useful error path when no study exists.
	base := workDir
	if executable != "" {
		base = filepath.Dir(executable)
	}
	return studyPaths(base)
}

func studyPaths(dir string) (string, string) {
	return filepath.Join(dir, "study.json"), filepath.Join(dir, "data")
}

func isStudyDir(dir string) bool {
	manifest, data := studyPaths(dir)
	manifestInfo, manifestErr := os.Stat(manifest)
	dataInfo, dataErr := os.Stat(data)
	return manifestErr == nil && !manifestInfo.IsDir() && dataErr == nil && dataInfo.IsDir()
}

func newestStudyDir(tempDir string) string {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return ""
	}

	var newest string
	var newestTime time.Time
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(tempDir, entry.Name())
		if !isStudyDir(candidate) {
			continue
		}
		info, err := os.Stat(filepath.Join(candidate, "study.json"))
		if err == nil && (newest == "" || info.ModTime().After(newestTime)) {
			newest = candidate
			newestTime = info.ModTime()
		}
	}
	return newest
}

func (a *App) LoadStudy() ViewerState {

	if e := a.validator.Validate(); e != nil {
		return ViewerState{Error: e.Error()}
	}

	jsonPath, dataPath := a.resolveStudyPath()
	a.activeDataDir = dataPath

	info, e := os.Stat(dataPath)
	if e != nil || !info.IsDir() {
		return ViewerState{DataPath: dataPath, DataFound: false, Error: fmt.Sprintf("no se encontró el directorio de imágenes DICOM: %s", dataPath)}
	}

	state := ViewerState{DataFound: true, DataPath: dataPath}
	entries, _ := os.ReadDir(dataPath)
	for _, x := range entries {
		if !x.IsDir() {
			state.ImageCount++
		}
	}

	raw, e := os.ReadFile(jsonPath)
	if e != nil {
		state.Error = fmt.Sprintf("no se pudo leer el manifiesto del estudio %s: %v", jsonPath, e)
		return state
	}

	var m ViewerStudy
	if e = json.Unmarshal(raw, &m); e != nil {
		state.Error = fmt.Sprintf("el manifiesto del estudio no es válido: %v", e)
		return state
	}
	state.Manifest = &m

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
