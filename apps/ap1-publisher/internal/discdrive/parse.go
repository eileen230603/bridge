package discdrive

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FormatBytes muestra capacidades como en la interfaz: «4,7 GB», «700 MB».
func FormatBytes(n int64) string {
	if n <= 0 {
		return "0 MB"
	}
	gb := float64(n) / 1e9
	if gb >= 1 {
		return strings.Replace(strconv.FormatFloat(gb, 'f', 1, 64), ".", ",", 1) + " GB"
	}
	return fmt.Sprintf("%d MB", n/1_000_000)
}

// mediaKind deduce CD/DVD/BD a partir del nombre del medio.
func mediaKind(t string) string {
	u := strings.ToUpper(t)
	switch {
	case strings.Contains(u, "BD"), strings.Contains(u, "BLU"):
		return "BD"
	case strings.Contains(u, "DVD"):
		return "DVD"
	case strings.Contains(u, "CD"):
		return "CD"
	}
	return ""
}

func isRewritableType(t string) bool {
	u := strings.ToUpper(t)
	return strings.Contains(u, "RW") || strings.Contains(u, "-RE") || strings.Contains(u, "RAM")
}

// ---------- macOS: drutil ----------

type drutilEntry struct {
	Index   int
	Vendor  string
	Product string
	Bus     string
}

// parseDrutilList interpreta «drutil list», cuyas columnas son de ancho fijo
// y se alinean con el encabezado (el modelo puede tener espacios dobles):
//
//	   Vendor   Product           Rev   Bus       SupportLevel
//	1  HL-DT-ST DVDRW  GX40N      RQ00  USB       Unsupported
func parseDrutilList(out string) []drutilEntry {
	var res []drutilEntry
	cols := map[string]int{}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Vendor") && strings.Contains(line, "Product") {
			for _, name := range []string{"Vendor", "Product", "Rev", "Bus", "SupportLevel"} {
				cols[name] = strings.Index(line, name)
			}
			continue
		}
		t := strings.TrimSpace(line)
		if t == "" || len(cols) == 0 {
			continue
		}
		lead := strings.TrimLeft(line, " ")
		idxEnd := strings.IndexFunc(lead, func(r rune) bool { return r < '0' || r > '9' })
		if idxEnd <= 0 {
			continue
		}
		idx, err := strconv.Atoi(lead[:idxEnd])
		if err != nil {
			continue
		}
		field := func(from, to string) string {
			a, b := cols[from], len(line)
			if to != "" && cols[to] > 0 && cols[to] < len(line) {
				b = cols[to]
			}
			if a < 0 || a >= len(line) || a >= b {
				return ""
			}
			return strings.TrimSpace(line[a:b])
		}
		e := drutilEntry{Index: idx, Vendor: field("Vendor", "Product"), Product: strings.Join(strings.Fields(field("Product", "Rev")), " "), Bus: field("Bus", "SupportLevel")}
		res = append(res, e)
	}
	return res
}

// parseDrutilInfo extrae qué medios graba la unidad (líneas CD-Write, DVD-Write, BD-Write).
func parseDrutilInfo(out string) (cd, dvd, bd bool) {
	for _, line := range strings.Split(out, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		has := strings.TrimSpace(v) != "" && !strings.EqualFold(strings.TrimSpace(v), "none")
		switch strings.TrimSpace(k) {
		case "CD-Write":
			cd = has
		case "DVD-Write":
			dvd = has
		case "BD-Write":
			bd = has
		}
	}
	return
}

var (
	reType   = regexp.MustCompile(`Type:\s*(.+?)(\s{2,}|\s+Name:|$)`)
	reName   = regexp.MustCompile(`Name:\s*(\S+)`)
	reBlocks = regexp.MustCompile(`blocks:\s*(\d+)`)
)

