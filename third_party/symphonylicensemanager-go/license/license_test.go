package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/symphonylicensemanager/go/machineid"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	if machineid.IsVM() {
		t.Skip("running in a virtual machine")
	}
	pub, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	original := publicKey
	publicKey = pub
	t.Cleanup(func() { publicKey = original })
	cases := []struct {
		name       string
		id         string
		indefinite bool
		expires    time.Time
		tamper     bool
		want       error
	}{
		{"permanent", "machine", true, time.Time{}, false, nil},
		{"dated", "machine", false, time.Now().Add(time.Hour), false, nil},
		{"different machine", "other", true, time.Time{}, false, ErrMachineMismatch},
		{"expired", "machine", false, time.Now().Add(-time.Hour), false, ErrExpired},
		{"tampered", "machine", true, time.Time{}, true, ErrInvalidSignature},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(License{Product: "symphony-ap1", MachineID: "machine", Indefinite: tc.indefinite, ExpiresAt: tc.expires})
			if err != nil {
				t.Fatal(err)
			}
			raw := append(payload, ed25519.Sign(private, payload)...)
			if tc.tamper {
				raw[len(raw)-1] ^= 1
			}
			got, err := Verify(base64.RawURLEncoding.EncodeToString(raw), tc.id)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if err == nil && got.Product != "symphony-ap1" {
				t.Fatalf("unexpected license: %+v", got)
			}
		})
	}
	_, otherKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"machine_id":"machine","indefinite":true}`)
	token := base64.RawURLEncoding.EncodeToString(append(payload, ed25519.Sign(otherKey, payload)...))
	if _, err := Verify(token, "machine"); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("foreign issuer accepted: %v", err)
	}
}
