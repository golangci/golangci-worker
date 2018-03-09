package analyze

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/processors"
	"github.com/golangci/golangci-worker/app/analyze/task"
	"github.com/golangci/golangci-worker/app/utils/github"
	"github.com/golangci/golangci-worker/app/utils/queue"
	"github.com/sirupsen/logrus"
	"github.com/stvp/rollbar"
)

var processorFactory = processors.NewGithubFactory()

func SetProcessorFactory(f processors.Factory) {
	processorFactory = f
}

func analyze(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string, userID uint) error {
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
		UserID:       userID,
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

func analyzeLogged(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string, userID uint) error {
	ctx = analytics.ContextWithEventPropsCollector(ctx, analytics.EventPRChecked)
	initalProps := map[string]interface{}{
		"repoName":     fmt.Sprintf("%s/%s", repoOwner, repoName),
		"provider":     "github",
		"prNumber":     pullRequestNumber,
		"userIDString": strconv.Itoa(int(userID)),
	}
	analytics.SaveEventProps(ctx, analytics.EventPRChecked, initalProps)

	startedAt := time.Now()
	err := analyze(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID)
	if err != nil {
		logrus.Errorf("processing failed: %s", err)
	}

	props := map[string]interface{}{
		"durationSeconds": time.Since(startedAt) / time.Second,
	}
	if err == nil {
		props["status"] = "ok"
	} else {
		props["status"] = "fail"
		props["error"] = err.Error()
	}
	analytics.SaveEventProps(ctx, analytics.EventPRChecked, props)

	tracker := analytics.GetTracker(ctx)
	tracker.Track(ctx, analytics.EventPRChecked)

	if err != nil {
		trackError(ctx, err, initalProps)
	}

	return err
}

func trackError(ctx context.Context, err error, props map[string]interface{}) {
	f := &rollbar.Field{
		Name: "props",
		Data: props,
	}
	rollbar.Error("ERROR", err, f)
	logrus.Warnf("Tracked error to rollbar: %s", err)
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