// parseDrutilStatus interpreta «drutil status -drive N».
func parseDrutilStatus(out string) (Media, string) {
	var m Media
	device := ""
	for _, line := range strings.Split(out, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "Type:"):
			if s := reType.FindStringSubmatch(t); s != nil {
				typ := strings.TrimSpace(s[1])
				if !strings.Contains(strings.ToLower(typ), "no media") {
					m.Present = true
					m.Type = typ
					m.Kind = mediaKind(typ)
					m.Rewritable = isRewritableType(typ)
				}
			}
			if s := reName.FindStringSubmatch(t); s != nil {
				device = s[1]
			}
		case strings.HasPrefix(t, "Space Free:"):
			if s := reBlocks.FindStringSubmatch(t); s != nil {
				blocks, _ := strconv.ParseInt(s[1], 10, 64)
				m.CapacityBytes = blocks * 2048
			}
		case strings.HasPrefix(t, "Writability:"):
			m.Blank = strings.Contains(strings.ToLower(t), "blank")
		}
	}
	if !m.Present {
		return Media{}, device
	}
	return m, device
}

// ---------- Windows: IMAPI2 (JSON del script de PowerShell) ----------

type imapiDrive struct {
	ID          string `json:"id"`
	Vendor      string `json:"vendor"`
	Product     string `json:"product"`
	RawPaths    json.RawMessage `json:"paths"`
	WriteTypes  json.RawMessage `json:"writeTypes"`
	Present     bool            `json:"present"`
	MediaType   int             `json:"mediaType"`
	Blank       bool            `json:"blank"`
	FreeSectors int64           `json:"freeSectors"`
}

// IMAPI_MEDIA_PHYSICAL_TYPE → nombre del medio.
var imapiMedia = map[int]string{
	1: "CD-ROM", 2: "CD-R", 3: "CD-RW", 4: "DVD-ROM", 5: "DVD-RAM", 6: "DVD+R", 7: "DVD+RW",
	8: "DVD+R DL", 9: "DVD-R", 10: "DVD-RW", 11: "DVD-R DL", 13: "DVD+RW DL",
	17: "BD-ROM", 18: "BD-R", 19: "BD-RE",
}

// ConvertTo-Json entrega un valor suelto cuando el arreglo tiene un solo elemento.
func jsonList[T any](raw json.RawMessage) []T {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var list []T
	if json.Unmarshal(raw, &list) == nil {
		return list
	}
	var one T
	if json.Unmarshal(raw, &one) == nil {
		return []T{one}
	}
	return nil
}

func parseIMAPI(out []byte) ([]Drive, error) {
	raw := json.RawMessage(strings.TrimSpace(string(out)))
	if len(raw) == 0 {
		return []Drive{}, nil
	}
	items := jsonList[imapiDrive](raw)
	if items == nil && string(raw) != "[]" {
		return nil, fmt.Errorf("respuesta de IMAPI2 no válida")
	}
	drives := make([]Drive, 0, len(items))
	for _, it := range items {
		paths := jsonList[string](it.RawPaths)
		d := Drive{ID: "win:" + it.ID, Name: strings.TrimSpace(it.Vendor + " " + it.Product), Device: it.ID}
		if len(paths) > 0 {
			d.Label = strings.TrimRight(paths[0], `\`)
		} else {
			d.Label = "Unidad"
		}
		for _, t := range jsonList[int](it.WriteTypes) {
			switch mediaKind(imapiMedia[t]) {
			case "CD":
				d.CanWriteCD = true
			case "DVD":
				d.CanWriteDVD = true
			case "BD":
				d.CanWriteBD = true
			}
		}
		if it.Present && it.MediaType != 0 {
			typ := imapiMedia[it.MediaType]
			if typ == "" {
				typ = fmt.Sprintf("Medio %d", it.MediaType)
			}
			d.Media = Media{Present: true, Type: typ, Kind: mediaKind(typ), Blank: it.Blank, Rewritable: isRewritableType(typ), CapacityBytes: it.FreeSectors * 2048}
		}
		drives = append(drives, d)
	}
	return drives, nil
}

// parsePrinters convierte una lista de nombres, uno por línea.
func parsePrinters(out string) []string {
	res := []string{}
	for _, line := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			res = append(res, t)
		}
	}
	return res
}
