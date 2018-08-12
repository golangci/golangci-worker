package workspaces

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/golangci/golangci-worker/app/lib/executors"
	"github.com/golangci/golangci-worker/app/lib/goutils/environments"
)

type Go struct {
	gopath string
	exec   executors.Executor
}

func NewGo(exec executors.Executor) *Go {
	return &Go{
		exec: exec,
	}
}

func (w *Go) Setup(ctx context.Context, projectPathParts ...string) error {
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
