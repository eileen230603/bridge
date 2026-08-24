package adapters

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
)

func TestFormatAPIDate(t *testing.T) {
	value, _ := time.Parse("2006-01-02", "2026-08-24")
	if got := formatAPIDate(value); got != "20260824" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildURL(t *testing.T) {
	r := NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: "http://example.test:4000/", GetStudiesPath: "/getestudios", DownloadStudyPath: "/DescargaEstudio", TimeoutSeconds: 15})
	from, _ := time.Parse("2006-01-02", "2022-01-01")
	to, _ := time.Parse("2006-01-02", "2026-01-01")
	got, err := r.buildURL(from, to)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://example.test:4000/getestudios?inicio=20220101&final=20260101" {
		t.Fatalf("unexpected URL: %s", got)
	}
	parsed, _ := url.Parse(got)
	if parsed.Scheme != "http" || parsed.Host != "example.test:4000" || parsed.Path != "/getestudios" || parsed.Query().Get("inicio") != "20220101" || parsed.Query().Get("final") != "20260101" {
		t.Fatalf("unexpected URL: %s", got)
	}
}

func TestHTTPStudyRepositoryMapsRealResponse(t *testing.T) {
	response := `[{"ID":15491,"FECHA":20251231,"NOMBRE":"PEREZ^MARIA","SEXO":"F","PAS_ID":"10095 PELVIS C/C","PAS_BD":"19750410","PAS_AGE":"050Y","EST_UID":"e57702f0-85ab520a-bb7c7c37-2c72599c-c96c176c","INFORM":"","SERIES":[{"SER_UID":1,"SER_ID":"series-uuid","MODALIDAD":"MR","NOMBRE":"T2","FILES":[{"POS":1,"INS_UID":"instance-uuid","AE":"DEV1"}]},{"SER_UID":2,"SER_ID":"series-2","MODALIDAD":"CT","NOMBRE":"LOCALIZER","FILES":[{"POS":1,"INS_UID":"instance-2","AE":"DEV1"}]}]}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	r := NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: server.URL, GetStudiesPath: "/getestudios", TimeoutSeconds: 1})
	from, _ := time.Parse("2006-01-02", "2025-12-31")
	studies, err := r.SearchStudies(context.Background(), from, from)
	if err != nil {
		t.Fatal(err)
	}
	if len(studies) != 1 {
		t.Fatalf("got %d studies", len(studies))
	}
	s := studies[0]
	if s.StudyID != 15491 || s.PatientName != "PEREZ MARIA" || s.SourcePatientName != "PEREZ^MARIA" || s.PatientID != "10095 PELVIS C/C" || s.PatientBirthDate != "19750410" || s.PatientAge != "050Y" || s.PatientSex != "F" || s.StudyDescription != "T2" || s.StudyDate != "2025-12-31" || s.Modality != "MR" || s.SeriesCount != 2 || s.InstanceCount != 2 {
		t.Fatalf("unexpected mapping: %+v", s)
	}
	if s.StudyInstanceUID != "e57702f0-85ab520a-bb7c7c37-2c72599c-c96c176c" {
		t.Fatalf("EST_UID not preserved: %q", s.StudyInstanceUID)
	}
	if len(s.Series) != 2 || len(s.Series[0].Files) != 1 || s.Series[0].Files[0].Position != 1 || s.Series[0].Files[0].InstanceUID != "instance-uuid" {
		t.Fatalf("series/files not preserved: %+v", s.Series)
	}
}

func TestDownloadStudyExtractsDICOMIntoData(t *testing.T) {
	zipBody := studyZIP(t, map[string][]byte{"series/image.dcm": dicomBytes()})
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestedPath = req.URL.EscapedPath()
		_, _ = w.Write(zipBody)
	}))
	defer server.Close()
	r := NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: server.URL, GetStudiesPath: "/getestudios", DownloadStudyPath: "/DescargaEstudio", TimeoutSeconds: 1})
	destination := filepath.Join(t.TempDir(), "data")
	if err := r.RetrieveStudy(context.Background(), "study-uuid", destination); err != nil {
		t.Fatal(err)
	}
	if requestedPath != "/DescargaEstudio/study-uuid" {
		t.Fatalf("unexpected path: %s", requestedPath)
	}
	if _, err := os.Stat(filepath.Join(destination, "series", "image.dcm")); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(destination), ".study-*.zip"))
	if len(matches) != 0 {
		t.Fatalf("temporary ZIP remains: %v", matches)
	}
}

func TestDownloadStudyRejectsZipSlip(t *testing.T) {
	zipBody := studyZIP(t, map[string][]byte{"../outside.dcm": dicomBytes()})
	r, closeServer := downloadRepository(t, http.StatusOK, zipBody)
	defer closeServer()
	destination := filepath.Join(t.TempDir(), "data")
	err := r.RetrieveStudy(context.Background(), "uuid", destination)
	if !errors.Is(err, ErrInvalidStudyZIP) {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(destination), "outside.dcm")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("outside file created: %v", err)
	}
}

func TestDownloadStudyRejectsEmptyZIP(t *testing.T) {
	r, closeServer := downloadRepository(t, http.StatusOK, studyZIP(t, nil))
	defer closeServer()
	err := r.RetrieveStudy(context.Background(), "uuid", filepath.Join(t.TempDir(), "data"))
	if !errors.Is(err, ErrNoDICOMFiles) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownloadStudyRejectsInvalidZIP(t *testing.T) {
	r, closeServer := downloadRepository(t, http.StatusOK, []byte("not a zip"))
	defer closeServer()
	err := r.RetrieveStudy(context.Background(), "uuid", filepath.Join(t.TempDir(), "data"))
	if !errors.Is(err, ErrInvalidStudyZIP) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownloadStudyHTTPError(t *testing.T) {
	r, closeServer := downloadRepository(t, http.StatusBadGateway, nil)
	defer closeServer()
	err := r.RetrieveStudy(context.Background(), "uuid", filepath.Join(t.TempDir(), "data"))
	if !errors.Is(err, ErrStudyDownload) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownloadStudyTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { time.Sleep(50 * time.Millisecond) }))
	defer server.Close()
	r := NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: server.URL, DownloadStudyPath: "/DescargaEstudio", TimeoutSeconds: 1})
	r.client.Timeout = 5 * time.Millisecond
	err := r.RetrieveStudy(context.Background(), "uuid", filepath.Join(t.TempDir(), "data"))
	if !errors.Is(err, ErrStudyDownload) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func downloadRepository(t *testing.T, status int, body []byte) (*HttpStudyRepository, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(status); _, _ = w.Write(body) }))
	return NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: server.URL, DownloadStudyPath: "/DescargaEstudio", TimeoutSeconds: 1}), server.Close
}

func studyZIP(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, data := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func dicomBytes() []byte { data := make([]byte, 132); copy(data[128:], "DICM"); return data }

func TestHTTPStudyRepositoryEmptyList(t *testing.T) {
	r, closeServer := testHTTPRepository(t, http.StatusOK, `[]`)
	defer closeServer()
	studies, err := r.SearchStudies(context.Background(), time.Now(), time.Now())
	if err != nil || studies == nil || len(studies) != 0 {
		t.Fatalf("studies=%v err=%v", studies, err)
	}
}

func TestHTTPStudyRepositoryFiltersServerResultsOutsideRequestedRange(t *testing.T) {
	response := `[
		{"FECHA":20251231,"NOMBRE":"OLD","PAS_ID":"OLD","SERIES":[]},
		{"FECHA":20260103,"NOMBRE":"FROM","PAS_ID":"FROM","SERIES":[]},
		{"FECHA":20260105,"NOMBRE":"TO","PAS_ID":"TO","SERIES":[]},
		{"FECHA":20260106,"NOMBRE":"NEW","PAS_ID":"NEW","SERIES":[]}
	]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	r := NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: server.URL, GetStudiesPath: "/getestudios", TimeoutSeconds: 1})
	from, _ := time.Parse("2006-01-02", "2026-01-03")
	to, _ := time.Parse("2006-01-02", "2026-01-05")
	studies, err := r.SearchStudies(context.Background(), from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(studies) != 2 || studies[0].StudyDate != "2026-01-03" || studies[1].StudyDate != "2026-01-05" {
		t.Fatalf("unexpected filtered studies: %+v", studies)
	}
}

