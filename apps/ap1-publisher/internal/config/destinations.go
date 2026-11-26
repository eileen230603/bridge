package config

import (
	"encoding/json"
	"errors"
	"strings"
)

// Modos de trabajo (de dónde salen los estudios). Solo uno está activo.
const (
	ModeServerV1 = "v1"
	ModeServerV2 = "v2"
	ModeDICOMSCP = "scp"
)

// ModeConfig guarda un espacio propio por modo: cambiar de modo nunca borra
// lo configurado en otro. Servidor V1 sigue en la clave histórica "studyApi".
type ModeConfig struct {
	Active string       `json:"active"`
	V2     ServerConfig `json:"v2"`
	SCP    SCPConfig    `json:"scp"`
}

// SCPConfig es el receptor DICOM (StoreSCP). La base SQLite es interna.
type SCPConfig struct {
	AETitle   string `json:"aeTitle"`
	Port      int    `json:"port"`
	AutoStart bool   `json:"autoStart"`
}

func (c SCPConfig) Validate() error {
	ae := strings.TrimSpace(c.AETitle)
	if ae == "" || len(ae) > 16 {
		return errors.New("El AE Title debe tener entre 1 y 16 caracteres.")
	}
	for _, r := range ae {
		if r < 0x20 || r > 0x7e || r == '\\' {
			return errors.New("El AE Title solo admite letras, números y símbolos básicos.")
		}
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("Puerto inválido.")
	}
	return nil
}

// ValidateActive valida solo el modo seleccionado; los demás pueden estar a medio llenar.
func ValidateMode(active string, v1, v2 ServerConfig, scp SCPConfig) error {
	switch active {
	case ModeServerV1:
		return v1.Validate()
	case ModeServerV2:
		if err := v2.Validate(); err != nil {
			return err
		}
		if strings.TrimSpace(v2.APIKey) == "" {
			return errors.New("Ingrese la API Key del servidor V2.")
		}
		return nil
	case ModeDICOMSCP:
		return scp.Validate()
	default:
		return errors.New("Modo de trabajo inválido.")
	}
}

// DriveConfig es el destino «Unidades CD/DVD» conectadas al equipo.
type DriveConfig struct {
	Enabled       bool     `json:"enabled"`
	Selected      []string `json:"selected"`
	Verify        bool     `json:"verify"`
	Eject         bool     `json:"eject"`
	DefaultCopies int      `json:"defaultCopies"`
}

func DefaultDriveConfig() DriveConfig {
	return DriveConfig{Enabled: false, Selected: []string{}, Verify: true, DefaultCopies: 1}
}

func (c DriveConfig) Validate() error {
	if c.DefaultCopies < 1 || c.DefaultCopies > 99 {
		return errors.New("Las copias deben estar entre 1 y 99.")
	}
	return nil
}

// Tamaños de etiqueta en papel adhesivo.
const (
	PaperSizeCircularCD = "circular-117"
	PaperSizeA4Two      = "a4-2"
	PaperSizeRoll62     = "roll-62"
)

type PaperSize struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func PaperSizes() []PaperSize {
	return []PaperSize{
		{PaperSizeCircularCD, "Circular CD · 117 mm"},
		{PaperSizeA4Two, "Hoja A4 · 2 etiquetas de CD"},
		{PaperSizeRoll62, "Rollo · 62 mm"},
	}
}

// PaperLabelConfig: impresión de la etiqueta en papel cuando se graba en una unidad.
type PaperLabelConfig struct {
	Enabled bool   `json:"enabled"`
	Printer string `json:"printer"`
	Size    string `json:"size"`
	Auto    bool   `json:"auto"`
}

func DefaultPaperLabelConfig() PaperLabelConfig {
	return PaperLabelConfig{Size: PaperSizeCircularCD, Auto: true}
}

func (c PaperLabelConfig) Validate() error {
	valid := false
	for _, s := range PaperSizes() {
		if s.ID == c.Size {
			valid = true
		}
	}
	if !valid {
		return errors.New("Tamaño de etiqueta inválido.")
	}
	if c.Enabled && strings.TrimSpace(c.Printer) == "" {
		return errors.New("Elija la impresora de la etiqueta en papel.")
	}
	return nil
}

// applyDestinationDefaults completa las secciones nuevas sin tocar las existentes.
// raw indica qué claves venían en el archivo, para distinguir «ausente» de «false».
func applyDestinationDefaults(c *Config, raw map[string]json.RawMessage) {
	if _, ok := raw["mode"]; !ok {
		c.Mode = ModeConfig{}
	}
	if c.Mode.Active == "" {
		c.Mode.Active = ModeServerV1
	}
	if c.Mode.V2.Protocol == "" {
		c.Mode.V2.Protocol = "http"
	}
	if c.Mode.V2.Port == 0 {
		c.Mode.V2.Port = 4000
	}
	if c.Mode.V2.TimeoutSeconds <= 0 {
		c.Mode.V2.TimeoutSeconds = 15
	}
	if c.Mode.SCP.AETitle == "" {
		c.Mode.SCP.AETitle = "SYMPHONY_SCP"
	}
	if c.Mode.SCP.Port == 0 {
		c.Mode.SCP.Port = 11112
	}
	if _, ok := raw["drives"]; !ok {
		c.Drives = DefaultDriveConfig()
	}
	if c.Drives.Selected == nil {
		c.Drives.Selected = []string{}
	}
	if c.Drives.DefaultCopies <= 0 {
		c.Drives.DefaultCopies = 1
	}
	if _, ok := raw["paperLabel"]; !ok {
		c.PaperLabel = DefaultPaperLabelConfig()
	}
	if c.PaperLabel.Size == "" {
		c.PaperLabel.Size = PaperSizeCircularCD
	}
}
