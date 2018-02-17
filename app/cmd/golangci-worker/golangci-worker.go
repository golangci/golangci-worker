package main

import (
	"github.com/golangci/golangci-worker/app/analyze"
	"github.com/golangci/golangci-worker/app/utils/queue"
	"github.com/sirupsen/logrus"
)

func main() {
	queue.Init()
	analyze.RegisterTasks()
	if err := analyze.RunWorker(); err != nil {
		logrus.Fatalf("Can't run analyze worker: %s", err)
	}
}
