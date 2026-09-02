package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type ServerConfig struct {
	Protocol          string `json:"protocol"`
	Host              string `json:"host"`
	Port              int    `json:"port"`
	TimeoutSeconds    int    `json:"timeoutSeconds"`
	BaseURL           string `json:"baseUrl,omitempty"`
	GetStudiesPath    string `json:"getStudiesPath,omitempty"`
	DownloadStudyPath string `json:"downloadStudyPath,omitempty"`
}

type StudyAPIConfig = ServerConfig

func (c ServerConfig) IsConfigured() bool { return c.Validate() == nil }

func (c ServerConfig) Validate() error {
	if c.Protocol == "" && c.Host == "" && c.BaseURL != "" {
		legacy := c
		migrateLegacyServerConfig(&legacy)
		return legacy.Validate()
	}
	if strings.TrimSpace(c.Host) == "" {
		return errors.New("Ingrese un servidor válido.")
	}
	if c.Protocol != "http" && c.Protocol != "https" {
		return errors.New("Protocolo inválido.")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("Puerto inválido.")
	}
	if c.TimeoutSeconds <= 0 {
		return errors.New("El timeout debe ser mayor a 0.")
	}
	if strings.ContainsAny(c.Host, "/?#") || (strings.Contains(c.Host, ":") && net.ParseIP(c.Host) == nil) {
		return errors.New("Ingrese un servidor válido.")
	}
	return nil
}

func (c ServerConfig) BaseAddress() (string, error) {
	if c.Protocol == "" && c.Host == "" && c.BaseURL != "" {
		migrateLegacyServerConfig(&c)
	}
	if err := c.Validate(); err != nil {
		return "", err
	}
	host := strings.TrimSpace(c.Host)
	if ip := net.ParseIP(host); ip != nil && strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return fmt.Sprintf("%s://%s:%d", c.Protocol, host, c.Port), nil
}

type EpsonConfig struct {
	MonitoringFolder string `json:"monitoringFolder"`
	StagingDirectory string `json:"stagingDirectory"`
	Enabled          bool   `json:"enabled"`
	DefaultCopies    int    `json:"defaultCopies"`
}

func Load(path string) (Config, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return Config{}, e
	}
	return LoadBytes(b, filepath.Dir(path))
}

// LoadBytes parses an embedded configuration and resolves its relative paths
// against baseDir in the same way Load resolves paths beside a config file.
func LoadBytes(b []byte, baseDir string) (Config, error) {
	var c Config
	e := json.Unmarshal(b, &c)
	if e != nil {
		return c, e
	}
	c.TemporaryDirectory = resolve(baseDir, c.TemporaryDirectory)
	c.CompletedDirectory = resolve(baseDir, c.CompletedDirectory)
	c.LogFile = resolveOptional(baseDir, c.LogFile)
	c.Epson.MonitoringFolder = resolveOptional(baseDir, c.Epson.MonitoringFolder)
	c.Epson.StagingDirectory = resolveOptional(baseDir, c.Epson.StagingDirectory)
	if c.Epson.DefaultCopies == 0 {
		c.Epson.DefaultCopies = 1
	}
	if c.StudyAPI.TimeoutSeconds <= 0 {
		c.StudyAPI.TimeoutSeconds = 15
	}
	migrateLegacyServerConfig(&c.StudyAPI)
	return c, nil
}

func Save(path string, c Config) error {
	c.StudyAPI.BaseURL, c.StudyAPI.GetStudiesPath, c.StudyAPI.DownloadStudyPath = "", "", ""
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func migrateLegacyServerConfig(c *ServerConfig) {
	if c.Protocol != "" || c.Host != "" || c.Port != 0 || c.BaseURL == "" {
		return
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return
	}
	c.Protocol, c.Host = u.Scheme, u.Hostname()
	if port, err := strconv.Atoi(u.Port()); err == nil {
		c.Port = port
	}
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
