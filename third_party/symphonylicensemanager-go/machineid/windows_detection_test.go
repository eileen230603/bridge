package machineid

import "testing"

func TestWindowsGuestModel(t *testing.T) {
	cases := []struct {
		model   string
		virtual bool
	}{
		{"Victus by HP Gaming Laptop 15-fb0xxx", false},
		{"Surface Laptop 5", false},
		{"ThinkPad T14", false},
		{"Virtual Machine", true},
		{"VMware Virtual Platform", true},
		{"VirtualBox", true},
		{"KVM", true},
		{"Parallels ARM Virtual Machine", true},
		{"QEMU", true},
		{"Xen", true},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			if got := windowsGuestModel(tc.model); got != tc.virtual {
				t.Fatalf("windowsGuestModel(%q) = %v, want %v", tc.model, got, tc.virtual)
			}
		})
	}
}
