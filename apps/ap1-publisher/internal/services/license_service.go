package services

import (
	"fmt"

	"github.com/symphonylicensemanager/go/license"
	"github.com/symphonylicensemanager/go/machineid"
)

type LicenseDetails struct {
	Subject string `json:"subject,omitempty"`
	Expires string `json:"expires,omitempty"`
}

type LicenseStatus struct {
	IsValid   bool            `json:"isValid"`
	MachineID string          `json:"machineId"`
	IsVM      bool            `json:"isVm"`
	Details   *LicenseDetails `json:"details,omitempty"`
	Error     string          `json:"error,omitempty"`
}

type LicenseService struct{}

func NewLicenseService() *LicenseService {
	return &LicenseService{}
}

// GetMachineID obtiene el identificador único de hardware del equipo
func (s *LicenseService) GetMachineID() (string, error) {
	id, err := machineid.ID()
	if err != nil {
		return "", fmt.Errorf("error al obtener Machine ID: %w", err)
	}
	return id, nil
}

// ValidateLicense comprueba la validez del token contra el Machine ID
func (s *LicenseService) ValidateLicense(token string) LicenseStatus {
	mID, err := machineid.ID()
	if err != nil {
		return LicenseStatus{
			IsValid: false,
			IsVM:    machineid.IsVM(),
			Error:   fmt.Sprintf("Error obteniendo Machine ID: %v", err),
		}
	}

	lic, err := license.Verify(token, mID)
	if err != nil {
		return LicenseStatus{
			IsValid:   false,
			MachineID: mID,
			IsVM:      machineid.IsVM(),
			Error:     err.Error(),
		}
	}

	// Mapear los campos de lic que existan o convertir de forma segura
	details := &LicenseDetails{
		Subject: fmt.Sprintf("%v", lic), // O usa lic.ExpiresAt / lic.IssuedTo según tu paquete
	}

	return LicenseStatus{
		IsValid:   true,
		MachineID: mID,
		IsVM:      machineid.IsVM(),
		Details:   details,
	}
}