func TestHTTPStudyRepositoryStatusError(t *testing.T) {
	r, closeServer := testHTTPRepository(t, http.StatusBadGateway, `{}`)
	defer closeServer()
	_, err := r.SearchStudies(context.Background(), time.Now(), time.Now())
	var statusErr HTTPStatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPStudyRepositoryInvalidJSON(t *testing.T) {
	r, closeServer := testHTTPRepository(t, http.StatusOK, `{invalid`)
	defer closeServer()
	_, err := r.SearchStudies(context.Background(), time.Now(), time.Now())
	if !errors.Is(err, ErrInvalidStudyResponse) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPStudyRepositoryTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()
	r := NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: server.URL, GetStudiesPath: "/getestudios", TimeoutSeconds: 1})
	r.client.Timeout = 5 * time.Millisecond
	_, err := r.SearchStudies(context.Background(), time.Now(), time.Now())
	if !errors.Is(err, ErrStudyServerUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func testHTTPRepository(t *testing.T, status int, body string) (*HttpStudyRepository, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(status); _, _ = w.Write([]byte(body)) }))
	return NewHttpStudyRepository(config.StudyAPIConfig{BaseURL: server.URL, GetStudiesPath: "/getestudios", DownloadStudyPath: "/DescargaEstudio", TimeoutSeconds: 1}), server.Close
}
