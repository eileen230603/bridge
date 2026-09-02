//go:build windows

package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const DRIVE_CDROM = 5

type CDExecutionValidator struct{}

func (CDExecutionValidator) Validate() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("no se pudo determinar la ubicación del ejecutable: %w", err)
	}

	vol := filepath.VolumeName(exe)
	if vol == "" {
		return fmt.Errorf("no se pudo determinar la unidad de ejecución")
	}
	volRoot := vol + `\`

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDriveType := kernel32.NewProc("GetDriveTypeW")

	volPtr, err := syscall.UTF16PtrFromString(volRoot)
	if err != nil {
		return fmt.Errorf("error al procesar la ruta de la unidad: %w", err)
	}

	driveType, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(volPtr)))

	if driveType != DRIVE_CDROM {
		return fmt.Errorf("ejecución bloqueada: el visor solo puede ejecutarse directamente desde el CD/DVD original (Unidad detectada: %s)", volRoot)
	}

	return nil
}