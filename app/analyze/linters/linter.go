package linters

import (
	"context"

	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/lib/executors"
)

//go:generate mockgen -package linters -source linter.go -destination linter_mock.go

type Linter interface {
	Run(ctx context.Context, exec executors.Executor) (*result.Result, error)
	Name() string
}
