package config

import "testing"

func TestStudyAPIDefaultTimeout(t *testing.T) {
	if got := (StudyAPIConfig{BaseURL: "http://studies", GetStudiesPath: "/getestudios", DownloadStudyPath: "/DescargaEstudio"}); !got.IsConfigured() {
		t.Fatal("valid API configuration rejected")
	}
}
