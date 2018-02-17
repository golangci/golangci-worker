package golinters

import (
	"context"
	"testing"

	"github.com/golangci/golangci-worker/app/analyze/environments"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/test"
	"github.com/golangci/golangci-worker/app/utils/fsutils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRunLintersInDocker(t *testing.T) {
	test.MarkAsSlow(t)

	exec := executors.NewDocker(fsutils.GetProjectRoot(), "/app/go/src/github.com/golangci/golangci-worker")
	goEnv := environments.NewGolang("/app/go")
	goEnv.Setup(exec)

	ctx := context.Background()
	linters := GetSupportedLinters()
	for _, lint := range linters {
		res, err := lint.Run(ctx, exec)
		assert.NoError(t, err)
		logrus.Infof("linter result: %+v", res.Issues)
	}
}
