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
	UpdateState(ctx context.Context, owner, name, analysisID string, state *State) error
	GetState(ctx context.Context, owner, name, analysisID string) (*State, error)
}
