package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	TemporaryDirectory string      `json:"temporaryDirectory"`
	CompletedDirectory string      `json:"completedDirectory"`
	Publisher          string      `json:"publisher"`
	MockHotFolder      string      `json:"mockHotFolder"`
	Epson              EpsonConfig `json:"epson"`
	CleanupAfterHours  int         `json:"cleanupAfterHours"`
	CleanupEnabled     bool        `json:"cleanupEnabled"`
}

type EpsonConfig struct {
	MonitoringFolder string `json:"monitoringFolder"`
	StagingDirectory string `json:"stagingDirectory"`
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
	c.MockHotFolder = resolveOptional(base, c.MockHotFolder)
	c.Epson.MonitoringFolder = resolveOptional(base, c.Epson.MonitoringFolder)
	c.Epson.StagingDirectory = resolveOptional(base, c.Epson.StagingDirectory)
	if c.Publisher == "" {
		c.Publisher = "mock"
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
