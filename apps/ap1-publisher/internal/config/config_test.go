package config

import "testing"

func TestPACSConfigIsConfigured(t *testing.T) {
	valid := PACSConfig{Host: "pacs.local", Port: 104, CalledAETitle: "PACS", CallingAETitle: "AP1", MoveDestinationAETitle: "AP1", ReceivePort: 11112}
	if !valid.IsConfigured() {
		t.Fatal("valid PACS configuration rejected")
	}
	valid.Host = ""
	if valid.IsConfigured() {
		t.Fatal("incomplete PACS configuration accepted")
	}
}

func TestStudyAPIDefaultTimeout(t *testing.T) {
	if got := (StudyAPIConfig{BaseURL: "http://studies", GetStudiesPath: "/getestudios", DownloadStudyPath: "/DescargaEstudio"}); !got.IsConfigured() {
		t.Fatal("valid API configuration rejected")
	}
}
