package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/local/dicom-disc-suite/shared/models"
)

type StudyPackageBuilder struct {
	Repository   StudyRetriever
	TempRoot     string
	ViewerBuilds fs.FS
	Logger       *slog.Logger
}

type StudyRetriever interface {
	RetrieveStudy(context.Context, string, string) error
}

func (b *StudyPackageBuilder) Build(ctx context.Context, study models.Study, labelBranding ...DiscLabelBranding) (models.DiscJob, error) {
	branding := DiscLabelBranding{}
	if len(labelBranding) > 0 {
		branding = labelBranding[0]
	}
	created := time.Now()
	now := models.DiscJob{
		ID:               fmt.Sprintf("job-%d", created.UnixNano()),
		StudyInstanceUID: study.StudyInstanceUID,
		PatientID:        study.PatientID,
		PatientName:      study.PatientName,
		StudyDescription: study.StudyDescription,
		Status:           models.Downloading,
		CreatedAt:        created,
		UpdatedAt:        created,
	}

	root := filepath.Join(b.TempRoot, study.StudyInstanceUID)
	now.TempPath = root
	now.DataPath = filepath.Join(root, "data")
	now.ViewerPath = root
	now.ManifestPath = filepath.Join(root, "study.dat")
	now.LabelPath = filepath.Join(root, "label", "label.png")

	b.Logger.Info("Preparing study package", "study_uid", study.StudyInstanceUID, "job_id", now.ID)

	if err := os.RemoveAll(root); err != nil {
		return now, err
	}

	for _, p := range []string{now.DataPath, filepath.Dir(now.LabelPath)} {
		if err := os.MkdirAll(p, 0755); err != nil {
			return now, err
		}
	}

	b.Logger.Info("Starting HTTP study download", "study_uid", study.StudyInstanceUID)
	if err := b.Repository.RetrieveStudy(ctx, study.StudyInstanceUID, now.DataPath); err != nil {
		return now, fmt.Errorf("download study: %w", err)
	}

	now.Status = models.Preparing

	if err := extractViewerBuilds(b.ViewerBuilds, root); err != nil {

		return now, fmt.Errorf("extract embedded viewer builds: %w", err)

	}

	// 1. Mapear la estructura del estudio
	manifest := MapSymphonyStudyToViewerStudy(study)

	// 2. Convertir a JSON
	jsonBytes, err := json.Marshal(manifest)
	if err != nil {
		return now, fmt.Errorf("marshal study manifest: %w", err)
	}

	// 3. Cifrar usando la función de crypto_utils.go
	encryptedBytes, err := EncryptData(jsonBytes, SecretKey)
	if err != nil {
		return now, fmt.Errorf("encrypt study manifest: %w", err)

	}

	// 4. Guardar como binario cifrado
	if err = os.WriteFile(now.ManifestPath, encryptedBytes, 0644); err != nil {
		return now, fmt.Errorf("write study.dat: %w", err)
	}
	b.Logger.Info("Encrypted study manifest created", "job_id", now.ID)

	// 5. Crear el archivo autorun.inf para la raíz del disco
    autorunContent := []byte("[autorun]\r\nopen=Symphony Viewer.exe\r\nicon=Symphony Viewer.exe\r\nlabel=Symphony Disc\r\n")
    autorunPath := filepath.Join(root, "autorun.inf")
    if err := os.WriteFile(autorunPath, autorunContent, 0644); err != nil {
        return now, fmt.Errorf("write autorun.inf: %w", err)
    }
    b.Logger.Info("autorun.inf created", "job_id", now.ID)
	// 6. Ocultar la ventana de ejecución del archivo autorun.inf
	hideFileWindows(autorunPath)

	if err = GenerateDiscLabel(now.LabelPath, study); err != nil {

	}

	// 4. Guardar como binario cifrado
	if err = os.WriteFile(now.ManifestPath, encryptedBytes, 0644); err != nil {
		return now, fmt.Errorf("write study.dat: %w", err)
	}
	b.Logger.Info("Encrypted study manifest created", "job_id", now.ID)

	if err = GenerateDiscLabelWithBranding(now.LabelPath, study, branding); err != nil {
		return now, fmt.Errorf("generate disc label: %w", err)
	}

	now.Status = models.Ready
	b.Logger.Info("Study package prepared", "job_id", now.ID)
	//launchViewer(root).  con esto hace que aparezca la ventana de Symphony Visor 
	return now, nil
}
// hideFileWindows aplica el atributo oculto al archivo si el sistema es Windows
// hideFileWindows aplica el atributo oculto al archivo si el sistema es Windows
func hideFileWindows(path string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("attrib", "+h", path)
		_ = cmd.Run()
	}
}

