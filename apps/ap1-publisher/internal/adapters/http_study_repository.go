package adapters

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/shared/models"
)

var (
	ErrStudyServerUnavailable = errors.New("No se pudo consultar el servidor de estudios.")
	ErrInvalidStudyResponse   = errors.New("La respuesta del servidor de estudios no tiene un formato válido.")
	ErrStudyDownload          = errors.New("No se pudo descargar el estudio.")
	ErrInvalidStudyZIP        = errors.New("El ZIP del estudio no es válido.")
	ErrNoDICOMFiles           = errors.New("El estudio descargado no contiene archivos DICOM.")
)

type HTTPStatusError struct{ StatusCode int }

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("Error consultando estudios: HTTP %d", e.StatusCode)
}

type HttpStudyRepository struct {
	config config.StudyAPIConfig
	client *http.Client
}

func NewHttpStudyRepository(cfg config.StudyAPIConfig) *HttpStudyRepository {
	return &HttpStudyRepository{config: cfg, client: &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second}}
}

func (r *HttpStudyRepository) SearchStudies(ctx context.Context, from, to time.Time) ([]models.Study, error) {
	requestURL, err := r.buildURL(from, to)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, ErrStudyServerUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, HTTPStatusError{StatusCode: resp.StatusCode}
	}
	var raw []apiStudy
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&raw); err != nil {
		return nil, ErrInvalidStudyResponse
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, ErrInvalidStudyResponse
	}
	studies := make([]models.Study, 0, len(raw))
	fromDate := dateNumber(from)
	toDate := dateNumber(to)
	for _, item := range raw {
		// The current server may return records older than `inicios`.
		// Enforce the requested inclusive range before exposing results to AP1.
		if item.Date >= fromDate && item.Date <= toDate {
			studies = append(studies, item.study())
		}
	}
	return studies, nil
}

func (r *HttpStudyRepository) buildURL(from, to time.Time) (string, error) {
	base, err := url.Parse(strings.TrimRight(r.config.BaseURL, "/"))
	if err != nil {
		return "", err
	}
	path, err := url.Parse("/" + strings.TrimLeft(r.config.GetStudiesPath, "/"))
	if err != nil {
		return "", err
	}
	endpoint := base.ResolveReference(path)
	endpoint.RawQuery = "inicio=" + url.QueryEscape(formatAPIDate(from)) + "&final=" + url.QueryEscape(formatAPIDate(to))
	return endpoint.String(), nil
}

func (r *HttpStudyRepository) RetrieveStudy(ctx context.Context, studyUID, destination string) error {
	requestURL, err := r.downloadURL(studyUID)
	if err != nil {
		return ErrStudyDownload
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("%w: %v", ErrStudyDownload, err)
	}
	zipFile, err := os.CreateTemp(filepath.Dir(destination), ".study-*.zip")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStudyDownload, err)
	}
	zipPath := zipFile.Name()
	defer os.Remove(zipPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		_ = zipFile.Close()
		return ErrStudyDownload
	}
	resp, err := r.client.Do(req)
	if err != nil {
		_ = zipFile.Close()
		return ErrStudyDownload
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = zipFile.Close()
		return ErrStudyDownload
	}
	if _, err = io.Copy(zipFile, resp.Body); err != nil {
		_ = zipFile.Close()
		return ErrStudyDownload
	}
	if err = zipFile.Close(); err != nil {
		return ErrStudyDownload
	}
	if err = extractStudyZIP(zipPath, destination); err != nil {
		return err
	}
	return nil
}

func (r *HttpStudyRepository) downloadURL(studyUID string) (string, error) {
	if strings.TrimSpace(studyUID) == "" {
		return "", errors.New("empty study UID")
	}
	base, err := url.Parse(strings.TrimRight(r.config.BaseURL, "/"))
	if err != nil {
		return "", err
	}
	path := "/" + strings.Trim(r.config.DownloadStudyPath, "/") + "/" + url.PathEscape(studyUID)
	return base.ResolveReference(&url.URL{Path: path}).String(), nil
}

