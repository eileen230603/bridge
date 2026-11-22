package config

import "testing"

func TestDiscFormatCombinations(t *testing.T) {
	valid := []EpsonConfig{
		{DiscType: "CD", Format: "ISO9660"}, {DiscType: "CD", Format: "UDF102"},
		{DiscType: "DVD", Format: "UDF102"}, {DiscType: "DVD", Format: "UDF150"}, {DiscType: "DVD", Format: "ISO9660"},
		{DiscType: "BD", Format: "UDF250"}, {DiscType: "dvd", Format: "udf102"},
	}
	for _, c := range valid {
		if err := c.ValidateDisc(); err != nil {
			t.Errorf("%s/%s debería ser válido: %v", c.DiscType, c.Format, err)
		}
	}
	invalid := []EpsonConfig{
		{DiscType: "BD", Format: "UDF102"}, {DiscType: "BD", Format: "ISO9660"}, {DiscType: "CD", Format: "UDF250"},
		{DiscType: "DVD", Format: "UDF250"}, {DiscType: "LP", Format: "UDF102"},
	}
	for _, c := range invalid {
		if err := c.ValidateDisc(); err == nil {
			t.Errorf("%s/%s debería ser inválido", c.DiscType, c.Format)
		}
	}
}

func TestDefaultFormatIsAlwaysValid(t *testing.T) {
	for _, p := range DiscProfiles() {
		c := EpsonConfig{DiscType: p.Type, Format: p.Default}
		if err := c.ValidateDisc(); err != nil {
			t.Errorf("el formato por defecto de %s no es válido: %v", p.Type, err)
		}
		if p.Formats[0].ID != p.Default {
			t.Errorf("%s: el primer formato debe ser el recomendado", p.Type)
		}
	}
}

func TestSetDefaultsUsesDiscRecommendedFormat(t *testing.T) {
	c := EpsonConfig{DiscType: "CD"}
	c.SetDefaults()
	if c.Format != "ISO9660" {
		t.Fatalf("CD sin formato debe usar ISO9660, obtuvo %s", c.Format)
	}
}

func TestLoadMigratesLegacyInvalidCombination(t *testing.T) {
	raw := []byte(`{"epson":{"discType":"BD","format":"UDF102","enabled":true}}`)
	c, err := LoadBytes(raw, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if c.Epson.Format != "UDF250" {
		t.Fatalf("BD con UDF102 debe migrar al formato recomendado, obtuvo %s", c.Epson.Format)
	}
}

func TestLoadRejectsUnknownDisc(t *testing.T) {
	if _, err := LoadBytes([]byte(`{"epson":{"discType":"LP","format":"UDF102"}}`), t.TempDir()); err == nil {
		t.Fatal("un tipo de disco desconocido debe rechazarse")
	}
}
