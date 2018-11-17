package consumers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

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
		"repoName":     repoName,
		"provider":     "github",
		"analysisGUID": analysisGUID,
		"branch":       branch,
	})

	if os.Getenv("DISABLE_REPO_ANALYSIS") == "1" {
		analytics.Log(ctx).Warnf("Repo analysis is disabled, return error to try it later")
		return errors.New("repo analysis is disabled")
	}

	_ = c.wrapConsuming(ctx, func() error {
		var cancel context.CancelFunc
		// If you change timeout value don't forget to change it
		// in golangci-api stale analyzes checker
		ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		return c.analyzeRepo(ctx, repoName, analysisGUID, branch)
	})

	// Don't return error to machinery: we will retry this task ourself from golangci-api
	return nil
}

func (c AnalyzeRepo) analyzeRepo(ctx context.Context, repoName, analysisGUID, branch string) error {
	p, err := processors.NewGithubGoRepo(ctx, processors.GithubGoRepoConfig{}, analysisGUID, repoName, branch)
	if err != nil {
		return fmt.Errorf("can't make github go repo processor: %s", err)
	}

	if err := p.Process(ctx); err != nil {
		return fmt.Errorf("can't process repo analysis for %s and branch %s: %s", repoName, branch, err)
	}

	return nil
}
