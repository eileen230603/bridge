package main

import (
	"context"
	"fmt"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/services"
	"github.com/local/dicom-disc-suite/shared/dicom"
	"github.com/local/dicom-disc-suite/shared/models"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type App struct {
	ctx       context.Context
	cfg       config.Config
	repo      dicom.StudyRepository
	publisher adapters.EpsonPublisher
	builder   *services.StudyPackageBuilder
	logger    *slog.Logger
	mu        sync.RWMutex
	studies   map[string]models.Study
	jobs      []models.DiscJob
}

func NewApp() (*App, error) {
	cfgPath := "config.json"
	if _, e := os.Stat(cfgPath); e != nil {
		cfgPath = filepath.Join("apps", "ap1-publisher", "config.json")
	}
	cfg, e := config.Load(cfgPath)
	if e != nil {
		return nil, e
	}
	os.MkdirAll("../../runtime/logs", 0755)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	repo := adapters.MockStudyRepository{}
	publisher, e := selectPublisher(cfg, logger)
	if e != nil {
		return nil, e
	}
	builder := &services.StudyPackageBuilder{Repository: repo, TempRoot: cfg.TemporaryDirectory, ViewerSource: filepath.Join("..", "ap2-viewer", "build", "bin", "Portable DICOM Viewer.exe"), Logger: logger}
	return &App{cfg: cfg, repo: repo, publisher: publisher, builder: builder, logger: logger, studies: map[string]models.Study{}}, nil
}

func selectPublisher(cfg config.Config, logger *slog.Logger) (adapters.EpsonPublisher, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Publisher)) {
	case "mock":
		return &adapters.MockEpsonPublisher{StagingDirectory: cfg.Epson.StagingDirectory, HotFolder: cfg.MockHotFolder, Logger: logger}, nil
	case "tdbridge":
		return &adapters.TdBridgePublisher{MonitoringFolder: cfg.Epson.MonitoringFolder, StagingDirectory: cfg.Epson.StagingDirectory, Logger: logger}, nil
	default:
		return nil, fmt.Errorf("unsupported publisher %q (expected mock or tdbridge)", cfg.Publisher)
	}
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("AP1 started")
	cleanup := services.TempCleanupService{Root: a.cfg.TemporaryDirectory, MaxAge: time.Duration(a.cfg.CleanupAfterHours) * time.Hour, Enabled: a.cfg.CleanupEnabled, Logger: a.logger}
	if e := cleanup.Scan(); e != nil {
		a.logger.Error("Cleanup scan failed", "error", e)
	}
}
func (a *App) SearchStudies(from, to string) ([]models.Study, error) {
	f, e := time.Parse("2006-01-02", from)
	if e != nil {
		return nil, e
	}
	t, e := time.Parse("2006-01-02", to)
	if e != nil {
		return nil, e
	}
	a.logger.Info("Searching studies", "from", from, "to", to)
	studies, e := a.repo.SearchStudies(a.ctx, f, t)
	if e == nil {
		a.mu.Lock()
		for _, s := range studies {
			a.studies[s.StudyInstanceUID] = s
		}
		a.mu.Unlock()
	}
	return studies, e
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
		return a.fail(job, e)
	}
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
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]models.DiscJob(nil), a.jobs...)
}
