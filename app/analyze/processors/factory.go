package processors

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-worker/app/analyze/task"
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
		return nil, fmt.Errorf("can't make github go processor: %s", err)
	}

	return p, nil
}
