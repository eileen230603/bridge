package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/discdrive"
)

type fakeDrives []discdrive.Drive

func (f fakeDrives) List(context.Context) ([]discdrive.Drive, error) { return f, nil }

func newSettingsApp(t *testing.T, drives discdrive.Detector) *App {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	if err := os.WriteFile(path, []byte(`{"temporaryDirectory":"tmp","completedDirectory":"done","studyApi":{"protocol":"http","host":"v1.local","port":4000,"timeoutSeconds":10},"epson":{"enabled":true,"defaultCopies":1,"monitoringFolder":"orders"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.Epson.MonitoringFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	return &App{cfg: cfg, configPath: path, devMode: true, drives: drives, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func TestSaveModeValidatesOnlyActiveAndKeepsOtherSpaces(t *testing.T) {
	app := newSettingsApp(t, nil)
	m := app.GetModeConfig()
	m.Active = config.ModeServerV2
	m.V2 = config.ServerConfig{Protocol: "https", Host: "v2.local", Port: 443, TimeoutSeconds: 10}
	if err := app.SaveModeConfig(m); err == nil || !strings.Contains(err.Error(), "API Key") {
		t.Fatalf("V2 sin clave: %v", err)
	}
	if app.cfg.Mode.Active != config.ModeServerV1 {
		t.Fatal("un guardado inválido cambió el modo")
	}
	m.V2.APIKey = " secreta "
	m.SCP = config.SCPConfig{AETitle: "", Port: 0} // a medio llenar: no se valida
	if err := app.SaveModeConfig(m); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(app.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Mode.Active != config.ModeServerV2 || loaded.Mode.V2.APIKey != "secreta" || loaded.StudyAPI.Host != "v1.local" {
		t.Fatalf("persistido: %+v %+v", loaded.Mode, loaded.StudyAPI)
	}
	if loaded.StudyAPI.APIKey != "" {
		t.Fatal("la clave de V2 se filtró a V1")
	}
	if got := activeServer(app.cfg); got.Host != "v2.local" || got.APIKey != "secreta" {
		t.Fatalf("servidor activo: %+v", got)
	}
}

func TestSaveDestinationRules(t *testing.T) {
	ready := discdrive.Drive{ID: "dev:D", Label: "D:", CanWriteDVD: true, Media: discdrive.Media{Present: true, Type: "DVD-R", Kind: "DVD", Blank: true, CapacityBytes: 4_700_000_000}}
	app := newSettingsApp(t, fakeDrives{ready})
	d := app.GetDestinationConfig()

	d.Robot.Enabled, d.Drives.Enabled = false, false
	if err := app.SaveDestinationConfig(d); err == nil {
		t.Fatal("se aceptó sin destinos habilitados")
	}
	d.Drives.Enabled = true
	if err := app.SaveDestinationConfig(d); err == nil {
		t.Fatal("se aceptó unidades sin seleccionar ninguna")
	}
	d.Drives.Selected = []string{"no-conectada"}
	if err := app.SaveDestinationConfig(d); err == nil {
		t.Fatal("se aceptó una unidad desconectada")
	}
	d.Drives.Selected = []string{"dev:D", "dev:D"}
	d.Drives.DefaultCopies = 2
	if err := app.SaveDestinationConfig(d); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(app.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Epson.Enabled || !loaded.Drives.Enabled || len(loaded.Drives.Selected) != 1 || loaded.Drives.DefaultCopies != 2 {
		t.Fatalf("persistido: %+v %+v", loaded.Epson, loaded.Drives)
	}

	// Robot habilitado pero TD Bridge apagado: no se guarda y nada cambia.
	before, _ := os.ReadFile(app.configPath)
	if err := os.RemoveAll(app.cfg.Epson.MonitoringFolder); err != nil {
		t.Fatal(err)
	}
	d.Robot.Enabled = true
	if err := app.SaveDestinationConfig(d); err == nil || !strings.Contains(err.Error(), "Robot Epson no disponible") {
		t.Fatalf("robot apagado: %v", err)
	}
	after, _ := os.ReadFile(app.configPath)
	if string(before) != string(after) {
		t.Fatal("un guardado rechazado modificó el archivo")
	}
}

func TestDestinationStatusReadyToBurn(t *testing.T) {
	noDisc := discdrive.Drive{ID: "dev:E", Label: "E:", CanWriteDVD: true}
	app := newSettingsApp(t, fakeDrives{noDisc})
	app.cfg.Epson.Enabled = false
	app.cfg.Drives = config.DriveConfig{Enabled: true, Selected: []string{"dev:E"}, DefaultCopies: 1}
	st := app.GetDestinationStatus()
	if st.ReadyToBurn || len(st.Drives) != 1 || st.Drives[0].State != discdrive.StateNoDisc || !st.Drives[0].Selected {
		t.Fatalf("sin disco: %+v", st)
	}
	app.cfg.Epson.Enabled = true
	if st := app.GetDestinationStatus(); !st.ReadyToBurn || !st.Robot.Available {
		t.Fatalf("robot listo: %+v", st)
	}
}

func TestSaveLabelSettingsValidatesPaper(t *testing.T) {
	app := newSettingsApp(t, nil)
	s := app.GetLabelSettings()
	s.Paper = config.PaperLabelConfig{Enabled: true, Size: config.PaperSizeCircularCD}
	if err := app.SaveLabelSettings(s); err == nil {
		t.Fatal("papel sin impresora aceptado")
	}
	s.Paper.Printer = "Brother"
	s.Label.HospitalName = "Hospital"
	if err := app.SaveLabelSettings(s); err != nil {
		t.Fatal(err)
	}
	loaded, _ := config.Load(app.configPath)
	if loaded.PaperLabel.Printer != "Brother" || loaded.DiscLabel.HospitalName != "Hospital" {
		t.Fatalf("persistido: %+v %+v", loaded.PaperLabel, loaded.DiscLabel)
	}
}
