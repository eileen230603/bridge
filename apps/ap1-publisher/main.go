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
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app, err := NewApp()
	if err != nil {
		logStartupFailure("initialize application", err)
		return
	}
	defer app.closeHistory()
	err = wails.Run(&options.App{Title: "Symphony Media Export", Width: 1280, Height: 820, MinWidth: 1100, MinHeight: 700, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 9, G: 18, B: 31, A: 1}, OnStartup: app.startup, Bind: []interface{}{app},
		// macOS: barra de título integrada al contenido (semáforos sobre el menú). El arrastre
		// lo da la franja .titlebar-drag del frontend (--wails-draggable).
		Mac: &mac.Options{TitleBar: mac.TitleBarHiddenInset()},
		// Windows: barra nativa (se puede mover/redimensionar) teñida con la paleta de la app.
		Windows: &windows.Options{Theme: windows.SystemDefault, CustomTheme: &windows.ThemeSettings{
			DarkModeTitleBar: windows.RGB(5, 22, 28), DarkModeTitleBarInactive: windows.RGB(8, 28, 35),
			DarkModeTitleText: windows.RGB(238, 246, 248), DarkModeTitleTextInactive: windows.RGB(146, 172, 181),
			DarkModeBorder: windows.RGB(20, 70, 84), DarkModeBorderInactive: windows.RGB(14, 48, 58),
			LightModeTitleBar: windows.RGB(232, 240, 243), LightModeTitleBarInactive: windows.RGB(240, 245, 247),
			LightModeTitleText: windows.RGB(13, 43, 53), LightModeTitleTextInactive: windows.RGB(82, 112, 123),
			LightModeBorder: windows.RGB(160, 196, 206), LightModeBorderInactive: windows.RGB(200, 220, 226),
		}}})
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
