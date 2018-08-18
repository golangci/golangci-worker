package repoinfo

import (
	"context"
	"encoding/json"

	info "github.com/golangci/getrepoinfo/pkg/repoinfo"
	"github.com/golangci/golangci-worker/app/lib/executors"
	"github.com/golangci/golangci-worker/app/lib/fetchers"
	"github.com/pkg/errors"
)

//go:generate mockgen -package repoinfo -source fetcher.go -destination fetcher_mock.go

type Info info.Info

type Fetcher interface {
	Fetch(ctx context.Context, repo *fetchers.Repo, exec executors.Executor) (*Info, error)
}

type CloningFetcher struct {
	repoFetcher fetchers.Fetcher
}

func NewCloningFetcher(repoFetcher fetchers.Fetcher) *CloningFetcher {
	return &CloningFetcher{
		repoFetcher: repoFetcher,
	}
}

func (f CloningFetcher) Fetch(ctx context.Context, repo *fetchers.Repo, exec executors.Executor) (*Info, error) {
	// fetch into the current dir
	if err := f.repoFetcher.Fetch(ctx, repo, exec); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch repo ref %q by url %q", repo.Ref, repo.CloneURL)
	}

	out, err := exec.Run(ctx, "getrepoinfo")
	if err != nil {
		return nil, errors.Wrap(err, "failed to run 'getrepoinfo'")
	}

	var ret Info
	if err = json.Unmarshal([]byte(out), &ret); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	return &ret, nil
}
