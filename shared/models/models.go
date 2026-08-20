package models

import "time"

type Patient struct {
	PatientID   string `json:"patientId"`
	PatientName string `json:"patientName"`
	BirthDate   string `json:"birthDate"`
	Sex         string `json:"sex"`
}
type Study struct {
	StudyInstanceUID string   `json:"studyInstanceUID"`
	PatientID        string   `json:"patientId"`
	PatientName      string   `json:"patientName"`
	StudyDescription string   `json:"studyDescription"`
	StudyDate        string   `json:"studyDate"`
	Modality         string   `json:"modality"`
	SeriesCount      int      `json:"seriesCount"`
	InstanceCount    int      `json:"instanceCount"`
	Series           []Series `json:"series,omitempty"`
}
type Series struct {
	SeriesInstanceUID string `json:"seriesInstanceUID"`
	SeriesNumber      int    `json:"seriesNumber"`
	SeriesDescription string `json:"description"`
	Modality          string `json:"modality"`
	InstanceCount     int    `json:"instanceCount"`
}
type JobStatus string

const (
	Pending        JobStatus = "Pending"
	Downloading    JobStatus = "Downloading"
	Preparing      JobStatus = "Preparing"
	Ready          JobStatus = "Ready"
	QueuedForEpson JobStatus = "QueuedForEpson"
	Publishing     JobStatus = "Publishing"
	Completed      JobStatus = "Completed"
	Failed         JobStatus = "Failed"
)

type DiscJob struct {
	ID               string    `json:"id"`
	StudyInstanceUID string    `json:"studyInstanceUID"`
	PatientID        string    `json:"patientId"`
	PatientName      string    `json:"patientName,omitempty"`
	Status           JobStatus `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	TempPath         string    `json:"tempPath"`
	DataPath         string    `json:"dataPath"`
	ViewerPath       string    `json:"viewerPath"`
	ManifestPath     string    `json:"manifestPath"`
	LabelPath        string    `json:"labelPath"`
	EpsonJobPath     string    `json:"epsonJobPath"`
	ErrorMessage     string    `json:"errorMessage,omitempty"`
}
type StudyManifest struct {
	StudyInstanceUID string          `json:"studyInstanceUID"`
	Patient          ManifestPatient `json:"patient"`
	StudyDescription string          `json:"studyDescription"`
	Modality         string          `json:"modality"`
	StudyDate        string          `json:"studyDate"`
	Series           []Series        `json:"series"`
}
type ManifestPatient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
