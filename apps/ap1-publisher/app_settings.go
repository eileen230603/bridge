package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/discdrive"
)

// ---------- Modo de trabajo (v2.1) ----------

// ModeSettings es lo que edita Configuración › Modo. Cada modo tiene su espacio.
type ModeSettings struct {
	Active string              `json:"active"`
	V1     config.ServerConfig `json:"v1"`
	V2     config.ServerConfig `json:"v2"`
	SCP    config.SCPConfig    `json:"scp"`
}

func (a *App) GetModeConfig() ModeSettings {
	a.mu.RLock()
	defer a.mu.RUnlock()
	v1 := a.cfg.StudyAPI
	v1.APIKey = ""
	return ModeSettings{Active: a.cfg.Mode.Active, V1: v1, V2: a.cfg.Mode.V2, SCP: a.cfg.Mode.SCP}
}

func cleanServer(s config.ServerConfig) config.ServerConfig {
	s.Protocol = strings.ToLower(strings.TrimSpace(s.Protocol))
	s.Host = strings.TrimSpace(s.Host)
	s.APIKey = strings.TrimSpace(s.APIKey)
	return s
}

// SaveModeConfig valida solo el modo seleccionado y guarda los tres espacios.
// Si la validación falla no se escribe nada y el modo anterior se mantiene.
func (a *App) SaveModeConfig(m ModeSettings) error {
	m.V1 = cleanServer(m.V1)
	m.V1.APIKey = ""
	m.V2 = cleanServer(m.V2)
	m.SCP.AETitle = strings.TrimSpace(m.SCP.AETitle)
	if err := config.ValidateMode(m.Active, m.V1, m.V2, m.SCP); err != nil {
		return err
	}
	a.mu.Lock()
	updated := a.cfg
	updated.StudyAPI = m.V1
	updated.Mode = config.ModeConfig{Active: m.Active, V2: m.V2, SCP: m.SCP}
	if err := config.Save(a.configPath, updated); err != nil {
		a.mu.Unlock()
		a.logger.Error("Error guardando el modo de trabajo", "error", err)
		return errors.New("No se pudo guardar la configuración.")
	}
	a.cfg = updated
	a.studyServerState = "No probado"
	active := activeServer(updated)
	a.mu.Unlock()
	if a.studyRepo != nil {
		a.studyRepo.UpdateConfig(active)
	}
	a.logger.Info("Modo de trabajo actualizado", "mode", m.Active)
	return nil
}

// activeServer es el servidor que usa Buscar según el modo guardado.
func activeServer(c config.Config) config.ServerConfig {
	if c.Mode.Active == config.ModeServerV2 {
		return c.Mode.V2
	}
	s := c.StudyAPI
	s.APIKey = ""
	return s
}

// ---------- Destinos de grabación (v2.2) ----------

// RobotSettings es la tarjeta «Robot Epson». Supported indica si este equipo
// puede usar el robot (solo Windows); es de solo lectura.
type RobotSettings struct {
	Enabled       bool   `json:"enabled"`
	Supported     bool   `json:"supported"`
	DiscType      string `json:"discType"`
	Format        string `json:"format"`
	PrintMode     int    `json:"printMode"`
	LabelType     int    `json:"labelType"`
	DefaultCopies int    `json:"defaultCopies"`
}

type DestinationSettings struct {
	Robot  RobotSettings      `json:"robot"`
	Drives config.DriveConfig `json:"drives"`
}

type RobotStatus struct {
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

type DriveStatus struct {
	discdrive.Drive
	Selected   bool   `json:"selected"`
	State      string `json:"state"`
	StateLabel string `json:"stateLabel"`
}

type DestinationStatus struct {
	Robot        RobotStatus   `json:"robot"`
	DrivesOn     bool          `json:"drivesEnabled"`
	Drives       []DriveStatus `json:"drives"`
	DrivesError  string        `json:"drivesError,omitempty"`
	ReadyToBurn  bool          `json:"readyToBurn"`
	ReadyMessage string        `json:"readyMessage,omitempty"`
}

func robotSupported(dev bool) bool { return dev || runtime.GOOS == "windows" }

func (a *App) GetDestinationConfig() DestinationSettings {
	a.mu.RLock()
	defer a.mu.RUnlock()
	e := a.cfg.Epson
	return DestinationSettings{
		Robot:  RobotSettings{Enabled: e.Enabled, Supported: robotSupported(a.devMode), DiscType: e.DiscType, Format: e.Format, PrintMode: e.PrintMode, LabelType: e.LabelType, DefaultCopies: e.DefaultCopies},
		Drives: a.cfg.Drives,
	}
}

// robotAvailability comprueba si TD Bridge está accesible.
func (a *App) robotAvailability(folder string) (bool, string) {
	if !robotSupported(a.devMode) {
		return false, "Disponible solo en Windows."
	}
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		return false, "No se encuentra la carpeta de TD Bridge. Encienda TD Bridge o deshabilite el robot."
	}
	return true, ""
}