const (
	windowsViewer = "viewer-builds/windows/Symphony Viewer.exe"
	macOSViewer   = "viewer-builds/macos/Symphony Viewer.app"
)

func extractViewerBuilds(builds fs.FS, destination string) error {
	if builds == nil {
		return fmt.Errorf("embedded viewer filesystem is not configured")
	}
	if info, err := fs.Stat(builds, windowsViewer); err != nil || !info.Mode().IsRegular() {
		if err == nil {
			err = fmt.Errorf("not a regular file")
		}
		return fmt.Errorf("embedded Windows build %q is missing: %w", windowsViewer, err)
	}
	if info, err := fs.Stat(builds, macOSViewer); err != nil || !info.IsDir() {
		if err == nil {
			err = fmt.Errorf("not a directory")
		}
		return fmt.Errorf("embedded macOS build %q is missing: %w", macOSViewer, err)
	}

	return fs.WalkDir(builds, "viewer-builds", func(source string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if source == "viewer-builds" {
			return nil
		}

		// Normalizar la ruta a barras inclinadas (/)
		slashSource := filepath.ToSlash(source)

		// Determinar la ruta relativa removiendo la subcarpeta de origen (windows/ o macos/)
		var relative string
		if strings.HasPrefix(slashSource, "viewer-builds/windows/") {
			relative = strings.TrimPrefix(slashSource, "viewer-builds/windows/")
		} else if strings.HasPrefix(slashSource, "viewer-builds/macos/") {
			relative = strings.TrimPrefix(slashSource, "viewer-builds/macos/")
		} else {
			return nil
		}

		if relative == "" {
			return nil
		}

		// Construir la ruta de destino directa en la raíz
		target := filepath.Join(destination, filepath.FromSlash(relative))

		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if slashSource == "viewer-builds/windows/README.txt" || slashSource == "viewer-builds/macos/README.txt" {
			return nil
		}

		content, err := fs.ReadFile(builds, source)
		if err != nil {
			return fmt.Errorf("read embedded viewer file %q: %w", source, err)
		}

		// Asignar permisos 0755 al .exe de Windows y a los binarios ejecutables de Mac
		mode := fs.FileMode(0o644)
		if slashSource == windowsViewer || strings.Contains(slashSource, "Contents/MacOS/") {
			mode = 0o755
		}

		if err := os.WriteFile(target, content, mode); err != nil {
			return fmt.Errorf("write viewer file %q: %w", target, err)
		}
		return nil
	})
}
// launchViewer ejecuta automáticamente el visor del estudio al finalizar la preparación
func launchViewer(studyRoot string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		exePath := filepath.Join(studyRoot, "Symphony Viewer.exe")
		cmd = exec.Command(exePath)
	} else if runtime.GOOS == "darwin" {
		appPath := filepath.Join(studyRoot, "Symphony Viewer.app")
		cmd = exec.Command("open", appPath)
	} else {
		return
	}

	cmd.Dir = studyRoot
	// Start() inicia el proceso en segundo plano sin bloquear AP1
	_ = cmd.Start()
}


