package main

import (
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"strings"
	"testing"
)

func TestCreateDiscJobRejectsUnlicensedBeforePreparingStudy(t *testing.T) {
	for _, token := range []string{"", "invalid-token"} {
		t.Run(token, func(t *testing.T) {
			app := &App{cfg: config.Config{LicenseToken: token}}
			_, err := app.CreateDiscJob("study")
			if err == nil || !strings.Contains(err.Error(), "licencia activa") {
				t.Fatalf("expected license rejection before accessing study or publisher, got %v", err)
			}
		})
	}
}
