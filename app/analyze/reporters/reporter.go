package reporters

import (
	"context"

	"github.com/golangci/golangci-worker/app/analyze/linters/result"
)

type Reporter interface {
	Report(ctx context.Context, ref string, issues []result.Issue) error
}
