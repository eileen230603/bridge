//go:build darwin

package discdrive

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// System devuelve el detector del sistema operativo actual.
func System() Detector { return drutilDetector{} }

type drutilDetector struct{}

func drutil(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "/usr/bin/drutil", args...).Output()
	return string(out), err
}

func (drutilDetector) List(ctx context.Context) ([]Drive, error) {
	out, err := drutil(ctx, "list")
	if err != nil {
		return nil, fmt.Errorf("drutil list: %w", err)
	}
	entries := parseDrutilList(out)
	drives := make([]Drive, 0, len(entries))
	seen := map[string]int{}
	for _, e := range entries {
		n := strconv.Itoa(e.Index)
		name := strings.TrimSpace(e.Vendor + " " + e.Product)
		id := "mac:" + name
		seen[id]++
		if seen[id] > 1 {
			id = fmt.Sprintf("%s#%d", id, seen[id])
		}
		d := Drive{ID: id, Name: name, Label: "Unidad " + n}
		if info, err := drutil(ctx, "info", "-drive", n); err == nil {
			d.CanWriteCD, d.CanWriteDVD, d.CanWriteBD = parseDrutilInfo(info)
		}
		if status, err := drutil(ctx, "status", "-drive", n); err == nil {
			d.Media, d.Device = parseDrutilStatus(status)
		}
		drives = append(drives, d)
	}
	return drives, nil
}

// Printers lista las impresoras de CUPS.
func Printers(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, "/usr/bin/lpstat", "-e").Output()
	if err != nil {
		return nil, fmt.Errorf("lpstat: %w", err)
	}
	return parsePrinters(string(out)), nil
}
