package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amrshadid/go-dicom/dataelem"
	"github.com/amrshadid/go-dicom/dataset"
	"github.com/amrshadid/go-dicom/network"
	"github.com/amrshadid/go-dicom/tag"
)

func TestWritePart10ProducesDICOMFile(t *testing.T) {
	ds := dataset.NewDataset()
	for _, elem := range []*dataelem.DataElement{
		dataelem.NewDataElement(tag.New(0x0008, 0x0016), dataelem.UI, []byte(network.CTImageStorageUID)),
		dataelem.NewDataElement(tag.New(0x0008, 0x0018), dataelem.UI, []byte("1.2.3.4.5")),
		dataelem.NewDataElement(tag.New(0x0020, 0x000D), dataelem.UI, []byte("1.2.3.4")),
	} {
		if err := ds.Add(elem); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "instance.dcm")
	if err := writePart10(path, ds, network.CTImageStorageUID, "1.2.3.4.5", network.ExplicitVRLittleEndianUID, "DICOM_DISC_AP1"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) <= 132 || string(raw[128:132]) != "DICM" {
		t.Fatalf("not a DICOM Part-10 file: %d bytes", len(raw))
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temporary file remains: %v", err)
	}
}

func TestDICOMFormatting(t *testing.T) {
	if got := formatPersonName("PEREZ^MARIA  "); got != "PEREZ MARIA" {
		t.Fatalf("unexpected person name: %q", got)
	}
	if got := formatDICOMDate("20260821"); got != "2026-08-21" {
		t.Fatalf("unexpected date: %q", got)
	}
}
