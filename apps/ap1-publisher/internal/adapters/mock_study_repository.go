package adapters

import (
	"context"
	"fmt"
	"github.com/local/dicom-disc-suite/shared/models"
	"os"
	"path/filepath"
	"time"
)

type MockStudyRepository struct{}

func (MockStudyRepository) SearchStudies(_ context.Context, from, to time.Time) ([]models.Study, error) {
	studies := []models.Study{
		{StudyInstanceUID: "1.2.840.10008.20260820.1", PatientID: "P-1001", PatientName: "EILEEN BALDERRAMA", StudyDescription: "Tórax", StudyDate: "2026-08-20", Modality: "CT", SeriesCount: 3, InstanceCount: 295},
		{StudyInstanceUID: "1.2.840.10008.20260820.2", PatientID: "P-1002", PatientName: "Juan López", StudyDescription: "Cerebro", StudyDate: "2026-08-20", Modality: "MR", SeriesCount: 4, InstanceCount: 182},
		{StudyInstanceUID: "1.2.840.10008.20260819.3", PatientID: "P-1003", PatientName: "Ana Torres", StudyDescription: "Abdomen", StudyDate: "2026-08-19", Modality: "CT", SeriesCount: 2, InstanceCount: 146},
		{StudyInstanceUID: "1.2.840.10008.20260818.4", PatientID: "P-1004", PatientName: "Carlos Rojas", StudyDescription: "Rodilla derecha", StudyDate: "2026-08-18", Modality: "MR", SeriesCount: 3, InstanceCount: 96},
		{StudyInstanceUID: "1.2.840.10008.20260817.5", PatientID: "P-1005", PatientName: "Elena Vargas", StudyDescription: "Tórax PA", StudyDate: "2026-08-17", Modality: "CR", SeriesCount: 1, InstanceCount: 2}}
	var out []models.Study
	for _, s := range studies {
		d, _ := time.Parse("2006-01-02", s.StudyDate)
		if !d.Before(from) && !d.After(to) {
			s.Series = []models.Series{{SeriesInstanceUID: s.StudyInstanceUID + ".1", SeriesNumber: 1, SeriesDescription: "AXIAL", Modality: s.Modality, InstanceCount: s.InstanceCount}}
			out = append(out, s)
		}
	}
	return out, nil
}
func (MockStudyRepository) RetrieveStudy(_ context.Context, uid, destination string) error {
	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}
	for i := 1; i <= 3; i++ {
		content := fmt.Sprintf("PLACEHOLDER ONLY - NOT A DICOM FILE\nStudyInstanceUID=%s\n", uid)
		if err := os.WriteFile(filepath.Join(destination, fmt.Sprintf("mock%03d.dcm", i)), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}
