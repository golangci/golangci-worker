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

const statusOk = "ok"
const statusFail = "fail"

func analyzePR(ctx context.Context, repoOwner, repoName, githubAccessToken string,
	pullRequestNumber int, APIRequestID string, userID uint, analysisGUID string) error {

	var cancel context.CancelFunc
	// If you change timeout value don't forget to change it
	// in golangci-api stale analyzes checker
	ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	t := &task.PRAnalysis{
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

func analyzePRWrappedV2(ctx context.Context, repoOwner, repoName, githubAccessToken string, pullRequestNumber int, APIRequestID string, userID uint, analysisGUID string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v, %s", r, debug.Stack())
			logrus.Error(err)
		}
	}()
	return analyzePRLogged(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID, analysisGUID)
}

func analyzePRLogged(ctx context.Context, repoOwner, repoName, githubAccessToken string,
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
	err := analyzePR(ctx, repoOwner, repoName, githubAccessToken, pullRequestNumber, APIRequestID, userID, analysisGUID)

	duration := time.Since(startedAt)
	analytics.Log(ctx).Infof("Finished analysis of %s/%s#%d for %s", repoOwner, repoName, pullRequestNumber, duration)

	props := map[string]interface{}{
		"durationSeconds": int(duration / time.Second),
	}
	if err == nil {
		props["status"] = statusOk
	} else {
		props["status"] = statusFail
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

func analyzeRepoWrapped(ctx context.Context, repoName, analysisGUID string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v, %s", r, debug.Stack())
			logrus.Error(err)
		}
	}()

	trackingProps := map[string]interface{}{
		"repoName": repoName,
		"provider": "github",
	}
	ctx = makeContext(ctx, trackingProps)

	analytics.Log(ctx).Infof("Starting repo analysis of %s...", repoName)

	startedAt := time.Now()
	err = analyzeRepo(ctx, repoName, analysisGUID)

	duration := time.Since(startedAt)
	analytics.Log(ctx).Infof("Finished repo analysis of %s for %s", repoName, duration)

	props := map[string]interface{}{
		"durationSeconds": int(duration / time.Second),
	}
	if err == nil {
		props["status"] = statusOk
	} else {
		props["status"] = statusFail
		props["error"] = err.Error()
	}
	analytics.SaveEventProps(ctx, analytics.EventRepoAnalyzed, props)

	if err != nil {
		analytics.Log(ctx).Errorf("repo processing failed: %s", err)
	}

	return err
}

func analyzeRepo(ctx context.Context, repoName, analysisGUID string) error {
	return nil
}

func RegisterTasks() {
	server := queue.GetServer()
	err := server.RegisterTasks(map[string]interface{}{
		"analyzeV2":   analyzePRWrappedV2,
		"analyzeRepo": analyzeRepoWrapped,
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
