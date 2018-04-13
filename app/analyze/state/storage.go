package state

import (
	"context"
)

type Storage interface {
	UpdateStatus(ctx context.Context, analysisID, status string) error
	GetStatus(ctx context.Context, analysisID string) (string, error)
}
