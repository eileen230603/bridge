package main

// Metadatos de versión de Symphony Media Export (AP1).
//
// La fuente de verdad es el archivo VERSION en la raíz del repositorio; el test
// version_test.go comprueba que Version, VERSION y wails.json coincidan.
// Commit y BuildDate los inyecta `make build-ap1` con -ldflags.
var (
	Version   = "2.0.0"
	Commit    = "dev"
	BuildDate = ""
)

const (
	productName = "Symphony Media Export"
	companyName = "MedicareSoft Ingeniería SRL"
)

// AppInfo se expone a la interfaz (pantalla "Acerca de").
type AppInfo struct {
	Name      string `json:"name"`
	Company   string `json:"company"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	DevMode   bool   `json:"devMode"`
}

func (a *App) GetAppInfo() AppInfo {
	return AppInfo{Name: productName, Company: companyName, Version: Version, Commit: Commit, BuildDate: BuildDate, DevMode: a.devMode}
}
