package machineid

import (
	"testing"
)

func TestIDDeterministic(t *testing.T) {
	if IsVM() {
		t.Skip("running in a virtual machine")
	}

	a, err := ID()
	if err != nil {
		t.Fatal(err)
	}
	b, err := ID()
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("ID not deterministic: %s != %s", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("unexpected ID length: %d", len(a))
	}
}
