package processors

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"

	"github.com/golangci/golangci-shared/pkg/logutil"
	"github.com/golangci/golangci-worker/app/lib/executors"
	"github.com/golangci/golangci-worker/app/lib/github"
	"github.com/pkg/errors"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func getEnabledRepoNamesForContainerExperiment() map[string]bool {
	repos := os.Getenv("CONTAINER_EXECUTOR_EXPERIMENT_REPOS")
	if repos == "" {
		return map[string]bool{}
	}

	repoList := strings.Split(repos, ",")
	ret := map[string]bool{}
	for _, r := range repoList {
		ret[r] = true
	}

	return ret
}

func isContainerExecutorExperimentEnabled(repo *github.Repo) bool {
	enabledRepos := getEnabledRepoNamesForContainerExperiment()
	if enabledRepos[repo.FullName()] {
		return true
	}

	percentStr := os.Getenv("CONTAINER_EXECUTOR_EXPERIMENT_PERCENT")
	if percentStr == "" {
		return false
	}

	percent, err := strconv.Atoi(percentStr)
	if err != nil {
		return false
	}

	if percent < 0 || percent > 100 {
		return false
	}

	hash := hash(fmt.Sprintf("%s/%s", repo.Owner, repo.Name))
	return uint32(percent) > (hash % 100)
}

func makeExecutor(ctx context.Context, repo *github.Repo) (executors.Executor, error) {
	var exec executors.Executor
	log := logutil.NewStderrLog("executor")
	log.SetLevel(logutil.LogLevelInfo)

	var useContainerExecutor bool
	useCE := os.Getenv("USE_CONTAINER_EXECUTOR")
	if useCE != "" { //nolint:gocritic
		useContainerExecutor = useCE == "1"
		log.Infof("Container executor is enabled by env var: %t", useContainerExecutor)
	} else if isContainerExecutorExperimentEnabled(repo) {
		useContainerExecutor = true
		log.Infof("Container executor is enabled by experiment")
	} else {
		log.Infof("Container executor is disabled, use remote shell")
	}

	if useContainerExecutor {
		ce, err := executors.NewContainer(log)
		if err != nil {
			return nil, errors.Wrap(err, "can't build container executor")
		}

		if err = ce.Setup(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to setup container executor")
		}
		exec = ce.WithWorkDir("/goapp")
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
