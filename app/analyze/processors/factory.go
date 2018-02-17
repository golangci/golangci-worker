package processors

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/golangci/golangci-worker/app/analyze/environments"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/fetchers"
	"github.com/golangci/golangci-worker/app/analyze/linters/golinters"
	"github.com/golangci/golangci-worker/app/analyze/reporters"
	"github.com/golangci/golangci-worker/app/analyze/task"
	"github.com/golangci/golangci-worker/app/utils/github"
)

type Factory interface {
	BuildProcessor(ctx context.Context, t *task.Task) (Processor, error)
}

type githubFactory struct{}

func NewGithubFactory() Factory {
	return githubFactory{}
}

func (gf githubFactory) BuildProcessor(ctx context.Context, t *task.Task) (Processor, error) {
	gc := github.NewMyClient()
	c := &t.Context

	f := fetchers.Git{}
	a := golinters.GetSupportedLinters()
	r := reporters.NewGithubReviewer(c, gc)

	exec, err := gf.makeExecutor(c)
	if err != nil {
		return nil, fmt.Errorf("can't make executor: %s", err)
	}

	return newGithubGo(f, a, r, exec, gc, c), nil
}

func (gf githubFactory) makeExecutor(c *github.Context) (executors.Executor, error) {
	repo := c.Repo
	exec, err := executors.NewTempDirShell("gopath")
	if err != nil {
		return nil, fmt.Errorf("can't create temp dir shell: %s", err)
	}

	gopath := exec.WorkDir()
	wd := path.Join(gopath, "src", "github.com", repo.Owner, repo.Name)
	if err := os.MkdirAll(wd, os.ModePerm); err != nil {
		return nil, fmt.Errorf("can't create project dir %q: %s", wd, err)
	}

	goEnv := environments.NewGolang(gopath)
	goEnv.Setup(exec)

	return exec, nil
}