func extractStudyZIP(zipPath, destination string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return ErrInvalidStudyZIP
	}
	defer archive.Close()
	root, err := filepath.Abs(destination)
	if err != nil {
		return ErrInvalidStudyZIP
	}
	dicomCount := 0
	for _, entry := range archive.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		target := filepath.Join(root, filepath.FromSlash(entry.Name))
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return ErrInvalidStudyZIP
		}
		rel, err := filepath.Rel(root, targetAbs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(entry.Name) {
			return ErrInvalidStudyZIP
		}
		if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
			return ErrInvalidStudyZIP
		}
		in, err := entry.Open()
		if err != nil {
			return ErrInvalidStudyZIP
		}
		out, err := os.OpenFile(targetAbs, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			in.Close()
			return ErrInvalidStudyZIP
		}
		_, copyErr := io.Copy(out, in)
		closeInErr, closeOutErr := in.Close(), out.Close()
		if copyErr != nil || closeInErr != nil || closeOutErr != nil {
			return ErrInvalidStudyZIP
		}
		if isPart10DICOM(targetAbs) {
			dicomCount++
		}
	}
	if dicomCount == 0 {
		return ErrNoDICOMFiles
	}
	return nil
}

func isPart10DICOM(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	header := make([]byte, 132)
	if _, err := io.ReadFull(reader, header); err != nil {
		return false
	}
	return string(header[128:132]) == "DICM"
}

func formatAPIDate(value time.Time) string { return value.Format("20060102") }

func dateNumber(value time.Time) int {
	return value.Year()*10000 + int(value.Month())*100 + value.Day()
}

func formatPersonName(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(value), "^", " ")), " ")
}

// These structs reflect the response observed from /getestudios on 2026-08-24.
type apiStudy struct {
	ID              int         `json:"ID"`
	Date            int         `json:"FECHA"`
	Name            string      `json:"NOMBRE"`
	Sex             string      `json:"SEXO"`
	PatientStudy    string      `json:"PAS_ID"`
	BirthDate       string      `json:"PAS_BD"`
	PatientAge      string      `json:"PAS_AGE"`
	ExternalStudyID string      `json:"EST_UID"`
	Report          string      `json:"INFORM"`
	Series          []apiSeries `json:"SERIES"`
}
type apiSeries struct {
	UID      int       `json:"SER_UID"`
	ID       string    `json:"SER_ID"`
	Modality string    `json:"MODALIDAD"`
	Name     string    `json:"NOMBRE"`
	Files    []apiFile `json:"FILES"`
}
type apiFile struct {
	Position int    `json:"POS"`
	UID      string `json:"INS_UID"`
	AE       string `json:"AE"`
}

func (s apiStudy) study() models.Study {
	instances := 0
	seriesModels := make([]models.Series, 0, len(s.Series))
	modality := ""
	description := ""
	for _, series := range s.Series {
		instances += len(series.Files)
		if modality == "" {
			modality = strings.TrimSpace(series.Modality)
		}
		if description == "" {
			description = strings.TrimSpace(series.Name)
		}
		files := make([]models.StudyFile, 0, len(series.Files))
		for _, file := range series.Files {
			files = append(files, models.StudyFile{Position: file.Position, InstanceUID: file.UID, AE: file.AE})
		}
		seriesModels = append(seriesModels, models.Series{SeriesID: fmt.Sprint(series.UID), SeriesUID: series.ID, SeriesInstanceUID: series.ID, SeriesDescription: series.Name, Modality: series.Modality, InstanceCount: len(series.Files), Files: files})
	}
	date := fmt.Sprintf("%08d", s.Date)
	if len(date) == 8 {
		date = date[:4] + "-" + date[4:6] + "-" + date[6:]
	}
	return models.Study{
		StudyID: s.ID, StudyInstanceUID: s.ExternalStudyID,
		PatientID: strings.TrimSpace(s.PatientStudy), PatientName: formatPersonName(s.Name),
		SourcePatientName: strings.TrimSpace(s.Name), PatientBirthDate: strings.TrimSpace(s.BirthDate),
		PatientAge: strings.TrimSpace(s.PatientAge), PatientSex: strings.TrimSpace(s.Sex), StudyDescription: description,
		StudyDate: date, Modality: modality, SeriesCount: len(s.Series), InstanceCount: instances, Series: seriesModels,
	}
}
