package adapters

import (
	"context"

	"github.com/local/dicom-disc-suite/shared/models"
)

type EpsonPublisher interface {
	CreateJob(context.Context, models.DiscJob) (string, error)
	SubmitJob(context.Context, string) error
}
