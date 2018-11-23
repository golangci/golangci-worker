package workspaces

import (
	"context"

	"github.com/golangci/golangci-api/pkg/goenv/result"
	"github.com/golangci/golangci-worker/app/lib/executors"
	"github.com/golangci/golangci-worker/app/lib/fetchers"
)

type Installer interface {
	Setup(ctx context.Context, repo *fetchers.Repo, projectPathParts ...string) (executors.Executor, *result.Log, error)
}
