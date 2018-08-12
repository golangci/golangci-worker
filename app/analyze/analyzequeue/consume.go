package analyzequeue

import (
	"fmt"
	"log"

	"github.com/golangci/golangci-worker/app/analyze/analyzequeue/consumers"
	"github.com/golangci/golangci-worker/app/lib/queue"
)

func RegisterTasks() {
	server := queue.GetServer()
	err := server.RegisterTasks(map[string]interface{}{
		"analyzeV2":   consumers.NewAnalyzePR().Consume,
		"analyzeRepo": consumers.NewAnalyzeRepo().Consume,
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
