package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/services"
	"github.com/local/dicom-disc-suite/shared/models"
)

type App struct {
	ctx              context.Context
	cfg              config.Config
	configPath       string
	studyRepo        *adapters.HttpStudyRepository
	publisher        adapters.EpsonPublisher
	monitor          adapters.EpsonJobMonitor
	builder          *services.StudyPackageBuilder
	logger           *slog.Logger
	mu               sync.RWMutex
	studies          map[string]models.Study
	jobs             []models.DiscJob
	studyServerState string
	searchCancel     context.CancelFunc
	searchID         uint64
}

type SystemStatus struct {
	StudyServer        string `json:"studyServer"`
	StudyServerAddress string `json:"studyServerAddress"`
	StudyAPIConfigured bool   `json:"studyApiConfigured"`
	TDbridge           string `json:"tdBridge"`
	TDbridgeConfigured bool   `json:"tdBridgeConfigured"`
}

type ConnectionTestResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewApp() (*App, error) {
	executable, _ := os.Executable()
	workDir, _ := os.Getwd()
	cfgPath, found := resolveAP1ConfigPath(strings.TrimSpace(os.Getenv("AP1_CONFIG")), executable, workDir)
	var cfg config.Config
	var e error
	if found {
		cfg, e = config.Load(cfgPath)
	} else {
		// Keep the existing ../../runtime layout relative to the portable EXE.
		portableBase := filepath.Join(filepath.Dir(executable), "apps", "ap1-publisher")
		cfg, e = config.LoadBytes(defaultConfig, portableBase)
		cfgPath = "embedded config.json"
	}
	if e != nil {
		return nil, fmt.Errorf("load AP1 configuration %q: %w", cfgPath, e)
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
	studyRepo := adapters.NewHttpStudyRepository(cfg.StudyAPI)
	if !cfg.Epson.Enabled {
		return nil, fmt.Errorf("TD Bridge está deshabilitado en la configuración")
	}
	publisher := &adapters.TdBridgePublisher{MonitoringFolder: cfg.Epson.MonitoringFolder, StagingDirectory: cfg.Epson.StagingDirectory, DefaultCopies: cfg.Epson.DefaultCopies, Logger: logger}
	builder := &services.StudyPackageBuilder{Repository: studyRepo, TempRoot: cfg.TemporaryDirectory, ViewerBuilds: viewerBuilds, Logger: logger}
	monitor := adapters.TdBridgeJobMonitor{MonitoringFolder: cfg.Epson.MonitoringFolder}
	return &App{cfg: cfg, configPath: cfgPath, studyRepo: studyRepo, publisher: publisher, monitor: monitor, builder: builder, logger: logger, studies: map[string]models.Study{}, studyServerState: "No probado"}, nil
}

func resolveAP1ConfigPath(explicit, executable, workDir string) (string, bool) {
	if explicit != "" {
		return filepath.Clean(explicit), true
	}

	candidates := []string{filepath.Join(workDir, "config.json")}
	if executable != "" {
		executableDir := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(executableDir, "config.json"),
			filepath.Join(executableDir, "..", "..", "config.json"),
		)
	}
	candidates = append(candidates, filepath.Join(workDir, "apps", "ap1-publisher", "config.json"))
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Clean(candidate), true
		}
	}
	return "", false
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("AP1 started")
	cleanup := services.TempCleanupService{Root: a.cfg.TemporaryDirectory, MaxAge: time.Duration(a.cfg.CleanupAfterHours) * time.Hour, Enabled: a.cfg.CleanupEnabled, Logger: a.logger}
	if e := cleanup.Scan(); e != nil {
		a.logger.Error("Cleanup scan failed", "error", e)
	}
}

func (a *App) GetSystemStatus() SystemStatus {
	a.mu.RLock()
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
	a.mu.RLock()
	serverConfig := a.cfg.StudyAPI
	a.mu.RUnlock()
	address := ""
	if serverConfig.Host != "" {
		address = fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	}
	return SystemStatus{StudyServer: studyServerState, StudyServerAddress: address, StudyAPIConfigured: serverConfig.IsConfigured(), TDbridge: tdState, TDbridgeConfigured: tdConfigured}
}

func (a *App) GetServerConfig() config.ServerConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.StudyAPI
}

