package processors

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/task"
	"github.com/golangci/golangci-worker/app/utils/github"
)

type Factory interface {
	BuildProcessor(ctx context.Context, t *task.Task) (Processor, error)
}

type githubFactory struct{}

func NewGithubFactory() Factory {
	return githubFactory{}
}

func (gf githubFactory) BuildProcessor(ctx context.Context, t *task.Task) (Processor, error) {
	p, err := newGithubGo(ctx, &t.Context, githubGoConfig{})
	if err != nil {
		if err == github.ErrPRNotFound {
			analytics.Log(ctx).Warnf("No pull request to analyze, skip current task: use nop processor")
			return NopProcessor{}, nil
		}
		return nil, fmt.Errorf("can't make github go processor: %s", err)
	}

	return p, nil
}
