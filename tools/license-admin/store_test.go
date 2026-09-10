package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStorePersistsAndImports(t *testing.T) {
	dir := t.TempDir()
	if err := initialize(dir); err != nil {
		t.Fatal(err)
	}
	key, err := loadKey(dir)
	if err != nil {
		t.Fatal(err)
	}
	id := strings.Repeat("ab", 32)
	token, err := issue(key, id, "permanente", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "licencia-anterior.txt"), []byte(token), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := newAdmin(dir)
	if err != nil {
		t.Fatal(err)
	}
	row, err := a.CreateLicense(IssueRequest{Client: "Clínica O'Connor", MachineID: id, Permanent: true})
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != "permanent" || row.ID == 0 {
		t.Fatalf("bad row: %+v", row)
	}
	if _, err := decodeIssued(row.Token, key.Public().(ed25519.PublicKey)); err != nil {
		t.Fatal(err)
	}
	a.db.Close()
	b, err := newAdmin(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.db.Close()
	rows, err := b.ListLicenses()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Client != "Clínica O'Connor" {
		t.Fatalf("persistence/import duplicate failure: %+v", rows)
	}
	bytes, err := os.ReadFile(filepath.Join(dir, "licenses.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(bytes), "SQLite format 3") {
		t.Fatal("not a SQLite database")
	}
	if _, err := b.CreateLicense(IssueRequest{Client: "Test", MachineID: id}); err == nil {
		t.Fatal("missing expiry accepted")
	}
	rows, err = b.ListLicenses()
	if err != nil || len(rows) != 2 {
		t.Fatal("invalid request changed database")
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 1
	if _, err := decodeIssued(base64.RawURLEncoding.EncodeToString(raw), key.Public().(ed25519.PublicKey)); err == nil {
		t.Fatal("tampered import accepted")
	}
}
func TestStatusBoundaries(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name      string
		delta     time.Duration
		permanent bool
		want      string
	}{
		{"permanent", 0, true, "permanent"}, {"expired", -time.Second, false, "expired"}, {"exact expiry", 0, false, "expired"}, {"one second", time.Second, false, "soon"}, {"30 days", 30 * 24 * time.Hour, false, "soon"}, {"beyond 30", 30*24*time.Hour + time.Second, false, "active"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := LicenseRow{Permanent: tc.permanent, ExpiresAt: now.Add(tc.delta).Format(time.RFC3339Nano)}
			if err := statusAt(&row, now); err != nil {
				t.Fatal(err)
			}
			if row.Status != tc.want {
				t.Fatalf("got %s want %s", row.Status, tc.want)
			}
		})
	}
}
