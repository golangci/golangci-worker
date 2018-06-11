package state

import (
	"context"
)

type State struct {
	Status              string
	ReportedIssuesCount int
	ResultJSON          interface{}
}

type Storage interface {
	UpdateState(ctx context.Context, analysisID string, state *State) error
	GetState(ctx context.Context, analysisID string) (*State, error)
}
