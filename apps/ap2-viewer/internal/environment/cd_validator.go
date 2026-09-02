//go:build !windows

package environment

type CDExecutionValidator struct{}

func (CDExecutionValidator) Validate() error {
	// En macOS/Linux u otros SOs, permitimos la ejecución para desarrollo/pruebas
	return nil
}