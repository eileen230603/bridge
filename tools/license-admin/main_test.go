package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/symphonylicensemanager/go/license"
	"strings"
	"testing"
	"time"
)

func TestIssue(t *testing.T) {
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	id := strings.Repeat("ab", 32)
	for _, expires := range []string{"permanente", "", "2026-09-10"} {
		token, err := issue(key, strings.ToUpper(id), expires, now)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			t.Fatal(err)
		}
		payload, sig := raw[:len(raw)-64], raw[len(raw)-64:]
		if !ed25519.Verify(pub, payload, sig) {
			t.Fatal("invalid signature")
		}
		var l license.License
		if err := json.Unmarshal(payload, &l); err != nil {
			t.Fatal(err)
		}
		if l.MachineID != id || l.Product != "symphony-ap1" {
			t.Fatalf("unexpected license: %+v", l)
		}
		if expires == "2026-09-10" {
			if l.Indefinite || l.ExpiresAt != time.Date(2026, 9, 10, 23, 59, 59, 999999999, time.UTC) {
				t.Fatalf("wrong expiry: %+v", l)
			}
		} else if !l.Indefinite {
			t.Fatal("expected permanent")
		}
	}
	for _, expires := range []string{"yesterday", "2026-09-08"} {
		if _, err := issue(key, id, expires, now); err == nil {
			t.Fatal("invalid expiry accepted")
		}
	}
	if _, err := issue(key, "invalid", "", now); err == nil {
		t.Fatal("invalid ID accepted")
	}
}
func TestInitializePreservesExistingKey(t *testing.T) {
	dir := t.TempDir()
	if err := initialize(dir); err != nil {
		t.Fatal(err)
	}
	before, err := loadKey(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := initialize(dir); err == nil {
		t.Fatal("existing key overwritten")
	}
	after, err := loadKey(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !before.Equal(after) {
		t.Fatal("key changed")
	}
}

// Local smoke test: the administrator key is intentionally absent in CI and Git.
func TestConfiguredIssuerCompatibility(t *testing.T) {
	key, err := loadKey("../../license-admin-private")
	if err != nil {
		t.Skip("administrator key is not installed")
	}
	id := strings.Repeat("ab", 32)
	token, err := issue(key, id, "permanente", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := license.Verify(token, id); err != nil {
		t.Fatalf("issuer and application keys do not match: %v", err)
	}
}
