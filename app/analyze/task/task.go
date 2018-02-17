package task

import "github.com/golangci/golangci-worker/app/utils/github"

type Task struct {
	github.Context
	APIRequestID string
}
