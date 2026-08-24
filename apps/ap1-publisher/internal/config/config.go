package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	TemporaryDirectory string         `json:"temporaryDirectory"`
	CompletedDirectory string         `json:"completedDirectory"`
	LogFile            string         `json:"logFile"`
	StudyAPI           StudyAPIConfig `json:"studyApi"`
	Epson              EpsonConfig    `json:"epson"`
	CleanupAfterHours  int            `json:"cleanupAfterHours"`
	CleanupEnabled     bool           `json:"cleanupEnabled"`
}

type StudyAPIConfig struct {
	BaseURL           string `json:"baseUrl"`
	GetStudiesPath    string `json:"getStudiesPath"`
	DownloadStudyPath string `json:"downloadStudyPath"`
	TimeoutSeconds    int    `json:"timeoutSeconds"`
}

func (c StudyAPIConfig) IsConfigured() bool {
	return c.BaseURL != "" && c.GetStudiesPath != "" && c.DownloadStudyPath != ""
}

type EpsonConfig struct {
	MonitoringFolder string `json:"monitoringFolder"`
	StagingDirectory string `json:"stagingDirectory"`
	Enabled          bool   `json:"enabled"`
	DefaultCopies    int    `json:"defaultCopies"`
}

func Load(path string) (Config, error) {
	var c Config
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	e = json.Unmarshal(b, &c)
	if e != nil {
		return c, e
	}
	base := filepath.Dir(path)
	c.TemporaryDirectory = resolve(base, c.TemporaryDirectory)
	c.CompletedDirectory = resolve(base, c.CompletedDirectory)
	c.LogFile = resolveOptional(base, c.LogFile)
	c.Epson.MonitoringFolder = resolveOptional(base, c.Epson.MonitoringFolder)
	c.Epson.StagingDirectory = resolveOptional(base, c.Epson.StagingDirectory)
	if c.Epson.DefaultCopies == 0 {
		c.Epson.DefaultCopies = 1
	}
	if c.StudyAPI.TimeoutSeconds <= 0 {
		c.StudyAPI.TimeoutSeconds = 15
	}
	return c, nil
}
func resolveOptional(base, p string) string {
	if p == "" {
		return ""
	}
	return resolve(base, p)
}
func resolve(base, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(base, p))
}