func (a *App) listDrives() ([]discdrive.Drive, error) {
	if a.drives == nil {
		return nil, discdrive.ErrUnsupported
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return a.drives.List(ctx)
}

// GetDestinationStatus detecta el estado real de cada destino (unidades y disco insertado).
func (a *App) GetDestinationStatus() DestinationStatus {
	a.mu.RLock()
	epson, drivesCfg := a.cfg.Epson, a.cfg.Drives
	a.mu.RUnlock()
	st := DestinationStatus{DrivesOn: drivesCfg.Enabled, Drives: []DriveStatus{}}
	st.Robot.Enabled = epson.Enabled
	st.Robot.Available, st.Robot.Message = a.robotAvailability(epson.MonitoringFolder)
	selected := map[string]bool{}
	for _, id := range drivesCfg.Selected {
		selected[id] = true
	}
	drives, err := a.listDrives()
	if err != nil {
		st.DrivesError = err.Error()
	}
	readyDrive := false
	for _, d := range drives {
		state, label := d.State()
		ds := DriveStatus{Drive: d, Selected: selected[d.ID], State: state, StateLabel: label}
		if ds.Selected && state == discdrive.StateReady {
			readyDrive = true
		}
		st.Drives = append(st.Drives, ds)
	}
	robotReady := st.Robot.Enabled && st.Robot.Available
	st.ReadyToBurn = robotReady || (drivesCfg.Enabled && readyDrive)
	if !st.ReadyToBurn {
		switch {
		case !epson.Enabled && !drivesCfg.Enabled:
			st.ReadyMessage = "No hay destinos habilitados."
		case drivesCfg.Enabled:
			st.ReadyMessage = "Inserte un disco virgen o encienda TD Bridge."
		default:
			st.ReadyMessage = st.Robot.Message
		}
	}
	return st
}

// SaveDestinationConfig guarda solo si cada destino habilitado está operativo
// y hay al menos uno habilitado (decisión 12). No toca Modo, Etiqueta ni licencia.
func (a *App) SaveDestinationConfig(d DestinationSettings) error {
	r := d.Robot
	epson := config.EpsonConfig{DiscType: r.DiscType, Format: r.Format, PrintMode: r.PrintMode, LabelType: r.LabelType}
	epson.SetDefaults()
	if err := epson.ValidateDisc(); err != nil {
		return err
	}
	if err := epson.ValidatePerformance(); err != nil {
		return err
	}
	if r.DefaultCopies < 1 || r.DefaultCopies > 99 {
		return errors.New("Las copias deben estar entre 1 y 99.")
	}
	if err := d.Drives.Validate(); err != nil {
		return err
	}
	if !r.Enabled && !d.Drives.Enabled {
		return errors.New("Habilite al menos un destino.")
	}
	a.mu.RLock()
	folder := a.cfg.Epson.MonitoringFolder
	a.mu.RUnlock()
	if r.Enabled {
		if ok, msg := a.robotAvailability(folder); !ok {
			return fmt.Errorf("Robot Epson no disponible: %s", msg)
		}
	}
	selected := uniqueIDs(d.Drives.Selected)
	if d.Drives.Enabled {
		if len(selected) == 0 {
			return errors.New("Seleccione al menos una unidad.")
		}
		drives, err := a.listDrives()
		if err != nil {
			return fmt.Errorf("Unidades CD/DVD no disponibles: %s", err.Error())
		}
		connected := false
		for _, drv := range drives {
			for _, id := range selected {
				if drv.ID == id {
					connected = true
				}
			}
		}
		if !connected {
			return errors.New("Unidades CD/DVD no disponibles: ninguna unidad seleccionada está conectada.")
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	updated := a.cfg
	updated.Epson.Enabled = r.Enabled
	updated.Epson.DiscType, updated.Epson.Format = epson.DiscType, epson.Format
	updated.Epson.PrintMode, updated.Epson.LabelType = epson.PrintMode, epson.LabelType
	updated.Epson.DefaultCopies = r.DefaultCopies
	updated.Drives = d.Drives
	updated.Drives.Selected = selected
	if err := config.Save(a.configPath, updated); err != nil {
		a.logger.Error("Error guardando destinos", "error", err)
		return errors.New("No se pudo guardar la configuración de destinos.")
	}
	a.cfg = updated
	if p, ok := a.publisher.(*adapters.TdBridgePublisher); ok {
		next := *p
		next.DiscType, next.Format = epson.DiscType, epson.Format
		next.PrintMode, next.LabelType = epson.PrintMode, epson.LabelType
		next.DefaultCopies = r.DefaultCopies
		a.publisher = &next
	}
	a.logger.Info("Destinos actualizados", "robot", r.Enabled, "drives", d.Drives.Enabled, "selected", len(selected))
	return nil
}

func uniqueIDs(ids []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// ---------- Etiqueta (común) y etiqueta en papel ----------

type LabelSettings struct {
	Label config.DiscLabelConfig  `json:"label"`
	Paper config.PaperLabelConfig `json:"paper"`
	Sizes []config.PaperSize      `json:"sizes"`
}

func (a *App) GetLabelSettings() LabelSettings {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return LabelSettings{Label: a.cfg.DiscLabel, Paper: a.cfg.PaperLabel, Sizes: config.PaperSizes()}
}

// SaveLabelSettings guarda hospital, logo y etiqueta en papel en una sola operación.
func (a *App) SaveLabelSettings(s LabelSettings) error {
	s.Paper.Printer = strings.TrimSpace(s.Paper.Printer)
	if err := s.Paper.Validate(); err != nil {
		return err
	}
	if err := a.SaveDiscLabelConfig(s.Label); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	updated := a.cfg
	updated.PaperLabel = s.Paper
	if err := config.Save(a.configPath, updated); err != nil {
		return errors.New("No se pudo guardar la configuración de la etiqueta.")
	}
	a.cfg = updated
	return nil
}

// ListPrinters devuelve las impresoras del sistema para la etiqueta en papel.
func (a *App) ListPrinters() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	printers, err := discdrive.Printers(ctx)
	if err != nil {
		return nil, errors.New("No se pudo obtener la lista de impresoras.")
	}
	return printers, nil
}
