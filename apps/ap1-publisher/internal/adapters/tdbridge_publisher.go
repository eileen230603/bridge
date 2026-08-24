package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/local/dicom-disc-suite/shared/models"
)

// TdBridgePublisher targets TD Bridge, not a particular Discproducer model.
type TdBridgePublisher struct {
	MonitoringFolder string
	StagingDirectory string
	DefaultCopies    int
	Logger           *slog.Logger
}

// BuildJDF implements the data-CD subset documented in TD Bridge Technical
// Reference Guide E/F revision 22, pages 48-67. Each DATA source is mapped
// explicitly so label/ and AP1 internals are not recorded on the disc.
func BuildJDF(job models.DiscJob, requestedCopies ...int) ([]byte, error) {
	if err := validateStudyPackage(job); err != nil {
		return nil, err
	}
	if !validJobID(job.ID) {
		return nil, errors.New("TD Bridge JOB_ID must be 1-40 characters using only letters, numbers, '-' or '_'")
	}
	copies := 1
	if len(requestedCopies) > 0 {
		copies = requestedCopies[0]
	}
	if copies < 1 || copies > 1000 {
		return nil, errors.New("TD Bridge COPIES must be between 1 and 1000")
	}
	paths := []string{job.DataPath, job.ViewerPath, job.ManifestPath}
	for _, path := range paths {
		if strings.ContainsAny(path, "\r\n\t") {
			return nil, errors.New("TD Bridge paths cannot contain tabs or line breaks")
		}
	}
	var out strings.Builder
	fmt.Fprintf(&out, "JOB_ID=%s\r\n", job.ID)
	fmt.Fprintf(&out, "COPIES=%d\r\n", copies)
	out.WriteString("DISC_TYPE=CD\r\nFORMAT=JOLIET\r\n")
	fmt.Fprintf(&out, "DATA=%s\tdata\r\n", job.DataPath)
	fmt.Fprintf(&out, "DATA=%s\tAP2\r\n", job.ViewerPath)
	fmt.Fprintf(&out, "DATA=%s\tstudy.json\r\n", job.ManifestPath)
	if job.LabelPath != "" {
		if info, err := os.Stat(job.LabelPath); err != nil || info.IsDir() {
			return nil, errors.New("label.png does not exist")
		}
		if strings.ContainsAny(job.LabelPath, "\r\n\t") {
			return nil, errors.New("TD Bridge label path cannot contain tabs or line breaks")
		}
		fmt.Fprintf(&out, "LABEL=%s\r\n", job.LabelPath)
	}
	return []byte(out.String()), nil
}

func (p *TdBridgePublisher) CreateJob(ctx context.Context, job models.DiscJob) (string, error) {
	p.logger().Info("[EPSON] Creating TD Bridge job", "job_id", job.ID)
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := validateStudyPackage(job); err != nil {
		return "", err
	}
	if strings.TrimSpace(p.StagingDirectory) == "" {
		return "", errors.New("TD Bridge staging directory not configured")
	}
	if err := os.MkdirAll(p.StagingDirectory, 0o755); err != nil {
		return "", fmt.Errorf("create TD Bridge staging directory: %w", err)
	}
	copies := p.DefaultCopies
	if copies == 0 {
		copies = 1
	}
	raw, err := BuildJDF(job, copies)
	if err != nil {
		return "", err
	}
	path := filepath.Join(p.StagingDirectory, job.ID+".jdf")
	if err := writeSyncedFile(path, raw); err != nil {
		return "", fmt.Errorf("stage JDF: %w", err)
	}
	p.logger().Info("[EPSON] JDF staged")
	return path, nil
}

func validJobID(id string) bool {
	if len(id) == 0 || len(id) > 40 {
		return false
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func (p *TdBridgePublisher) SubmitJob(ctx context.Context, jobFile string) error {
	p.logger().Info("[EPSON] Submitting job to TD Bridge")
	p.logger().Info("[EPSON] Monitoring folder: " + p.MonitoringFolder)
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(p.MonitoringFolder) == "" {
		return errors.New("TD Bridge monitoring folder not configured")
	}
	info, err := os.Stat(p.MonitoringFolder)
	if errors.Is(err, os.ErrNotExist) || (err == nil && !info.IsDir()) {
		return errors.New("TD Bridge monitoring folder does not exist")
	}
	if err != nil {
		return fmt.Errorf("inspect TD Bridge monitoring folder: %w", err)
	}
	jobInfo, err := os.Stat(jobFile)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("JDF file does not exist")
	}
	if err != nil {
		return fmt.Errorf("inspect JDF file: %w", err)
	}
	if jobInfo.IsDir() || jobInfo.Size() == 0 {
		return errors.New("JDF file is empty")
	}
	final := filepath.Join(p.MonitoringFolder, filepath.Base(jobFile))
	if _, err := os.Stat(final); err == nil {
		return fmt.Errorf("TD Bridge job already exists: %s", final)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect TD Bridge destination: %w", err)
	}
	temporary := final + ".tmp"
	if err := copyFileSynced(jobFile, temporary); err != nil {
		return err
	}
	if err := os.Rename(temporary, final); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("publish JDF atomically: %w", err)
	}
	p.logger().Info("[EPSON] Job submitted")
	return nil
}

func validateStudyPackage(job models.DiscJob) error {
	if info, err := os.Stat(job.TempPath); err != nil || !info.IsDir() {
		return errors.New("study package does not exist")
	}
	for _, item := range []struct{ path, name string }{{job.DataPath, "data directory"}, {job.ViewerPath, "AP2 directory"}} {
		if info, err := os.Stat(item.path); err != nil || !info.IsDir() {
			return fmt.Errorf("%s does not exist", item.name)
		}
	}
	if info, err := os.Stat(job.ManifestPath); err != nil || info.IsDir() {
		return errors.New("study.json does not exist")
	}
	if job.LabelPath != "" {
		info, err := os.Stat(job.LabelPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect label: %w", err)
		}
		if err == nil && info.IsDir() {
			return errors.New("configured label path is not a file")
		}
	}
	root, err := filepath.Abs(job.TempPath)
	if err != nil {
		return err
	}
	for _, path := range []string{job.DataPath, job.ViewerPath, job.ManifestPath} {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, abs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return errors.New("study package content must be inside TempPath")
		}
	}
	return nil
}

func copyFileSynced(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open staged JDF: %w", err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create temporary JDF in monitoring folder: %w", err)
	}
	defer func() {
		_ = out.Close()
		if err != nil {
			_ = os.Remove(dst)
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy JDF: %w", err)
	}
	if err = out.Sync(); err != nil {
		return fmt.Errorf("sync JDF: %w", err)
	}
	if err = out.Close(); err != nil {
		return fmt.Errorf("close copied JDF: %w", err)
	}
	return nil
}

func writeSyncedFile(path string, data []byte) (err error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return f.Close()
}

func (p *TdBridgePublisher) logger() *slog.Logger {
	if p.Logger != nil {
		return p.Logger
	}
	return slog.Default()
}
