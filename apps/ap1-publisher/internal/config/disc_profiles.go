package config

import (
	"fmt"
	"strings"
)

// DiscFormat es un sistema de archivos que TD Bridge puede grabar (valor FORMAT del JDF).
type DiscFormat struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// DiscProfile agrupa lo válido para un tipo de disco: es la ÚNICA fuente de verdad
// de qué formatos se pueden combinar con cada disco (la interfaz la lee de aquí).
type DiscProfile struct {
	Type    string       `json:"type"`    // valor DISC_TYPE del JDF
	Label   string       `json:"label"`   // nombre para mostrar
	Default string       `json:"default"` // formato recomendado
	Formats []DiscFormat `json:"formats"` // formatos válidos, el primero es el recomendado
	// Note se muestra en la interfaz cuando algo del perfil requiere atención.
	Note string `json:"note,omitempty"`
}

var (
	fmtISO    = DiscFormat{"ISO9660", "ISO 9660"}
	fmtUDF102 = DiscFormat{"UDF102", "UDF 1.02"}
	fmtUDF150 = DiscFormat{"UDF150", "UDF 1.50"}
	fmtUDF250 = DiscFormat{"UDF250", "UDF 2.50"}
)

var discProfiles = []DiscProfile{
	{Type: "CD", Label: "CD", Default: "ISO9660", Formats: []DiscFormat{fmtISO, fmtUDF102}},
	{Type: "DVD", Label: "DVD", Default: "UDF102", Formats: []DiscFormat{fmtUDF102, fmtUDF150, fmtISO}},
	// Blu-ray exige UDF 2.50 o superior según su especificación. TODO: confirmar con la
	// documentación de TD Bridge que acepta FORMAT=UDF250 (no está en el repositorio).
	{Type: "BD", Label: "Blu-ray", Default: "UDF250", Formats: []DiscFormat{fmtUDF250},
		Note: "Formato UDF 2.50 pendiente de confirmar con TD Bridge."},
}

// DiscProfiles devuelve los perfiles de disco soportados (copia).
func DiscProfiles() []DiscProfile {
	out := make([]DiscProfile, len(discProfiles))
	copy(out, discProfiles)
	return out
}

func profileFor(discType string) (DiscProfile, bool) {
	t := strings.ToUpper(strings.TrimSpace(discType))
	for _, p := range discProfiles {
		if p.Type == t {
			return p, true
		}
	}
	return DiscProfile{}, false
}

// DefaultFormatFor devuelve el formato recomendado para un tipo de disco ("" si no existe).
func DefaultFormatFor(discType string) string {
	if p, ok := profileFor(discType); ok {
		return p.Default
	}
	return ""
}

// ValidateDisc comprueba que el tipo de disco exista y que el formato sea válido para él.
func (c EpsonConfig) ValidateDisc() error {
	p, ok := profileFor(c.DiscType)
	if !ok {
		return fmt.Errorf("Tipo de disco inválido: %q.", c.DiscType)
	}
	f := strings.ToUpper(strings.TrimSpace(c.Format))
	for _, allowed := range p.Formats {
		if allowed.ID == f {
			return nil
		}
	}
	return fmt.Errorf("El formato %s no es válido para %s.", c.Format, p.Label)
}

// normalizeDisc corrige configuraciones antiguas: pasa a mayúsculas y, si el formato no
// corresponde al disco (p. ej. Blu-ray con UDF 1.02), usa el recomendado para ese disco.
func (c *EpsonConfig) normalizeDisc() {
	c.DiscType = strings.ToUpper(strings.TrimSpace(c.DiscType))
	c.Format = strings.ToUpper(strings.TrimSpace(c.Format))
	if _, ok := profileFor(c.DiscType); !ok {
		return // lo rechazará ValidateDisc
	}
	if c.ValidateDisc() != nil {
		c.Format = DefaultFormatFor(c.DiscType)
	}
}
