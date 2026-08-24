package main

import (
	"embed"
	"log"

	"github.com/local/dicom-disc-suite/apps/ap2-viewer/internal/environment"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Usar CDExecutionValidator para producción o DevelopmentEnvironmentValidator para dev
validator := environment.DevelopmentEnvironmentValidator{}
	app := NewApp(validator)

	err := wails.Run(&options.App{
		Title:            "Portable DICOM Viewer",
		Width:            1200,
		Height:           780,
		MinWidth:         900,
		MinHeight:        600,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 4, G: 13, B: 23, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}