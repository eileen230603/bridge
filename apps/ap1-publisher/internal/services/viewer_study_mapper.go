package services

import (
	"sort"
	"strconv"
	"strings"

	"github.com/local/dicom-disc-suite/shared/models"
)

// ViewerStudy is the study.json contract consumed by @medicaresoft/dicom-viewer.
// Filesystem paths and Symphony-only fields are intentionally excluded.
type ViewerStudy struct {
	ID               string         `json:"id"`
	StudyInstanceUID string         `json:"studyInstanceUid"`
	PatientName      string         `json:"patientName"`
	PatientID        string         `json:"patientId"`
	PatientBirthDate *string        `json:"patientBirthDate"`
	PatientAge       *string        `json:"patientAge"`
	PatientSex       *string        `json:"patientSex"`
	StudyDate        *string        `json:"studyDate"`
	StudyTime        *string        `json:"studyTime"`
	StudyDescription *string        `json:"studyDescription"`
	Series           []ViewerSeries `json:"series"`
}

type ViewerSeries struct {
	ID                string       `json:"id"`
	SeriesInstanceUID string       `json:"seriesInstanceUid"`
	Modality          string       `json:"modality"`
	Name              string       `json:"name"`
	Files             []ViewerFile `json:"files"`
}

type ViewerFile struct {
	Position    int    `json:"position"`
	InstanceUID string `json:"instanceUid"`
}

func MapSymphonyStudyToViewerStudy(study models.Study) ViewerStudy {
	patientName := strings.TrimSpace(study.SourcePatientName)
	if patientName == "" {
		patientName = study.PatientName
	}
	description := strings.TrimSpace(study.StudyDescription)
	if description == "" {
		for _, series := range study.Series {
			if name := strings.TrimSpace(series.SeriesDescription); name != "" {
				description = name
				break
			}
		}
	}

	result := ViewerStudy{
		ID:               strconv.Itoa(study.StudyID),
		StudyInstanceUID: study.StudyInstanceUID,
		PatientName:      patientName,
		PatientID:        study.PatientID,
		PatientBirthDate: optionalString(study.PatientBirthDate),
		PatientAge:       optionalString(study.PatientAge),
		PatientSex:       optionalString(study.PatientSex),
		StudyDate:        optionalString(strings.ReplaceAll(study.StudyDate, "-", "")),
		StudyDescription: stringPointer(description),
		Series:           make([]ViewerSeries, 0, len(study.Series)),
	}
	for _, series := range study.Series {
		mapped := ViewerSeries{
			ID:                series.SeriesID,
			SeriesInstanceUID: series.SeriesInstanceUID,
			Modality:          series.Modality,
			Name:              series.SeriesDescription,
			Files:             make([]ViewerFile, 0, len(series.Files)),
		}
		for _, file := range series.Files {
			mapped.Files = append(mapped.Files, ViewerFile{Position: file.Position, InstanceUID: file.InstanceUID})
		}
		sort.SliceStable(mapped.Files, func(i, j int) bool {
			return mapped.Files[i].Position < mapped.Files[j].Position
		})
		result.Series = append(result.Series, mapped)
	}
	return result
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func stringPointer(value string) *string { return &value }
