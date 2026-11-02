package main

import (
	"fmt"
	"github.com/local/dicom-disc-suite/shared/models"
)

func (a *App) recordJob(job models.DiscJob) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.history != nil {
		if err := a.history.Save(job); err != nil {
			a.historyError = "No se pudo guardar el historial. Compruebe espacio y permisos de la carpeta de datos."
			return fmt.Errorf("guardar historial: %w", err)
		}
		a.historyError = ""
	}
	for i := range a.jobs {
		if a.jobs[i].ID == job.ID {
			a.jobs[i] = job
			return nil
		}
	}
	a.jobs = append(a.jobs, job)
	return nil
}

func (a *App) GetHistoryError() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.historyError
}

func (a *App) closeHistory() {
	if a.history != nil {
		_ = a.history.Close()
	}
}
