package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/services"
	"github.com/local/dicom-disc-suite/shared/dicom"
	"github.com/local/dicom-disc-suite/shared/models"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type App struct {
	ctx              context.Context
	cfg              config.Config
	repo             dicom.StudyRepository
	studyRepo        *adapters.HttpStudyRepository
	publisher        adapters.EpsonPublisher
	monitor          adapters.EpsonJobMonitor
	builder          *services.StudyPackageBuilder
	logger           *slog.Logger
	mu               sync.RWMutex
	studies          map[string]models.Study
	jobs             []models.DiscJob
	pacsState        string
	studyServerState string
}

type SystemStatus struct {
	StudyServer        string `json:"studyServer"`
	StudyAPIConfigured bool   `json:"studyApiConfigured"`
	PACS               string `json:"pacs"`
	PACSConfigured     bool   `json:"pacsConfigured"`
	TDbridge           string `json:"tdBridge"`
	TDbridgeConfigured bool   `json:"tdBridgeConfigured"`
}

func NewApp() (*App, error) {
	cfgPath := strings.TrimSpace(os.Getenv("AP1_CONFIG"))
	if cfgPath == "" {
		cfgPath = "config.json"
		if _, e := os.Stat(cfgPath); e != nil {
			cfgPath = filepath.Join("apps", "ap1-publisher", "config.json")
		}
	}
	cfg, e := config.Load(cfgPath)
	if e != nil {
		return nil, e
	}
	logOutput := io.Writer(os.Stdout)
	if cfg.LogFile != "" {
		if e := os.MkdirAll(filepath.Dir(cfg.LogFile), 0o755); e != nil {
			return nil, fmt.Errorf("create log directory: %w", e)
		}
		logFile, e := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if e != nil {
			return nil, fmt.Errorf("open log file: %w", e)
		}
		logOutput = io.MultiWriter(os.Stdout, logFile)
	}
	logger := slog.New(slog.NewJSONHandler(logOutput, nil))
	repo := adapters.NewPacsStudyRepository(cfg.PACS, logger)
	studyRepo := adapters.NewHttpStudyRepository(cfg.StudyAPI)
	if !cfg.Epson.Enabled {
		return nil, fmt.Errorf("TD Bridge está deshabilitado en la configuración")
	}
	publisher := &adapters.TdBridgePublisher{MonitoringFolder: cfg.Epson.MonitoringFolder, StagingDirectory: cfg.Epson.StagingDirectory, DefaultCopies: cfg.Epson.DefaultCopies, Logger: logger}
	builder := &services.StudyPackageBuilder{Repository: studyRepo, TempRoot: cfg.TemporaryDirectory, ViewerSource: filepath.Join("..", "ap2-viewer", "build", "bin", "Portable DICOM Viewer.exe"), Logger: logger}
	monitor := adapters.TdBridgeJobMonitor{MonitoringFolder: cfg.Epson.MonitoringFolder}
	return &App{cfg: cfg, repo: repo, studyRepo: studyRepo, publisher: publisher, monitor: monitor, builder: builder, logger: logger, studies: map[string]models.Study{}, pacsState: "No probado", studyServerState: "No probado"}, nil
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("AP1 started")
	if repo, ok := a.repo.(*adapters.PacsStudyRepository); ok {
		// The Storage SCP owns its lifetime: it starts with AP1 and is stopped
		// explicitly by shutdown. Do not bind it to the startup callback context.
		if e := repo.Start(context.Background()); e != nil {
			a.mu.Lock()
			a.pacsState = "Error"
			a.mu.Unlock()
			a.logger.Error("Storage SCP failed", "error", e)
		}
	}
	cleanup := services.TempCleanupService{Root: a.cfg.TemporaryDirectory, MaxAge: time.Duration(a.cfg.CleanupAfterHours) * time.Hour, Enabled: a.cfg.CleanupEnabled, Logger: a.logger}
	if e := cleanup.Scan(); e != nil {
		a.logger.Error("Cleanup scan failed", "error", e)
	}
}

func (a *App) shutdown(_ context.Context) {
	if repo, ok := a.repo.(*adapters.PacsStudyRepository); ok {
		_ = repo.Close()
	}
}

func (a *App) TestPacsConnection() error {
	if !a.cfg.PACS.IsConfigured() {
		return errors.New("PACS no configurado")
	}
	if e := a.repo.Echo(a.ctx); e != nil {
		a.mu.Lock()
		a.pacsState = "Error"
		a.mu.Unlock()
		a.logger.Error("PACS C-ECHO failed", "error", e)
		return errors.New("No se pudo conectar al PACS.")
	}
	a.mu.Lock()
	a.pacsState = "Conectado"
	a.mu.Unlock()
	return nil
}

