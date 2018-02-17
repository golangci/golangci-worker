package fetchers

import (
	"context"

	"github.com/golangci/golangci-worker/app/analyze/executors"
)

type Fetcher interface {
	Fetch(ctx context.Context, url, ref, destDir string, exec executors.Executor) error
}
