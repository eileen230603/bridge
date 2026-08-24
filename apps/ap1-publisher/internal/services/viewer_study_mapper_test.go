package services

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/local/dicom-disc-suite/shared/models"
)

func TestMapSymphonyStudyToViewerStudyComplete(t *testing.T) {
	study := models.Study{
		StudyID: 15491, StudyInstanceUID: "study-uid", PatientID: "patient-42",
		PatientName: "PEREZ MARIA", SourcePatientName: "PEREZ^MARIA",
		PatientBirthDate: "19750410", PatientAge: "050Y", PatientSex: "F",
		StudyDate: "2025-12-31", StudyDescription: "PELVIS",
		Series: []models.Series{
			{SeriesID: "11", SeriesInstanceUID: "series-a", Modality: "MR", SeriesDescription: "T2", Files: []models.StudyFile{
				{Position: 3, InstanceUID: "instance-a"}, {Position: 1, InstanceUID: "instance-b"}, {Position: 3, InstanceUID: "instance-c"},
			}},
			{SeriesID: "12", SeriesInstanceUID: "series-b", Modality: "CT", SeriesDescription: "LOCALIZER", Files: []models.StudyFile{}},
		},
	}

	got := MapSymphonyStudyToViewerStudy(study)
	if got.ID != "15491" || got.StudyInstanceUID != "study-uid" || got.PatientName != "PEREZ^MARIA" || got.PatientID != "patient-42" {
		t.Fatalf("unexpected identity mapping: %+v", got)
	}
	assertStringPointer(t, "patientBirthDate", got.PatientBirthDate, "19750410")
	assertStringPointer(t, "patientAge", got.PatientAge, "050Y")
	assertStringPointer(t, "patientSex", got.PatientSex, "F")
	assertStringPointer(t, "studyDate", got.StudyDate, "20251231")
	assertStringPointer(t, "studyDescription", got.StudyDescription, "PELVIS")
	if got.StudyTime != nil {
		t.Fatalf("studyTime must be null when Symphony does not provide it: %q", *got.StudyTime)
	}
	if len(got.Series) != 2 || got.Series[0].ID != "11" || got.Series[0].SeriesInstanceUID != "series-a" || got.Series[0].Modality != "MR" || got.Series[0].Name != "T2" {
		t.Fatalf("unexpected series mapping: %+v", got.Series)
	}
	wantOrder := []string{"instance-b", "instance-a", "instance-c"}
	actualOrder := make([]string, 0, len(got.Series[0].Files))
	for _, file := range got.Series[0].Files {
		actualOrder = append(actualOrder, file.InstanceUID)
	}
	if !reflect.DeepEqual(actualOrder, wantOrder) {
		t.Fatalf("files are not stably sorted by POS: got %v want %v", actualOrder, wantOrder)
	}
}

func TestMapSymphonyStudyToViewerStudyMissingOptionalFieldsAndFallbacks(t *testing.T) {
	got := MapSymphonyStudyToViewerStudy(models.Study{
		StudyID: 7, PatientName: "VISIBLE NAME",
		Series: []models.Series{{SeriesID: "1", SeriesInstanceUID: "series", SeriesDescription: "FIRST SERIES"}},
	})
	if got.PatientName != "VISIBLE NAME" {
		t.Fatalf("patient name fallback failed: %q", got.PatientName)
	}
	if got.PatientBirthDate != nil || got.PatientAge != nil || got.PatientSex != nil || got.StudyDate != nil || got.StudyTime != nil {
		t.Fatalf("missing optional values must be null: %+v", got)
	}
	assertStringPointer(t, "studyDescription", got.StudyDescription, "FIRST SERIES")
}

func TestViewerStudyJSONUsesExactContractAndExcludesPaths(t *testing.T) {
	manifest := MapSymphonyStudyToViewerStudy(models.Study{
		StudyID: 9, StudyInstanceUID: "study", SourcePatientName: "DOE^JANE", PatientID: "p-9",
		Series: []models.Series{{SeriesID: "2", SeriesInstanceUID: "series", Modality: "MR", SeriesDescription: "T1", Files: []models.StudyFile{{Position: 2, InstanceUID: "instance", AE: `C:\\internal\\path`}}}},
	})
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	wantKeys := []string{"id", "studyInstanceUid", "patientName", "patientId", "patientBirthDate", "patientAge", "patientSex", "studyDate", "studyTime", "studyDescription", "series"}
	if len(top) != len(wantKeys) {
		t.Fatalf("unexpected top-level JSON: %s", raw)
	}
	for _, key := range wantKeys {
		if _, ok := top[key]; !ok {
			t.Fatalf("missing key %q in %s", key, raw)
		}
	}
	if string(top["patientAge"]) != "null" || string(top["studyTime"]) != "null" {
		t.Fatalf("optional absent values must serialize as null: %s", raw)
	}
	if containsAny(string(raw), []string{"studyInstanceUID", "seriesInstanceUID", `"insUid":`, `"pos":`, `"ae":`, `C:\\internal\\path`}) {
		t.Fatalf("legacy or internal data leaked into study.json: %s", raw)
	}
}

func assertStringPointer(t *testing.T, field string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("%s = %v, want %q", field, got, want)
	}
}

func containsAny(value string, fragments []string) bool {
	for _, fragment := range fragments {
		if len(fragment) > 0 && contains(value, fragment) {
			return true
		}
	}
	return false
}

func contains(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
