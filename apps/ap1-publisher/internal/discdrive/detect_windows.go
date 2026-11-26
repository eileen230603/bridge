//go:build windows

package discdrive

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"syscall"
	"unicode/utf16"
)

// System devuelve el detector del sistema operativo actual.
func System() Detector { return imapiDetector{} }

type imapiDetector struct{}

// listScript recorre las grabadoras con IMAPI2 y devuelve JSON. Las propiedades
// del medio lanzan una excepción cuando no hay disco: se tratan como «sin disco».
const listScript = `
$ErrorActionPreference = 'Stop'
$out = @()
$master = New-Object -ComObject IMAPI2.MsftDiscMaster2
foreach ($id in $master) {
  $r = New-Object -ComObject IMAPI2.MsftDiscRecorder2
  $r.InitializeDiscRecorder($id)
  $d = [ordered]@{ id = $id; vendor = "$($r.VendorId)".Trim(); product = "$($r.ProductId)".Trim(); paths = @($r.VolumePathNames); writeTypes = @(); present = $false; mediaType = 0; blank = $false; freeSectors = 0 }
  $f = New-Object -ComObject IMAPI2.MsftDiscFormat2Data
  try { $f.Recorder = $r; $d.writeTypes = @($f.SupportedMediaTypes | ForEach-Object { [int]$_ }) } catch {}
  try {
    $d.mediaType = [int]$f.CurrentPhysicalMediaType
    $d.present = $true
    $d.blank = [bool]$f.MediaHeuristicallyBlank
    $d.freeSectors = [int64]$f.FreeSectorsOnMedia
  } catch {}
  $out += [pscustomobject]$d
}
ConvertTo-Json -InputObject @($out) -Compress -Depth 4
`

const printersScript = `Get-CimInstance Win32_Printer | ForEach-Object { $_.Name }`

func powershell(ctx context.Context, script string) ([]byte, error) {
	u := utf16.Encode([]rune(script))
	b := make([]byte, len(u)*2)
	for i, c := range u {
		b[i*2], b[i*2+1] = byte(c), byte(c>>8)
	}
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", base64.StdEncoding.EncodeToString(b))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	return cmd.Output()
}

func (imapiDetector) List(ctx context.Context) ([]Drive, error) {
	out, err := powershell(ctx, listScript)
	if err != nil {
		return nil, fmt.Errorf("IMAPI2: %w", err)
	}
	return parseIMAPI(out)
}

// Printers lista las impresoras instaladas en Windows.
func Printers(ctx context.Context) ([]string, error) {
	out, err := powershell(ctx, printersScript)
	if err != nil {
		return nil, fmt.Errorf("impresoras: %w", err)
	}
	return parsePrinters(string(out)), nil
}
