package state

import (
	"context"
	"time"
)

type State struct {
	CreatedAt           time.Time
	Status              string
	ReportedIssuesCount int
	ResultJSON          interface{}
}

type Storage interface {
	UpdateState(ctx context.Context, owner, name, analysisID string, state *State) error
	GetState(ctx context.Context, owner, name, analysisID string) (*State, error)
}
