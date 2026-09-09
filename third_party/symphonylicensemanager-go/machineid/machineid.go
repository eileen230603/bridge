package machineid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	ErrVirtualMachine = fmt.Errorf("machineid: cannot activate on virtual machines")
	vmOnce            sync.Once
	vmVal             bool
)

func ID() (string, error) {
	if IsVM() {
		return "", ErrVirtualMachine
	}

	raw, err := hardwareID()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte("symphony|" + raw))
	return hex.EncodeToString(sum[:]), nil
}

func IsVM() bool {
	vmOnce.Do(func() { vmVal = isVM() })
	return vmVal
}

func isVM() bool {
	switch runtime.GOOS {
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "kern.hv_vmm_present").Output(); err == nil && strings.TrimSpace(string(out)) == "1" {
			return true
		}
		if out, err := exec.Command("sysctl", "-n", "hw.model").Output(); err == nil {
			model := strings.ToLower(string(out))
			for _, m := range []string{"vmware", "parallels", "virtual", "qemu", "innotek", "kvm"} {
				if strings.Contains(model, m) {
					return true
				}
			}
		}
	case "linux":
		if b, err := os.ReadFile("/sys/hypervisor/type"); err == nil && strings.TrimSpace(string(b)) != "" {
			return true
		}
		if b, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(line, "flags") && strings.Contains(line, " hypervisor ") {
					return true
				}
			}
		}
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "$cs = Get-CimInstance Win32_ComputerSystem; Write-Output ($cs.HypervisorPresent.ToString() + '|' + $cs.Model)").Output()
		if err != nil {
			return false
		}
		fields := strings.SplitN(strings.TrimSpace(string(out)), "|", 2)
		if len(fields) == 2 {
			return windowsGuestModel(fields[1])
		}
	}
	return false
}

func hardwareID() (string, error) {
	var raw string

	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sh", "-c", "ioreg -rd1 -c IOPlatformExpertDevice | grep IOPlatformUUID").Output()
		if err != nil {
			return "", err
		}
		fields := strings.Fields(string(out))
		if len(fields) < 3 {
			return "", fmt.Errorf("machineid: unable to read platform UUID")
		}
		raw = strings.Trim(fields[2], `"`)
	case "linux":
		b, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
		if err != nil {
			b, err = os.ReadFile("/etc/machine-id")
			if err != nil {
				return "", err
			}
		}
		raw = strings.TrimSpace(string(b))
		if raw == "" || raw == "None" || raw == "00000000-0000-0000-0000-000000000000" {
			b, err := os.ReadFile("/etc/machine-id")
			if err != nil {
				return "", err
			}
			raw = strings.TrimSpace(string(b))
		}
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_ComputerSystemProduct).UUID").Output()
		if err != nil {
			return "", err
		}
		raw = strings.TrimSpace(string(out))
	default:
		return "", fmt.Errorf("machineid: unsupported OS: %s", runtime.GOOS)
	}

	if raw == "" {
		return "", fmt.Errorf("machineid: unable to determine machine identifier")
	}

	return raw, nil
}

// HypervisorPresent is also true on physical Windows hosts using Hyper-V or VBS.
// Only guest model identifiers are evidence that Windows itself runs in a VM.
func windowsGuestModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	for _, marker := range []string{"vmware", "virtualbox", "virtual", "qemu", "kvm", "xen", "parallels", "innotek"} {
		if strings.Contains(model, marker) {
			return true
		}
	}
	return false
}
