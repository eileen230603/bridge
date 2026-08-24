package main

import (
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"log"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}
	err = wails.Run(&options.App{Title: "DICOM Disc Publisher", Width: 1280, Height: 820, MinWidth: 1000, MinHeight: 650, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 9, G: 18, B: 31, A: 1}, OnStartup: app.startup, Bind: []interface{}{app}})
	if err != nil {
		log.Fatal(err)
	}
}
