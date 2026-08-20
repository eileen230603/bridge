package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	TemporaryDirectory string `json:"temporaryDirectory"`
	CompletedDirectory string `json:"completedDirectory"`
	EpsonHotFolder     string `json:"epsonHotFolder"`
	CleanupAfterHours  int    `json:"cleanupAfterHours"`
	CleanupEnabled     bool   `json:"cleanupEnabled"`
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
	c.EpsonHotFolder = resolve(base, c.EpsonHotFolder)
	return c, nil
}
func resolve(base, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(base, p))
}
