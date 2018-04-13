package analyzerqueue

import (
	"fmt"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/golangci/golangci-worker/app/analyze/task"
	"github.com/golangci/golangci-worker/app/utils/queue"
)

func Send(t *task.Task) error {
	signature := &tasks.Signature{
		Name: "analyze",
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: t.Repo.Owner,
			},
			{
				Type:  "string",
				Value: t.Repo.Name,
			},
			{
				Type:  "string",
				Value: t.GithubAccessToken,
			},
			{
				Type:  "int",
				Value: t.PullRequestNumber,
			},
			{
				Type:  "string",
				Value: t.APIRequestID,
			},
			{
				Type:  "uint",
				Value: t.UserID,
			},
		},
		RetryCount:   3,
		RetryTimeout: 600, // 600 sec
	}

	_, err := queue.GetServer().SendTask(signature)
	if err != nil {
		return fmt.Errorf("failed to send the task %v to analyze queue: %s", t, err)
	}

	return nil
}
