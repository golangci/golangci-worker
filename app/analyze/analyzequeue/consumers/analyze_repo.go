package consumers

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/processors"
)

type AnalyzeRepo struct {
	baseConsumer
}

func NewAnalyzeRepo() *AnalyzeRepo {
	return &AnalyzeRepo{
		baseConsumer: baseConsumer{
			eventName: analytics.EventRepoAnalyzed,
		},
	}
}

func (c AnalyzeRepo) Consume(ctx context.Context, repoName, analysisGUID, branch string) error {
	ctx = c.prepareContext(ctx, map[string]interface{}{
		"repoName": repoName,
		"provider": "github",
	})

	return c.wrapConsuming(ctx, func() error {
		return c.analyzeRepo(ctx, repoName, analysisGUID, branch)
	})
}

func (c AnalyzeRepo) analyzeRepo(ctx context.Context, repoName, analysisGUID, branch string) error {
	p, err := processors.NewGithubGoRepo(ctx, processors.GithubGoRepoConfig{}, analysisGUID, repoName, branch)
	if err != nil {
		return fmt.Errorf("can't make github go repo proessor: %s", err)
	}

	return p.Process(ctx)
}
