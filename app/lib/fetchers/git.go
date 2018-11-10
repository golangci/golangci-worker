package fetchers

import (
	"context"
	"strconv"
	"strings"

	"github.com/golangci/golangci-worker/app/lib/executors"
	"github.com/pkg/errors"
)

var ErrNoBranchOrRepo = errors.New("repo or branch not found")

type Git struct{}

func NewGit() *Git {
	return &Git{}
}

func (gf Git) Fetch(ctx context.Context, repo *Repo, exec executors.Executor) error {
	args := []string{"clone", "-q", "--depth", "1", "--branch",
		strconv.Quote(repo.Ref), strconv.Quote(repo.CloneURL), "."}
	if out, err := exec.Run(ctx, "git", args...); err != nil {
		noBranchOrRepo := strings.Contains(err.Error(), "could not read Username for") ||
			strings.Contains(err.Error(), "Could not find remote branch")
		if noBranchOrRepo {
			return errors.Wrap(ErrNoBranchOrRepo, err.Error())
		}

		return errors.Wrapf(err, "can't run git cmd %v: %s", args, out)
	}

	return nil
}
