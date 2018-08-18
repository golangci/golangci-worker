package fetchers

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-worker/app/lib/executors"
)

type Git struct{}

func NewGit() *Git {
	return &Git{}
}

func (gf Git) Fetch(ctx context.Context, repo *Repo, exec executors.Executor) error {
	args := []string{"clone", "-q", "--depth", "1", "--branch", repo.Ref, repo.CloneURL, "."}
	if out, err := exec.Run(ctx, "git", args...); err != nil {
		return fmt.Errorf("can't run git cmd %v: %s, out is %s", args, err, out)
	}

	return nil
}
