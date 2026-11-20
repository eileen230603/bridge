package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/shared/models"
)

// Modo desarrollo (AP1_DEV=1): estudios dummy, licencia aceptada y estados de
// Epson simulados. Todo se escribe en runtime/dev, nunca en la config real.

func devModeEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("AP1_DEV")))
	return v == "1" || v == "true"
}

// devRuntimeDir devuelve <repo>/runtime/dev, subiendo desde el directorio de trabajo.
func devRuntimeDir(workDir string) string {
	for dir := workDir; ; dir = filepath.Dir(dir) {
		if info, err := os.Stat(filepath.Join(dir, "runtime")); err == nil && info.IsDir() {
			return filepath.Join(dir, "runtime", "dev")
		}
		if filepath.Dir(dir) == dir {
			return filepath.Join(workDir, "runtime", "dev")
		}
	}
}

func applyDevConfig(cfg *config.Config, dir string) {
	cfg.TemporaryDirectory = filepath.Join(dir, "temp")
	cfg.CompletedDirectory = filepath.Join(dir, "completed")
	cfg.LogFile = filepath.Join(dir, "ap1.log")
	cfg.HistoryFile = filepath.Join(dir, "history.sqlite")
	cfg.Epson.MonitoringFolder = filepath.Join(dir, "tdbridge-orders")
	cfg.Epson.StagingDirectory = filepath.Join(dir, "epson-staging")
	cfg.Epson.Enabled = true
	if cfg.StudyAPI.Host == "" {
		cfg.StudyAPI.Host = "dev.local" // ficticio; el repositorio dummy no usa red
	}
	_ = os.MkdirAll(cfg.Epson.MonitoringFolder, 0o755)
}

type studyRepository interface {
	SearchStudies(context.Context, time.Time, time.Time) ([]models.Study, error)
	RetrieveStudy(context.Context, string, string) error
	UpdateConfig(config.ServerConfig)
}

type devStudyRepository struct{}

func (devStudyRepository) UpdateConfig(config.ServerConfig) {}

func (devStudyRepository) SearchStudies(_ context.Context, from, to time.Time) ([]models.Study, error) {
	type seed struct {
		id, name, desc, mod, sex string
		series, inst             int
	}
	seeds := []seed{
		{"P-1001", "PEREZ^MARIA", "TORAX", "CT", "F", 3, 295},
		{"P-1002", "LOPEZ^JUAN", "CEREBRO", "MR", "M", 4, 182},
		{"P-1003", "TORRES^ANA", "ABDOMEN", "CT", "F", 2, 146},
		{"P-1004", "ROJAS^CARLOS", "RODILLA DERECHA", "MR", "M", 3, 96},
		{"P-1005", "VARGAS^ELENA", "TORAX PA (falla en Epson)", "CR", "F", 1, 2},
		{"P-1006", "MENDEZ^LUIS", "COLUMNA LUMBAR", "MR", "M", 5, 210},
	}
	var out []models.Study
	day := time.Now()
	for i, s := range seeds {
		d := day.AddDate(0, 0, -i/2)
		if d.Before(from.AddDate(0, 0, -1)) || d.After(to.AddDate(0, 0, 1)) {
			continue
		}
		uid := fmt.Sprintf("1.2.840.10008.dev.%d", 1000+i)
		st := models.Study{StudyID: 1000 + i, StudyInstanceUID: uid, PatientID: s.id, PatientName: s.name,
			PatientBirthDate: "19800101", PatientSex: s.sex, StudyDescription: s.desc,
			StudyDate: d.Format("2006-01-02"), Modality: s.mod, SeriesCount: s.series, InstanceCount: s.inst}
		for n := 1; n <= s.series; n++ {
			st.Series = append(st.Series, models.Series{SeriesID: fmt.Sprint(n), SeriesUID: fmt.Sprintf("%s.%d", uid, n),
				SeriesInstanceUID: fmt.Sprintf("%s.%d", uid, n), SeriesNumber: n, SeriesDescription: fmt.Sprintf("SERIE %d", n),
				Modality: s.mod, InstanceCount: s.inst / s.series})
		}
		out = append(out, st)
	}
	return out, nil
}

func (devStudyRepository) RetrieveStudy(ctx context.Context, uid, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	select {
	case <-time.After(1500 * time.Millisecond): // simula la descarga
	case <-ctx.Done():
		return ctx.Err()
	}
	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf("PLACEHOLDER - NOT A DICOM FILE\nStudyInstanceUID=%s\n", uid)
		if err := os.WriteFile(filepath.Join(dest, fmt.Sprintf("dev%03d.dcm", i)), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// devJobMonitor avanza el estado según la edad del trabajo. Los pacientes
// P-1005 terminan en error para poder ver ese caso en la UI.
type devJobMonitor struct{}

func (devJobMonitor) GetStatus(_ context.Context, job models.DiscJob) (adapters.EpsonJobState, error) {
	age := time.Since(job.CreatedAt)
	switch {
	case job.PatientID == "P-1005" && age > 12*time.Second:
		return adapters.EpsonJobState{Status: models.Failed, Technical: "Error", ErrorCode: "E-DEV", ErrorMessage: "Error simulado de TD Bridge (modo dev)."}, nil
	case age > 25*time.Second:
		return adapters.EpsonJobState{Status: models.Completed, Technical: "Completed"}, nil
	case age > 12*time.Second:
		return adapters.EpsonJobState{Status: models.Publishing, Technical: "Burning"}, nil
	case age > 6*time.Second:
		return adapters.EpsonJobState{Status: models.Processing, Technical: "Processing"}, nil
	}
	return adapters.EpsonJobState{Status: models.QueuedForEpson, Technical: "Queued"}, nil
}
