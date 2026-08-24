package adapters

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/local/dicom-disc-suite/shared/models"
)

type EpsonJobMonitor interface {
	GetStatus(context.Context, models.DiscJob) (EpsonJobState, error)
}

type EpsonJobState struct {
	Status          models.JobStatus
	Technical       string
	ErrorCode       string
	TechnicalStatus string
	DetailStatus    string
	ErrorMessage    string
}

// TdBridgeJobMonitor implements the minimum status mechanism documented by
// Epson: TD Bridge renames the submitted job as it advances.
type TdBridgeJobMonitor struct {
	MonitoringFolder string
}

func (m TdBridgeJobMonitor) GetStatus(ctx context.Context, job models.DiscJob) (EpsonJobState, error) {
	current := EpsonJobState{Status: job.Status, Technical: job.EpsonState, ErrorCode: job.ErrorCode, TechnicalStatus: job.TechnicalStatus, DetailStatus: job.DetailStatus, ErrorMessage: job.ErrorMessage}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return current, err
		}
	}
	base := strings.TrimSpace(job.ID)
	if base == "" || filepath.Base(base) != base {
		return current, errors.New("invalid Epson job ID")
	}
	for _, candidate := range []struct {
		extension string
		status    models.JobStatus
	}{
		{".ERR", models.Failed},
		{".DON", models.Completed},
		{".STP", models.Processing},
		{".INP", models.Processing},
		{".RJD", models.Processing},
		{".JDF", models.QueuedForEpson},
	} {
		path := filepath.Join(m.MonitoringFolder, base+candidate.extension)
		_, err := os.Stat(path)
		if err == nil {
			state := EpsonJobState{Status: candidate.status, Technical: strings.TrimPrefix(candidate.extension, ".")}
			if candidate.status == models.Failed {
				status, readErr := readMonitoringStatus(m.MonitoringFolder)
				if readErr == nil {
					if section, ok := status[base]; ok {
						state.ErrorCode = section["ERROR"]
						state.TechnicalStatus = section["STATUS"]
						state.DetailStatus = section["DETAIL_STATUS"]
					}
				}
				state.ErrorMessage = TdBridgeErrorMessage(state.ErrorCode, state.DetailStatus)
			}
			return state, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return current, err
		}
	}
	return current, nil
}

func readMonitoringStatus(folder string) (map[string]map[string]string, error) {
	status, err := readTDBStatus(filepath.Join(folder, "TDBStatus.txt"))
	if err == nil {
		return status, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return readTDBStatus(filepath.Join(folder, "TDBStatus"))
}

func TdBridgeErrorMessage(code, _ string) string {
	code = strings.TrimSpace(code)
	switch code {
	case "JDF0203":
		return "No hay una Epson Discproducer configurada."
	case "":
		return "TD Bridge reportó un error."
	default:
		return fmt.Sprintf("TD Bridge reportó un error (código: %s).", code)
	}
}

func readTDBStatus(path string) (map[string]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sections := make(map[string]map[string]string)
	var current map[string]string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := strings.TrimSpace(line[1 : len(line)-1])
			if name == "" {
				current = nil
				continue
			}
			current = make(map[string]string)
			sections[name] = current
			continue
		}
		if current == nil {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		current[strings.ToUpper(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sections, nil
}
