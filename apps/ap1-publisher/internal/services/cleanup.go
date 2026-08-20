package services

import (
	"log/slog"
	"os"
	"time"
)

type TempCleanupService struct {
	Root    string
	MaxAge  time.Duration
	Enabled bool
	Logger  *slog.Logger
}

func (s TempCleanupService) Scan() error {
	if !s.Enabled {
		s.Logger.Info("Temporary cleanup disabled")
		return nil
	}
	entries, err := os.ReadDir(s.Root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	cutoff := time.Now().Add(-s.MaxAge)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, er := e.Info()
		if er == nil && info.ModTime().Before(cutoff) {
			s.Logger.Info("Cleanup dry run: directory would be removed", "directory", e.Name(), "modified_at", info.ModTime())
		}
	}
	return nil
}