func (a *App) GetSystemStatus() SystemStatus {
	a.mu.RLock()
	pacsState := a.pacsState
	studyServerState := a.studyServerState
	a.mu.RUnlock()
	tdConfigured := false
	if a.cfg.Epson.Enabled {
		if info, err := os.Stat(a.cfg.Epson.MonitoringFolder); err == nil && info.IsDir() {
			tdConfigured = true
		}
	}
	tdState := "Error"
	if tdConfigured {
		tdState = "Configurado"
	}
	return SystemStatus{StudyServer: studyServerState, StudyAPIConfigured: a.cfg.StudyAPI.IsConfigured(), PACS: pacsState, PACSConfigured: a.cfg.PACS.IsConfigured(), TDbridge: tdState, TDbridgeConfigured: tdConfigured}
}
func (a *App) SearchStudies(from, to string) ([]models.Study, error) {
	a.mu.Lock()
	a.studies = make(map[string]models.Study)
	a.mu.Unlock()
	f, e := time.Parse("2006-01-02", from)
	if e != nil {
		return nil, e
	}
	t, e := time.Parse("2006-01-02", to)
	if e != nil {
		return nil, e
	}
	if t.Before(f) {
		return nil, errors.New("La fecha Hasta no puede ser anterior a Desde.")
	}
	a.logger.Info("Searching studies", "from", from, "to", to)
	if !a.cfg.StudyAPI.IsConfigured() {
		return nil, errors.New("Servidor de estudios no configurado")
	}
	studies, e := a.studyRepo.SearchStudies(a.ctx, f, t)
	if e != nil {
		a.mu.Lock()
		a.studyServerState = "Error"
		a.mu.Unlock()
		a.logger.Error("HTTP study search failed", "error", e)
		return nil, e
	}
	a.mu.Lock()
	a.studyServerState = "Conectado"
	a.studies = make(map[string]models.Study, len(studies))
	for _, s := range studies {
		if s.StudyInstanceUID != "" {
			a.studies[s.StudyInstanceUID] = s
		}
	}
	a.mu.Unlock()
	return studies, nil
}
func (a *App) CreateDiscJob(uid string) (models.DiscJob, error) {
	a.mu.RLock()
	study, ok := a.studies[uid]
	a.mu.RUnlock()
	if !ok {
		return models.DiscJob{}, fmt.Errorf("study not found; run search first")
	}
	a.logger.Info("Study selected", "study_uid", uid)
	job, e := a.builder.Build(a.ctx, study)
	if e != nil {
		if errors.Is(e, adapters.ErrStudyDownload) {
			e = adapters.ErrStudyDownload
		} else if !errors.Is(e, adapters.ErrInvalidStudyZIP) && !errors.Is(e, adapters.ErrNoDICOMFiles) {
			e = errors.New("No se pudo preparar el paquete del CD.")
		}
		return a.fail(job, e)
	}
	job.Status = models.Publishing
	job.UpdatedAt = time.Now()
	path, e := a.publisher.CreateJob(a.ctx, job)
	if e != nil {
		return a.fail(job, e)
	}
	job.EpsonJobPath = path
	if e = a.publisher.SubmitJob(a.ctx, path); e != nil {
		return a.fail(job, e)
	}
	job.Status = models.QueuedForEpson
	job.UpdatedAt = time.Now()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = job.UpdatedAt
	}
	a.mu.Lock()
	a.jobs = append(a.jobs, job)
	a.mu.Unlock()
	return job, nil
}
func (a *App) fail(j models.DiscJob, e error) (models.DiscJob, error) {
	j.Status = models.Failed
	j.ErrorMessage = e.Error()
	j.UpdatedAt = time.Now()
	a.logger.Error("Disc job failed", "job_id", j.ID, "error", e)
	a.mu.Lock()
	a.jobs = append(a.jobs, j)
	a.mu.Unlock()
	return j, e
}
func (a *App) ListJobs() []models.DiscJob {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.monitor != nil {
		for i := range a.jobs {
			status, err := a.monitor.GetStatus(a.ctx, a.jobs[i])
			if err != nil {
				a.logger.Error("Epson status check failed", "job_id", a.jobs[i].ID, "error", err)
				continue
			}
			if status != a.jobs[i].Status {
				a.jobs[i].Status = status
				a.jobs[i].UpdatedAt = time.Now()
			}
		}
	}
	return append([]models.DiscJob(nil), a.jobs...)
}
