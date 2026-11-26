// Package discdrive detecta las grabadoras CD/DVD/Blu-ray conectadas al equipo
// y el disco que tienen insertado. Windows usa IMAPI2 y macOS drutil.
package discdrive

import (
	"context"
	"errors"
)

// Media describe el disco insertado en una unidad.
type Media struct {
	Present       bool   `json:"present"`
	Type          string `json:"type"` // p. ej. "DVD-R", "CD-RW", "BD-R"
	Kind          string `json:"kind"` // CD, DVD o BD
	Blank         bool   `json:"blank"`
	Rewritable    bool   `json:"rewritable"`
	CapacityBytes int64  `json:"capacityBytes"`
}

// Drive es una grabadora conectada.
type Drive struct {
	ID          string `json:"id"`    // estable entre reinicios (no depende de la letra)
	Name        string `json:"name"`  // fabricante y modelo
	Label       string `json:"label"` // «D:» en Windows, «Unidad 1» en macOS
	Device      string `json:"device"`
	CanWriteCD  bool   `json:"canWriteCd"`
	CanWriteDVD bool   `json:"canWriteDvd"`
	CanWriteBD  bool   `json:"canWriteBd"`
	Media       Media  `json:"media"`
}

// Estados de una unidad para la interfaz.
const (
	StateReady    = "listo"
	StateNoDisc   = "sin-disco"
	StateNotBlank = "no-virgen"
	StateNoWrite  = "no-compatible"
)

// State resume si la unidad puede grabar ahora mismo. Solo se aceptan discos
// vírgenes no regrabables (decisión de diseño v2.2).
func (d Drive) State() (string, string) {
	m := d.Media
	switch {
	case !m.Present:
		return StateNoDisc, "Sin disco"
	case m.Rewritable || !m.Blank:
		return StateNotBlank, "Inserte un disco virgen"
	case !d.canWrite(m.Kind):
		return StateNoWrite, "La unidad no graba " + m.Kind
	default:
		return StateReady, m.Type + " virgen · " + FormatBytes(m.CapacityBytes)
	}
}

func (d Drive) canWrite(kind string) bool {
	switch kind {
	case "CD":
		return d.CanWriteCD
	case "DVD":
		return d.CanWriteDVD
	case "BD":
		return d.CanWriteBD
	}
	return false
}

// Detector lista las unidades. Cada sistema operativo tiene el suyo.
type Detector interface {
	List(ctx context.Context) ([]Drive, error)
}

// ErrUnsupported indica que este sistema operativo no tiene detección de unidades.
var ErrUnsupported = errors.New("La grabación en unidades no está disponible en este sistema operativo.")
