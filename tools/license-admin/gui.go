package main

import (
	"context"
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"io/fs"
)

//go:embed ui/*
var ui embed.FS

func runGUI(dir string) error {
	admin, err := newAdmin(dir)
	if err != nil {
		admin = &Admin{startupErr: err}
	} else {
		defer admin.db.Close()
	}
	assets, err := fs.Sub(ui, "ui")
	if err != nil {
		return err
	}
	return wails.Run(&options.App{
		OnStartup: func(ctx context.Context) { admin.ctx = ctx },
		Title:     "Symphony · Licencias", Width: 1220, Height: 820, MinWidth: 960, MinHeight: 680,
		BackgroundColour: &options.RGBA{R: 246, G: 247, B: 251, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets}, Bind: []interface{}{admin},
	})
}
