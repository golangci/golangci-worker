package analyze

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
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

func analyze(ctx context.Context, repoOwner, repoName, githubAccessToken string,
	pullRequestNumber int, APIRequestID string, userID uint, analysisGUID string) error {

	var cancel context.CancelFunc
	// If you change timeout value don't forget to change it
	// in golangci-api stale analyzes checker
	ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

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
		AnalysisGUID: analysisGUID,
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
	return analyzeWrappedV2(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID, "")
}

func analyzeWrappedV2(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string, userID uint, analysisGUID string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v, %s", r, debug.Stack())
			logrus.Error(err)
		}
	}()
	return analyzeLogged(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID, analysisGUID)
}

func analyzeLogged(ctx context.Context, repoOwner, repoName, githubAccessToken string,
	pullRequestNumber int, APIRequestID string, userID uint, analysisGUID string) error {

	trackingProps := map[string]interface{}{
		"repoName":     fmt.Sprintf("%s/%s", repoOwner, repoName),
		"provider":     "github",
		"prNumber":     pullRequestNumber,
		"userIDString": strconv.Itoa(int(userID)),
		"analysisGUID": analysisGUID,
	}
	ctx = makeContext(ctx, trackingProps)

	analytics.Log(ctx).Infof("Starting analysis of %s/%s#%d...", repoOwner, repoName, pullRequestNumber)

	startedAt := time.Now()
	err := analyze(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID, analysisGUID)

	duration := time.Since(startedAt)
	analytics.Log(ctx).Infof("Finished analysis of %s/%s#%d for %s", repoOwner, repoName, pullRequestNumber, duration)

	props := map[string]interface{}{
		"durationSeconds": int(duration / time.Second),
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
	err := server.RegisterTasks(map[string]interface{}{
		"analyze":   analyzeWrapped,
		"analyzeV2": analyzeWrappedV2,
	})
	if err != nil {
		log.Fatalf("Can't register queue tasks: %s", err)
	}
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
