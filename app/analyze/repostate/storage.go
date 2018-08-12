package repostate

import (
	"context"
)

//go:generate mockgen -package repostate -source storage.go -destination storage_mock.go

type State struct {
	Status     string
	ResultJSON interface{}
}

type Storage interface {
	UpdateState(ctx context.Context, owner, name, analysisID string, state *State) error
	GetState(ctx context.Context, owner, name, analysisID string) (*State, error)
}
