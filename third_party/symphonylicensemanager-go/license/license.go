package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/symphonylicensemanager/go/machineid"
)

var (
	ErrInvalidSignature = errors.New("license: invalid signature")
	ErrExpired          = errors.New("license: expired")
	ErrMachineMismatch  = errors.New("license: machine mismatch")
	ErrVirtualMachine   = errors.New("license: cannot activate on virtual machines")
)

const publicKeyB64 = "eD6aQZgkVTwUgKhZ1LFLe5r6Gnk996MEuqrTZZGDc5I"

var publicKey = mustPublicKey()

func mustPublicKey() ed25519.PublicKey {
	k, err := base64.RawURLEncoding.DecodeString(publicKeyB64)
	if err != nil || len(k) != ed25519.PublicKeySize {
		panic("license: invalid embedded public key")
	}
	return ed25519.PublicKey(k)
}

type License struct {
	Product    string    `json:"product"`
	MachineID  string    `json:"machine_id"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Indefinite bool      `json:"indefinite,omitempty"`
}

func Verify(token string, machineID string) (License, error) {
	var l License

	if machineid.IsVM() {
		return l, ErrVirtualMachine
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return l, err
	}

	if len(raw) < ed25519.SignatureSize {
		return l, ErrInvalidSignature
	}

	payload := raw[:len(raw)-ed25519.SignatureSize]
	sig := raw[len(raw)-ed25519.SignatureSize:]

	if !ed25519.Verify(publicKey, payload, sig) {
		return l, ErrInvalidSignature
	}

	if err := json.Unmarshal(payload, &l); err != nil {
		return l, err
	}

	if l.MachineID != machineID {
		return l, ErrMachineMismatch
	}

	if !l.Indefinite && time.Now().After(l.ExpiresAt) {
		return l, ErrExpired
	}

	return l, nil
}
