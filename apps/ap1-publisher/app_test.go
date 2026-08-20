package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
)

func TestSelectPublisher(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tests := []struct {
		name string
		want any
	}{
		{"mock", &adapters.MockEpsonPublisher{}},
		{"tdbridge", &adapters.TdBridgePublisher{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectPublisher(config.Config{Publisher: tt.name}, logger)
			if err != nil { t.Fatal(err) }
			switch tt.want.(type) {
			case *adapters.MockEpsonPublisher:
				if _, ok := got.(*adapters.MockEpsonPublisher); !ok { t.Fatalf("got %T", got) }
			case *adapters.TdBridgePublisher:
				if _, ok := got.(*adapters.TdBridgePublisher); !ok { t.Fatalf("got %T", got) }
			}
		})
	}
}
