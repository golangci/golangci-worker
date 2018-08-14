package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/linters"
	"github.com/golangci/golangci-worker/app/analyze/linters/golinters"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/analyze/repostate"
	"github.com/golangci/golangci-worker/app/lib/errorutils"
	"github.com/golangci/golangci-worker/app/lib/executors"
	"github.com/golangci/golangci-worker/app/lib/fetchers"
	"github.com/golangci/golangci-worker/app/lib/github"
	"github.com/golangci/golangci-worker/app/lib/goutils/workspaces"
	"github.com/golangci/golangci-worker/app/lib/httputils"
)

type GithubGoRepoConfig struct {
	repoFetcher fetchers.Fetcher
	linters     []linters.Linter
	runner      linters.Runner
	exec        executors.Executor
	state       repostate.Storage
}

type GithubGoRepo struct {
	analysisGUID string
	branch       string
	gw           *workspaces.Go
	repo         *github.Repo

	GithubGoRepoConfig
	resultCollector
}

func NewGithubGoRepo(ctx context.Context, cfg GithubGoRepoConfig, analysisGUID, repoName, branch string) (*GithubGoRepo, error) {
	parts := strings.Split(repoName, "/")
	repo := &github.Repo{
		Owner: parts[0],
		Name:  parts[1],
	}
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo name %s", repoName)
	}

	if cfg.exec == nil {
		var err error
		cfg.exec, err = makeExecutor(ctx)
		if err != nil {
			return nil, fmt.Errorf("can't make executor: %s", err)
		}
	}

	if cfg.repoFetcher == nil {
		cfg.repoFetcher = fetchers.NewGit()
	}

	if cfg.linters == nil {
		cfg.linters = []linters.Linter{
			golinters.GolangciLint{},
		}
	}

	if cfg.runner == nil {
		cfg.runner = linters.SimpleRunner{}
	}

	if cfg.state == nil {
		cfg.state = repostate.NewAPIStorage(httputils.GrequestsClient{})
	}

	return &GithubGoRepo{
		GithubGoRepoConfig: cfg,
		analysisGUID:       analysisGUID,
		branch:             branch,
		repo:               repo,
	}, nil
}

func (g *GithubGoRepo) prepareRepo(ctx context.Context) error {
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", g.repo.Owner, g.repo.Name)

	var err error
	g.trackTiming("Clone", func() {
		err = g.repoFetcher.Fetch(ctx, cloneURL, g.branch, g.exec)
	})
	if err != nil {
		return &errorutils.InternalError{
			PublicDesc:  "can't clone git repo",
			PrivateDesc: fmt.Sprintf("can't clone git repo: %s", err),
		}
	}

	g.trackTiming("Deps", func() {
		err = g.gw.FetchDeps(ctx)
	})
	if err != nil {
		g.publicWarn("prepare", "Can't fetch deps")
		analytics.Log(ctx).Warnf("Can't fetch deps: %s", err)
	}

	return nil
}

func (g GithubGoRepo) updateAnalysisState(ctx context.Context, res *result.Result, status, publicError string) {
	resJSON := &resultJSON{
		Version: 1,
		WorkerRes: workerRes{
			Timings:  g.timings,
			Warnings: g.warnings,
			Error:    publicError,
		},
	}

	if res != nil {
		resJSON.GolangciLintRes = res.ResultJSON
	}
	s := &repostate.State{
		Status:     status,
		ResultJSON: resJSON,
	}

	jsonBytes, err := json.Marshal(*resJSON)
	if err == nil {
		analytics.Log(ctx).Infof("Save repo analysis status: status=%s, result_json=%s", status, string(jsonBytes))
	}

	if err := g.state.UpdateState(ctx, g.repo.Owner, g.repo.Name, g.analysisGUID, s); err != nil {
		analytics.Log(ctx).Warnf("Can't set analysis %s status to '%v': %s", g.analysisGUID, s, err)
	}
}

func (g *GithubGoRepo) processWithGuaranteedGithubStatus(ctx context.Context) error {
	res, err := g.work(ctx)
	analytics.Log(ctx).Infof("timings: %s", g.timings)

	ctx = context.Background() // no timeout for state and status saving: it must be durable

	var status string
	var publicError string
	if err != nil {
		if ierr, ok := err.(*errorutils.InternalError); ok {
			publicError = ierr.PublicDesc
		} else {
			publicError = internalError
		}
		status = string(github.StatusError)
	} else {
		status = statusProcessed
	}

	g.updateAnalysisState(ctx, res, status, publicError)
	return err
}

func (g *GithubGoRepo) work(ctx context.Context) (res *result.Result, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			err = &errorutils.InternalError{
				PublicDesc:  "golangci-worker panic-ed",
				PrivateDesc: fmt.Sprintf("panic occured: %s, %s", rerr, debug.Stack()),
			}
		}
	}()

	if err = g.prepareRepo(ctx); err != nil {
		return nil, err // don't wrap error, need to save it's type
	}

	g.trackTiming("Analysis", func() {
		res, err = g.runner.Run(ctx, g.linters, g.exec)
	})
	if err != nil {
		return nil, err // don't wrap error, need to save it's type
	}

	return res, nil
}

func (g GithubGoRepo) Process(ctx context.Context) error {
	defer g.exec.Clean()

	g.gw = workspaces.NewGo(g.exec)
	if err := g.gw.Setup(ctx, "github.com", g.repo.Owner, g.repo.Name); err != nil {
		return fmt.Errorf("can't setup go workspace: %s", err)
	}
	g.exec = g.gw.Executor()

	curState, err := g.state.GetState(ctx, g.repo.Owner, g.repo.Name, g.analysisGUID)
	if err != nil {
		return fmt.Errorf("can't get current state: %s", err)
	}

	if curState.Status == statusSentToQueue {
		g.addTimingFrom("In Queue", fromDBTime(curState.CreatedAt))
		curState.Status = statusProcessing
		if err = g.state.UpdateState(ctx, g.repo.Owner, g.repo.Name, g.analysisGUID, curState); err != nil {
			analytics.Log(ctx).Warnf("Can't update repo analysis %s state with setting status to 'processing': %s", g.analysisGUID, err)
		}
	}

	return g.processWithGuaranteedGithubStatus(ctx)
}