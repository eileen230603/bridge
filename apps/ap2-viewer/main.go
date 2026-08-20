package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	if err := wails.Run(&options.App{Title: "Portable DICOM Viewer", Width: 1200, Height: 780, MinWidth: 900, MinHeight: 600, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 4, G: 13, B: 23, A: 1}, OnStartup: app.startup, Bind: []interface{}{app}}); err != nil {
		log.Fatal(err)
	}
}
