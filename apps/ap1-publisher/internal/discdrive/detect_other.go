//go:build !darwin && !windows

package discdrive

import "context"

// System devuelve el detector del sistema operativo actual.
func System() Detector { return unsupported{} }

type unsupported struct{}

func (unsupported) List(context.Context) ([]Drive, error) { return nil, ErrUnsupported }

// Printers no está disponible en este sistema operativo.
func Printers(context.Context) ([]string, error) { return nil, ErrUnsupported }
