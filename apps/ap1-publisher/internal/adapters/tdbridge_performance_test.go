package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintPerformanceOverrides(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mode, label int
		valid       bool
	}{
		{"inherit", 0, 0, true},
		{"standard fast", 2, 1, true},
		{"high quality fast", 2, 2, true},
		{"glossy quality", 1, 3, true},
		{"unknown surface fast", 2, 0, false},
		{"glossy fast", 2, 3, false},
		{"glossy inherited mode", 0, 3, false},
		{"AP-only mode", 3, 1, false},
		{"invalid label", 1, 9, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			job := validDiscJob(t, root)
			staging := filepath.Join(root, "staging")
			p := TdBridgePublisher{StagingDirectory: staging, PrintMode: tc.mode, LabelType: tc.label}
			path, err := p.CreateJob(context.Background(), job)
			if !tc.valid {
				if err == nil {
					t.Fatal("invalid performance settings accepted")
				}
				entries, _ := os.ReadDir(staging)
				if len(entries) != 0 {
					t.Fatal("invalid job was staged")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			text := string(raw)
			if strings.Contains(text, "WRITING_SPEED=") || strings.Contains(text, "COMPARE=") {
				t.Fatal("performance override must preserve automatic write speed and comparison behavior")
			}
			if tc.mode == 0 {
				if strings.Contains(text, "PRINT_MODE=") || strings.Contains(text, "LABEL_TYPE=") {
					t.Fatal("legacy configuration must inherit Epson settings")
				}
			} else {
				if !strings.Contains(text, "PRINT_MODE=") || !strings.Contains(text, "LABEL_TYPE=") {
					t.Fatalf("missing print overrides: %s", text)
				}
				if tc.mode == 2 && !strings.Contains(text, "PRINT_MODE=2\r\n") {
					t.Fatalf("wrong fast mode: %s", text)
				}
			}
		})
	}
}
