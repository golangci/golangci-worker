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

func makeContext(ctx context.Context, trackingProps map[string]interface{}) context.Context {
	ctx = analytics.ContextWithEventPropsCollector(ctx, analytics.EventPRChecked)
	ctx = analytics.ContextWithTrackingProps(ctx, trackingProps)
	return ctx
}

func analyzeWrapped(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string, userID uint) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			logrus.Error(err)
		}
	}()
	return analyzeLogged(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID)
}

func analyzeLogged(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string, userID uint) error {
	trackingProps := map[string]interface{}{
		"repoName":     fmt.Sprintf("%s/%s", repoOwner, repoName),
		"provider":     "github",
		"prNumber":     pullRequestNumber,
		"userIDString": strconv.Itoa(int(userID)),
	}
	ctx = makeContext(ctx, trackingProps)

	startedAt := time.Now()
	err := analyze(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID)

	props := map[string]interface{}{
		"durationSeconds": int(time.Since(startedAt) / time.Second),
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
		analytics.Log(ctx).Errorf("processing failed: %s", err)
	}

	return err
}

func RegisterTasks() {
	server := queue.GetServer()
	server.RegisterTasks(map[string]interface{}{
		"analyze": analyzeWrapped,
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
