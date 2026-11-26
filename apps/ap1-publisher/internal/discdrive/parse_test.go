package discdrive

import "testing"

const drutilListOut = `   Vendor   Product           Rev   Bus       SupportLevel
1  HL-DT-ST DVDRW  GX40N      RQ00  USB       Unsupported
2  PIONEER  BD-RW   BDR-XD08  1.02  USB       Apple Shipping
`

func TestParseDrutilList(t *testing.T) {
	got := parseDrutilList(drutilListOut)
	if len(got) != 2 {
		t.Fatalf("got %+v", got)
	}
	if got[0].Index != 1 || got[0].Vendor != "HL-DT-ST" || got[0].Product != "DVDRW GX40N" || got[0].Bus != "USB" {
		t.Fatalf("first: %+v", got[0])
	}
	if got[1].Index != 2 || got[1].Vendor != "PIONEER" || got[1].Product != "BD-RW BDR-XD08" {
		t.Fatalf("second: %+v", got[1])
	}
	if len(parseDrutilList("   Vendor   Product           Rev   Bus       SupportLevel\n\n")) != 0 {
		t.Fatal("sin unidades debe dar lista vacía")
	}
}

func TestParseDrutilInfo(t *testing.T) {
	cd, dvd, bd := parseDrutilInfo(`   Interconnect: USB
    CD-Write: -R, -RW, BUFE, Test, IndexPts, ISRC
   DVD-Write: -R, -R DL, -RW, +R, +R DL, +RW, BUFE, Test
  Strategies: CD-TAO, CD-SAO, DVD-DAO`)
	if !cd || !dvd || bd {
		t.Fatalf("cd=%v dvd=%v bd=%v", cd, dvd, bd)
	}
}

func TestParseDrutilStatus(t *testing.T) {
	m, dev := parseDrutilStatus(`           Type: DVD-R                Name: /dev/disk4
       Sessions: 0                  Tracks: 0
   Overwritable:   00:00:00         blocks:        0 /   0.00MB /   0.00MiB
     Space Free: 510:03:59        blocks:  2295104 /   4.70GB /   4.38GiB
     Space Used:   00:00:00         blocks:        0 /   0.00MB /   0.00MiB
    Writability: appendable, blank, overwritable`)
	if !m.Present || m.Type != "DVD-R" || m.Kind != "DVD" || !m.Blank || m.Rewritable || m.CapacityBytes != 2295104*2048 || dev != "/dev/disk4" {
		t.Fatalf("media: %+v dev=%q", m, dev)
	}
	empty, _ := parseDrutilStatus(`           Type: No Media Inserted`)
	if empty.Present {
		t.Fatalf("sin disco: %+v", empty)
	}
	rw, _ := parseDrutilStatus("Type: CD-RW   Name: /dev/disk4\nWritability: appendable, blank")
	if !rw.Rewritable || rw.Kind != "CD" {
		t.Fatalf("regrabable: %+v", rw)
	}
}

func TestParseIMAPI(t *testing.T) {
	out := `[{"id":"{abc}","vendor":"ASUS","product":"SDRW-08D2S-U","paths":["D:\\"],"writeTypes":[2,3,6,9,10],"present":true,"mediaType":9,"blank":true,"freeSectors":2298496},{"id":"{def}","vendor":"PIONEER","product":"BDR-XD08","paths":"E:\\","writeTypes":[2,9,18],"present":false,"mediaType":0,"blank":false,"freeSectors":0}]`
	drives, err := parseIMAPI([]byte(out))
	if err != nil || len(drives) != 2 {
		t.Fatalf("%v %+v", err, drives)
	}
	d := drives[0]
	if d.ID != "win:{abc}" || d.Label != "D:" || !d.CanWriteCD || !d.CanWriteDVD || d.CanWriteBD || d.Media.Type != "DVD-R" || !d.Media.Blank {
		t.Fatalf("D: %+v", d)
	}
	if st, _ := d.State(); st != StateReady {
		t.Fatalf("estado D: %s", st)
	}
	e := drives[1]
	if e.Label != "E:" || !e.CanWriteBD || e.Media.Present {
		t.Fatalf("E: %+v", e)
	}
	if st, _ := e.State(); st != StateNoDisc {
		t.Fatalf("estado E: %s", st)
	}
	// Un solo elemento: ConvertTo-Json devuelve un objeto, no un arreglo.
	one, err := parseIMAPI([]byte(`{"id":"{x}","vendor":"V","product":"P","paths":[],"writeTypes":9,"present":false}`))
	if err != nil || len(one) != 1 || !one[0].CanWriteDVD {
		t.Fatalf("uno: %v %+v", err, one)
	}
	none, err := parseIMAPI([]byte(`[]`))
	if err != nil || len(none) != 0 {
		t.Fatalf("ninguno: %v %+v", err, none)
	}
}

func TestDriveStateRejectsUsedOrRewritable(t *testing.T) {
	d := Drive{CanWriteDVD: true, Media: Media{Present: true, Type: "DVD-RW", Kind: "DVD", Blank: true, Rewritable: true}}
	if st, _ := d.State(); st != StateNotBlank {
		t.Fatalf("regrabable: %s", st)
	}
	d.Media = Media{Present: true, Type: "DVD-R", Kind: "DVD", Blank: false}
	if st, _ := d.State(); st != StateNotBlank {
		t.Fatalf("usado: %s", st)
	}
	d.Media = Media{Present: true, Type: "BD-R", Kind: "BD", Blank: true}
	if st, _ := d.State(); st != StateNoWrite {
		t.Fatalf("no compatible: %s", st)
	}
}

func TestFormatBytes(t *testing.T) {
	if got := FormatBytes(4_700_000_000); got != "4,7 GB" {
		t.Fatal(got)
	}
	if got := FormatBytes(700_000_000); got != "700 MB" {
		t.Fatal(got)
	}
}
