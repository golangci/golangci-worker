package processors

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/analyzequeue/task"
	"github.com/golangci/golangci-worker/app/utils/github"
)

type Factory interface {
	BuildProcessor(ctx context.Context, t *task.PRAnalysis) (Processor, error)
}

type githubFactory struct{}

func NewGithubFactory() Factory {
	return githubFactory{}
}

func (gf githubFactory) BuildProcessor(ctx context.Context, t *task.PRAnalysis) (Processor, error) {
	p, err := newGithubGo(ctx, &t.Context, githubGoConfig{}, t.AnalysisGUID)
	if err != nil {
		if !github.IsRecoverableError(err) {
			analytics.Log(ctx).Warnf("%s: skip current task: use nop processor", err)
			return NopProcessor{}, nil
		}
		return nil, fmt.Errorf("can't make github go processor: %s", err)
	}

	return p, nil
}
