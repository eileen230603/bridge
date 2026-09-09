package license

import (
	"errors"
	"testing"

	"github.com/symphonylicensemanager/go/machineid"
)

const (
	indefiniteToken = "eyJwcm9kdWN0IjoidGVzdCIsIm1hY2hpbmVfaWQiOiJ0ZXN0LW1hY2hpbmUiLCJpc3N1ZWRfYXQiOiIyMDI2LTA5LTAzVDE0OjQ5OjU0Ljc5NTM4N1oiLCJleHBpcmVzX2F0IjoiMjAyNy0wOS0wM1QxNDo0OTo1NC43OTUzODdaIiwiaW5kZWZpbml0ZSI6dHJ1ZX0X4utM4FxdAvv2sA02KweNTN_omJztiTWef3PJ1vv0Jp8DvbsKG3DoXYJXRjkU8bhIZLBsGKfwTCNcBdwdfsML"
	expiredToken    = "eyJwcm9kdWN0IjoidGVzdCIsIm1hY2hpbmVfaWQiOiJ0ZXN0LW1hY2hpbmUiLCJpc3N1ZWRfYXQiOiIyMDI2LTA5LTAzVDE0OjQ5OjU0LjQwNDc0OVoiLCJleHBpcmVzX2F0IjoiMjAyNi0wOC0wNFQxNDo0OTo1NC40MDQ3NDlaIn1FbSbDrYgrRKDSq1ecbDqXkJh6xb2p9B9FCokPTRb48L3o6Ll664AEs2Jg-BbaEcnMrUgFFUCBOHOoWP0qfj4B"
)

func skipIfVM(t *testing.T) {
	if machineid.IsVM() {
		t.Skip("running in a virtual machine")
	}
}

func TestVerifyIndefinite(t *testing.T) {
	skipIfVM(t)

	l, err := Verify(indefiniteToken, "test-machine")
	if err != nil {
		t.Fatal(err)
	}
	if !l.Indefinite {
		t.Fatal("expected indefinite license")
	}
	if l.Product != "test" {
		t.Fatalf("unexpected product: %s", l.Product)
	}
}

func TestVerifyMachineMismatch(t *testing.T) {
	skipIfVM(t)

	_, err := Verify(indefiniteToken, "other-machine")
	if !errors.Is(err, ErrMachineMismatch) {
		t.Fatalf("expected ErrMachineMismatch, got %v", err)
	}
}

func TestVerifyExpired(t *testing.T) {
	skipIfVM(t)

	_, err := Verify(expiredToken, "test-machine")
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	skipIfVM(t)

	tampered := indefiniteToken[:len(indefiniteToken)-1] + "A"
	_, err := Verify(tampered, "test-machine")
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}
