package linters

import (
	"context"

	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
)

type Linter interface {
	Run(ctx context.Context, exec executors.Executor) (*result.Result, error)
	Name() string
}
