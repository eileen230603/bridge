package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStudyAPIDefaultTimeout(t *testing.T) {
	if got := (StudyAPIConfig{BaseURL: "http://studies:80", GetStudiesPath: "/getestudios", DownloadStudyPath: "/DescargaEstudio", TimeoutSeconds: 15}); !got.IsConfigured() {
		t.Fatal("valid API configuration rejected")
	}
}

func TestServerConfigValidationAndBaseAddress(t *testing.T) {
	valid := ServerConfig{Protocol: "http", Host: "server.local", Port: 4000, TimeoutSeconds: 60}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	if got, _ := valid.BaseAddress(); got != "http://server.local:4000" {
		t.Fatalf("got %q", got)
	}
	cases := []ServerConfig{
		{Protocol: "http", Port: 4000, TimeoutSeconds: 60},
		{Protocol: "http", Host: "host", Port: 0, TimeoutSeconds: 60},
		{Protocol: "http", Host: "host", Port: 4000, TimeoutSeconds: 0},
		{Protocol: "ftp", Host: "host", Port: 4000, TimeoutSeconds: 60},
	}
	for _, c := range cases {
		if c.Validate() == nil {
			t.Fatalf("accepted invalid config: %+v", c)
		}
	}
}

func TestLoadAndSaveServerConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"temporaryDirectory":"tmp","completedDirectory":"done","studyApi":{"protocol":"https","host":"pacs.local","port":8443,"timeoutSeconds":9},"epson":{"defaultCopies":1}}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil || cfg.StudyAPI.Host != "pacs.local" {
		t.Fatalf("load: %+v %v", cfg, err)
	}
	cfg.StudyAPI.Port = 9443
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Load(path)
	if err != nil || reloaded.StudyAPI.Port != 9443 {
		t.Fatalf("reload: %+v %v", reloaded, err)
	}
}
