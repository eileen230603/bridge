package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app, err := NewApp()
	if err != nil {
		logStartupFailure("initialize application", err)
		return
	}
	err = wails.Run(&options.App{Title: "DICOM Disc Publisher", Width: 1280, Height: 820, MinWidth: 1000, MinHeight: 650, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 9, G: 18, B: 31, A: 1}, OnStartup: app.startup, Bind: []interface{}{app}})
	if err != nil {
		logStartupFailure("run Wails", err)
	}
}

func logStartupFailure(stage string, err error) {
	message := fmt.Sprintf("%s AP1 startup failed during %s: %v\n", time.Now().Format(time.RFC3339), stage, err)
	log.Print(message)

	cacheDir, cacheErr := os.UserCacheDir()
	if cacheErr != nil {
		log.Printf("cannot resolve startup log directory: %v", cacheErr)
		return
	}
	logDir := filepath.Join(cacheDir, "Symphony", "AP1")
	if mkdirErr := os.MkdirAll(logDir, 0o755); mkdirErr != nil {
		log.Printf("cannot create startup log directory %s: %v", logDir, mkdirErr)
		return
	}
	logPath := filepath.Join(logDir, "startup.log")
	file, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if openErr != nil {
		log.Printf("cannot open startup log %s: %v", logPath, openErr)
		return
	}
	defer file.Close()
	if _, writeErr := file.WriteString(message); writeErr != nil {
		log.Printf("cannot write startup log %s: %v", logPath, writeErr)
	}
}
