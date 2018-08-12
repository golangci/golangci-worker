package processors

import (
	"context"
	"fmt"
	"os"

	"github.com/golangci/golangci-worker/app/lib/executors"
)

func makeExecutor(ctx context.Context) (executors.Executor, error) {
	var exec executors.Executor
	useDockerExecutor := os.Getenv("USE_DOCKER_EXECUTOR") == "1"
	if useDockerExecutor {
		var err error
		exec, err = executors.NewDocker(ctx)
		if err != nil {
			return nil, fmt.Errorf("can't build docker executor: %s", err)
		}
	} else {
		s := executors.NewRemoteShell(
			os.Getenv("REMOTE_SHELL_USER"),
			os.Getenv("REMOTE_SHELL_HOST"),
			os.Getenv("REMOTE_SHELL_KEY_FILE_PATH"),
		)
		if err := s.SetupTempWorkDir(ctx); err != nil {
			return nil, fmt.Errorf("can't setup temp work dir: %s", err)
		}

		exec = s
	}

	return exec, nil
}
