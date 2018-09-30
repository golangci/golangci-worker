package workspaces

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/repoinfo"
	"github.com/golangci/golangci-worker/app/lib/executors"
	"github.com/golangci/golangci-worker/app/lib/fetchers"
	"github.com/golangci/golangci-worker/app/lib/goutils/environments"
	"github.com/pkg/errors"
)

type Go struct {
	gopath      string
	exec        executors.Executor
	infoFetcher repoinfo.Fetcher
}

func NewGo(exec executors.Executor, infoFetcher repoinfo.Fetcher) *Go {
	return &Go{
		exec:        exec,
		infoFetcher: infoFetcher,
	}
}

func (w *Go) Setup(ctx context.Context, repo *fetchers.Repo, projectPathParts ...string) error {
	repoInfo, err := w.infoFetcher.Fetch(ctx, repo, w.exec)
	if err != nil {
		return errors.Wrap(err, "failed to fetch repo info")
	}

	if repoInfo != nil && repoInfo.CanonicalImportPath != "" {
		newProjectPathParts := strings.Split(repoInfo.CanonicalImportPath, "/")
		analytics.Log(ctx).Infof("change canonical project path: %s -> %s", projectPathParts, newProjectPathParts)
		projectPathParts = newProjectPathParts
	}

	if _, err := w.exec.Run(ctx, "rm", "-rf", "*"); err != nil {
		return errors.Wrap(err, "failed to cleanup after repo info fetcher")
	}

	gopath := w.exec.WorkDir()
	wdParts := []string{gopath, "src"}
	wdParts = append(wdParts, projectPathParts...)
	wd := filepath.Join(wdParts...)
	if out, err := w.exec.Run(ctx, "mkdir", "-p", wd); err != nil {
		return fmt.Errorf("can't create project dir %q: %s, %s", wd, err, out)
	}

	goEnv := environments.NewGolang(gopath)
	goEnv.Setup(w.exec)

	w.exec = w.exec.WithWorkDir(wd) // XXX: clean gopath, but work in subdir of gopath

	w.gopath = gopath
	return nil
}

func (w Go) Executor() executors.Executor {
	return w.exec
}

func (w Go) Gopath() string {
	return w.gopath
}

func (w Go) FetchDeps(ctx context.Context) error {
	depsPath := filepath.Join("/app", "ensure_deps.sh")
	out, err := w.exec.Run(ctx, "bash", depsPath)
	if err != nil {
		return fmt.Errorf("can't call /app/ensure_deps.sh: %s, %s", err, out)
	}

	return nil
}