func (a *App) SaveServerConfig(server config.ServerConfig) error {
	server.Protocol = strings.ToLower(strings.TrimSpace(server.Protocol))
	server.Host = strings.TrimSpace(server.Host)
	if err := server.Validate(); err != nil {
		return err
	}
	a.mu.Lock()
	updated := a.cfg
	updated.StudyAPI = server
	if err := config.Save(a.configPath, updated); err != nil {
		a.mu.Unlock()
		return errors.New("No se pudo guardar la configuración.")
	}
	a.cfg = updated
	a.studyServerState = "No probado"
	a.mu.Unlock()
	a.studyRepo.UpdateConfig(server)
	return nil
}

func (a *App) TestServerConnection(server config.ServerConfig) (result ConnectionTestResult) {
	defer func() {
		a.mu.Lock()
		a.studyServerState = result.Status
		a.mu.Unlock()
	}()
	server.Protocol = strings.ToLower(strings.TrimSpace(server.Protocol))
	server.Host = strings.TrimSpace(server.Host)
	base, err := server.BaseAddress()
	if err != nil {
		return ConnectionTestResult{Status: "Error", Message: err.Error()}
	}
	today := time.Now()
	endpoint, _ := url.Parse(base + "/getestudios")
	query := endpoint.Query()
	query.Set("inicio", today.Format("20060102"))
	query.Set("final", today.AddDate(0, 0, 1).Format("20060102"))
	endpoint.RawQuery = query.Encode()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(server.TimeoutSeconds)*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	resp, err := (&http.Client{Timeout: time.Duration(server.TimeoutSeconds) * time.Second}).Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return ConnectionTestResult{Status: "Error", Message: "Tiempo de espera agotado."}
		}
		return ConnectionTestResult{Status: "Error", Message: "No se pudo conectar al servidor."}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ConnectionTestResult{Status: "Error", Message: "No se pudo conectar al servidor."}
	}
	return ConnectionTestResult{Status: "Conectado"}
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
	a.mu.Lock()
	if a.searchCancel != nil {
		a.searchCancel()
	}
	a.searchID++
	searchID := a.searchID
	ctx, cancel := context.WithCancel(a.ctx)
	a.searchCancel = cancel
	a.mu.Unlock()
	defer func() {
		cancel()
		a.mu.Lock()
		if a.searchID == searchID {
			a.searchCancel = nil
		}
		a.mu.Unlock()
	}()

	studies, e := a.studyRepo.SearchStudies(ctx, f, t)
	if e != nil {
		if errors.Is(e, context.Canceled) {
			return nil, errors.New("Búsqueda cancelada.")
		}
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

func (a *App) CancelSearch() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.searchCancel != nil {
		a.searchCancel()
	}
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
		a.logger.Error("Study package preparation failed", "study_uid", uid, "job_id", job.ID, "error", e)
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
			state, err := a.monitor.GetStatus(a.ctx, a.jobs[i])
			if err != nil {
				a.logger.Error("Epson status check failed", "job_id", a.jobs[i].ID, "error", err)
				continue
			}
			if state.Status != a.jobs[i].Status || state.Technical != a.jobs[i].EpsonState || state.ErrorCode != a.jobs[i].ErrorCode || state.TechnicalStatus != a.jobs[i].TechnicalStatus || state.DetailStatus != a.jobs[i].DetailStatus || state.ErrorMessage != a.jobs[i].ErrorMessage {
				previous := a.jobs[i].Status
				a.jobs[i].Status = state.Status
				a.jobs[i].EpsonState = state.Technical
				a.jobs[i].ErrorCode = state.ErrorCode
				a.jobs[i].TechnicalStatus = state.TechnicalStatus
				a.jobs[i].DetailStatus = state.DetailStatus
				a.jobs[i].ErrorMessage = state.ErrorMessage
				a.jobs[i].UpdatedAt = time.Now()
				a.logger.Info("[EPSON] Job state changed", "job_id", a.jobs[i].ID, "previous_status", previous, "new_status", state.Status, "technical_state", state.Technical)
				if state.Status == models.Failed {
					a.logger.Error("[EPSON] Job failed", "job_id", a.jobs[i].ID, "error_code", state.ErrorCode, "status", state.TechnicalStatus, "detail_status", state.DetailStatus)
				}
			}
		}
	}
	return append([]models.DiscJob(nil), a.jobs...)
}
