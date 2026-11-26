package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Una configuración v2.0 (sin mode/drives/paperLabel) se migra sin perder nada.
func TestLegacyConfigMigratesWithoutLoss(t *testing.T) {
	raw := `{"temporaryDirectory":"tmp","completedDirectory":"done","studyApi":{"protocol":"http","host":"pacs","port":4000,"timeoutSeconds":20},"epson":{"enabled":true,"defaultCopies":2,"discType":"CD","format":"ISO9660L2","printMode":1,"labelType":1}}`
	cfg, err := LoadBytes([]byte(raw), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode.Active != ModeServerV1 || cfg.StudyAPI.Host != "pacs" {
		t.Fatalf("modo: %+v", cfg.Mode)
	}
	if !cfg.Epson.Enabled || cfg.Epson.DiscType != "CD" || cfg.Epson.DefaultCopies != 2 || cfg.Epson.PrintMode != 1 {
		t.Fatalf("robot alterado: %+v", cfg.Epson)
	}
	if cfg.Drives.Enabled || !cfg.Drives.Verify || cfg.Drives.DefaultCopies != 1 || cfg.Drives.Selected == nil {
		t.Fatalf("unidades por defecto: %+v", cfg.Drives)
	}
	if cfg.PaperLabel.Enabled || !cfg.PaperLabel.Auto || cfg.PaperLabel.Size != PaperSizeCircularCD {
		t.Fatalf("papel por defecto: %+v", cfg.PaperLabel)
	}
	if cfg.Mode.SCP.Port != 11112 || cfg.Mode.SCP.AETitle == "" {
		t.Fatalf("scp por defecto: %+v", cfg.Mode.SCP)
	}
}

// Un «false» guardado explícitamente no se pisa con el valor por defecto.
func TestExplicitFalseIsKept(t *testing.T) {
	raw := `{"epson":{"defaultCopies":1},"drives":{"enabled":true,"selected":["mac:A"],"verify":false,"eject":true,"defaultCopies":3},"paperLabel":{"enabled":false,"printer":"P","size":"a4-2","auto":false}}`
	cfg, err := LoadBytes([]byte(raw), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Drives.Verify || !cfg.Drives.Eject || cfg.Drives.DefaultCopies != 3 || len(cfg.Drives.Selected) != 1 {
		t.Fatalf("unidades: %+v", cfg.Drives)
	}
	if cfg.PaperLabel.Auto || cfg.PaperLabel.Size != PaperSizeA4Two {
		t.Fatalf("papel: %+v", cfg.PaperLabel)
	}
}

// Guardar y recargar conserva cada espacio por separado.
func TestModeSpacesRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"epson":{"defaultCopies":1}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.StudyAPI = ServerConfig{Protocol: "http", Host: "v1", Port: 4000, TimeoutSeconds: 10}
	cfg.Mode.Active = ModeServerV2
	cfg.Mode.V2 = ServerConfig{Protocol: "https", Host: "v2", Port: 443, TimeoutSeconds: 10, APIKey: "k"}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.StudyAPI.Host != "v1" || got.Mode.V2.Host != "v2" || got.Mode.V2.APIKey != "k" || got.Mode.Active != ModeServerV2 {
		t.Fatalf("round trip: %+v %+v", got.StudyAPI, got.Mode)
	}
}

func TestValidateMode(t *testing.T) {
	ok := ServerConfig{Protocol: "http", Host: "h", Port: 1, TimeoutSeconds: 1}
	if ValidateMode(ModeServerV2, ServerConfig{}, ok, SCPConfig{}) == nil {
		t.Fatal("V2 sin API Key aceptado")
	}
	withKey := ok
	withKey.APIKey = "x"
	if err := ValidateMode(ModeServerV2, ServerConfig{}, withKey, SCPConfig{}); err != nil {
		t.Fatal(err)
	}
	// Solo se valida el modo activo: V1 vacío no impide guardar SCP.
	if err := ValidateMode(ModeDICOMSCP, ServerConfig{}, ServerConfig{}, SCPConfig{AETitle: "AP1", Port: 104}); err != nil {
		t.Fatal(err)
	}
	for _, scp := range []SCPConfig{{AETitle: "", Port: 104}, {AETitle: "ESTE_AE_TITLE_ES_LARGO", Port: 104}, {AETitle: "A\\B", Port: 104}, {AETitle: "AP1", Port: 0}} {
		if scp.Validate() == nil {
			t.Fatalf("SCP inválido aceptado: %+v", scp)
		}
	}
	if ValidateMode("otro", ok, ok, SCPConfig{}) == nil {
		t.Fatal("modo desconocido aceptado")
	}
}

func TestPaperLabelValidation(t *testing.T) {
	if (PaperLabelConfig{Enabled: true, Size: PaperSizeRoll62}).Validate() == nil {
		t.Fatal("habilitado sin impresora aceptado")
	}
	if (PaperLabelConfig{Size: "otro"}).Validate() == nil {
		t.Fatal("tamaño desconocido aceptado")
	}
	if err := (PaperLabelConfig{Enabled: true, Printer: "P", Size: PaperSizeA4Two}).Validate(); err != nil {
		t.Fatal(err)
	}
}
