package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/local/dicom-disc-suite/shared/models"
	_ "modernc.org/sqlite"
)

// JobHistory stores the latest snapshot of each job. Epson is never resubmitted on load.
type JobHistory struct{ db *sql.DB }

func OpenJobHistory(path string) (*JobHistory, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = FULL",
		"CREATE TABLE IF NOT EXISTS disc_jobs (id TEXT PRIMARY KEY, created_at TEXT NOT NULL, snapshot TEXT NOT NULL)",
	} {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &JobHistory{db: db}, nil
}

func (s *JobHistory) Save(job models.DiscJob) error {
	if job.ID == "" {
		return fmt.Errorf("el trabajo no tiene identificador")
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO disc_jobs(id,created_at,snapshot) VALUES(?,?,?)
 ON CONFLICT(id) DO UPDATE SET snapshot=excluded.snapshot`, job.ID, job.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000000Z"), string(payload))
	return err
}

func (s *JobHistory) Load() ([]models.DiscJob, error) {
	rows, err := s.db.Query("SELECT snapshot FROM disc_jobs ORDER BY created_at, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]models.DiscJob, 0)
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var job models.DiscJob
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			return nil, fmt.Errorf("historial inválido: %w", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (s *JobHistory) Close() error { return s.db.Close() }
