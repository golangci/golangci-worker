package analyze

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golangci/golangci-worker/app/analyze/analyzerqueue"
	"github.com/golangci/golangci-worker/app/analyze/processors"
	"github.com/golangci/golangci-worker/app/analyze/task"
	"github.com/golangci/golangci-worker/app/test"
	"github.com/golangci/golangci-worker/app/utils/github"
	"github.com/golangci/golangci-worker/app/utils/queue"
	"github.com/stretchr/testify/assert"
)

type processorMocker struct {
	prevProcessorFactory processors.Factory
}

func (pm processorMocker) restore() {
	processorFactory = pm.prevProcessorFactory
}

func mockProcessor(newProcessorFactory processors.Factory) *processorMocker {
	ret := &processorMocker{
		prevProcessorFactory: newProcessorFactory,
	}
	processorFactory = newProcessorFactory
	return ret
}

type testProcessor struct {
	notifyCh chan bool
}

func (tp testProcessor) Process(ctx context.Context) error {
	tp.notifyCh <- true
	return nil
}

type testProcessorFatory struct {
	t        *testing.T
	expTask  *task.Task
	notifyCh chan bool
}

func (tpf testProcessorFatory) BuildProcessor(ctx context.Context, t *task.Task) (processors.Processor, error) {
	assert.Equal(tpf.t, tpf.expTask, t)
	return testProcessor{
		notifyCh: tpf.notifyCh,
	}, nil
}

func TestSendReceiveProcessing(t *testing.T) {
	task := &task.Task{
		Context:      github.FakeContext,
		APIRequestID: "req_id",
	}

	notifyCh := make(chan bool)
	defer mockProcessor(testProcessorFatory{
		t:        t,
		expTask:  task,
		notifyCh: notifyCh,
	}).restore()

	test.Init()
	queue.Init()
	RegisterTasks()
	go RunWorker()

	assert.NoError(t, analyzerqueue.Send(task))

	select {
	case <-notifyCh:
		return
	case <-time.After(time.Second * 1):
		t.Fatalf("Timeouted waiting of processing")
	}
}

func TestAnalyzeSelfRepo(t *testing.T) {
	test.MarkAsSlow(t)
	test.Init()

	prNumber := 1
	if pr := os.Getenv("PR"); pr != "" {
		var err error
		prNumber, err = strconv.Atoi(pr)
		assert.NoError(t, err)
	}
	const userID = 1

	err := analyzeLogged(context.Background(), "golangci", "golangci-worker",
		os.Getenv("TEST_GITHUB_TOKEN"), prNumber, "", userID)
	assert.NoError(t, err)
}
