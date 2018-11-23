package analyzequeue

import (
	"fmt"

	"github.com/golangci/golangci-shared/pkg/apperrors"
	"github.com/golangci/golangci-shared/pkg/config"
	"github.com/golangci/golangci-shared/pkg/logutil"
	"github.com/golangci/golangci-worker/app/analyze/analyzequeue/consumers"
	"github.com/golangci/golangci-worker/app/analyze/processors"
	"github.com/golangci/golangci-worker/app/lib/experiments"
	"github.com/golangci/golangci-worker/app/lib/queue"
)

func RegisterTasks() {
	log := logutil.NewStderrLog("repo analysis")
	log.SetLevel(logutil.LogLevelInfo)
	cfg := config.NewEnvConfig(log)
	et := apperrors.GetTracker(cfg, log, "worker")

	trackedLog := apperrors.WrapLogWithTracker(log, nil, et)
	ec := experiments.NewChecker(cfg, trackedLog)

	rpf := processors.NewRepoProcessorFactory(&processors.StaticRepoConfig{}, trackedLog)
	repoAnalyzer := consumers.NewAnalyzeRepo(ec, rpf)

	server := queue.GetServer()
	err := server.RegisterTasks(map[string]interface{}{
		"analyzeV2":   consumers.NewAnalyzePR().Consume,
		"analyzeRepo": repoAnalyzer.Consume,
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
