package analyze

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-worker/app/analyze/processors"
	"github.com/golangci/golangci-worker/app/analyze/task"
	"github.com/golangci/golangci-worker/app/utils/github"
	"github.com/golangci/golangci-worker/app/utils/queue"
	"github.com/sirupsen/logrus"
)

var processorFactory = processors.NewGithubFactory()

func SetProcessorFactory(f processors.Factory) {
	processorFactory = f
}

func analyze(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string) error {
	t := &task.Task{
		Context: github.Context{
			Repo: github.Repo{
				Owner: repoOwner,
				Name:  repoName,
			},
			GithubAccessToken: githubAccessToken,
			PullRequestNumber: pullRequestNumber,
		},
		APIRequestID: APIRequestID,
	}

	p, err := processorFactory.BuildProcessor(ctx, t)
	if err != nil {
		return fmt.Errorf("can't build processor for task %+v: %s", t, err)
	}

	if err = p.Process(ctx); err != nil {
		return fmt.Errorf("can't process task %+v: %s", t, err)
	}

	return nil
}

func analyzeLogged(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string) error {
	err := analyze(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID)
	if err != nil {
		logrus.Errorf("processing failed: %s", err)
	}

	return err
}

func RegisterTasks() {
	server := queue.GetServer()
	server.RegisterTasks(map[string]interface{}{
		"analyze": analyzeLogged,
	})
}

func RunWorker() error {
	server := queue.GetServer()
	worker := server.NewWorker("worker_name", 1)
	err := worker.Launch()
	if err != nil {
		return fmt.Errorf("can't launch worker: %s", err)
	}

	return nil
}
