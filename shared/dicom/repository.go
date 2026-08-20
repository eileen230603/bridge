package dicom

import (
	"context"
	"github.com/local/dicom-disc-suite/shared/models"
	"time"
)

type StudyRepository interface {
	SearchStudies(context.Context, time.Time, time.Time) ([]models.Study, error)
	RetrieveStudy(context.Context, string, string) error
}
