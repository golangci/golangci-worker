package processors

import (
	"context"
	"fmt"
	"path"

	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/fetchers"
	"github.com/golangci/golangci-worker/app/analyze/linters"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	lp "github.com/golangci/golangci-worker/app/analyze/linters/result/processors"
	"github.com/golangci/golangci-worker/app/analyze/reporters"
	"github.com/golangci/golangci-worker/app/utils/fsutils"
	"github.com/golangci/golangci-worker/app/utils/github"
	gh "github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

type githubGo struct {
	pr *gh.PullRequest

	repoFetcher fetchers.Fetcher
	linters     []linters.Linter
	reporter    reporters.Reporter
	exec        executors.Executor
	client      github.Client
	context     *github.Context

	status string
}

func newGithubGo(f fetchers.Fetcher, a []linters.Linter,
	r reporters.Reporter, exec executors.Executor, client github.Client, c *github.Context) *githubGo {

	return &githubGo{
		repoFetcher: f,
		linters:     a,
		reporter:    r,
		exec:        exec,
		client:      client,
		context:     c,
	}
}

func (g githubGo) prepareRepo(ctx context.Context) error {
	cloneURL := g.pr.GetHead().GetRepo().GetSSHURL()
	clonePath := "." // Must be already in needed dir
	ref := g.pr.GetHead().GetRef()
	if err := g.repoFetcher.Fetch(ctx, cloneURL, ref, clonePath, g.exec); err != nil {
		return fmt.Errorf("can't fetch git repo: %s", err)
	}

	projRoot := fsutils.GetProjectRoot()
	depsPath := path.Join(projRoot, "app", "scripts", "ensure_deps.sh")
	if out, err := g.exec.Run(ctx, "bash", depsPath); err != nil {
		logrus.Warnf("Can't ensure deps: %s, %s", err, out)
	}

	return nil
}

func (g githubGo) runLinters(ctx context.Context) ([]result.Issue, error) {
	patch, err := g.client.GetPullRequestPatch(ctx, g.context)
	if err != nil {
		return nil, fmt.Errorf("can't get patch: %s", err)
	}
	r := linters.SimpleRunner{
		Processors: []lp.Processor{
			lp.NewExcludeProcessor(`(should have comment)`),
			lp.UniqByLineProcessor{},
			lp.NewDiffProcessor(patch),
			lp.MaxLinterIssuesPerFile{},
		},
	}

	return r.Run(ctx, g.linters, g.exec)
}

func (g githubGo) processInWorkDir(ctx context.Context) error {
	status := github.StatusSuccess // Hide all out internal errors
	statusDesc := "No issues found!"
	defer func() {
		g.setCommitStatus(ctx, status, statusDesc)
	}()

	if err := g.prepareRepo(ctx); err != nil {
		return err
	}

	issues, err := g.runLinters(ctx)
	if err != nil {
		return err
	}

	logrus.Infof("Linters found next issues: %+v", issues)
	if err = g.reporter.Report(ctx, g.pr.GetHead().GetSHA(), issues); err != nil {
		return fmt.Errorf("can't report: %s", err)
	}

	switch len(issues) {
	case 0:
		return nil // Status is really success
	case 1:
		statusDesc = "1 issue found"
	default:
		statusDesc = fmt.Sprintf("%d issues found", len(issues))
	}

	status = github.StatusFailure
	return nil
}

func (g githubGo) setCommitStatus(ctx context.Context, status github.Status, desc string) {
	err := g.client.SetCommitStatus(ctx, g.context, g.pr.GetHead().GetSHA(), status, desc)
	if err != nil {
		logrus.Warnf("Can't set commit status: %s", err)
	}
}

func (g githubGo) Process(ctx context.Context) error {
	defer g.exec.Clean()

	var err error
	g.pr, err = g.client.GetPullRequest(ctx, g.context)
	if err != nil {
		return fmt.Errorf("can't get pull request: %s", err)
	}

	g.setCommitStatus(ctx, github.StatusPending, "GolangCI is reviewing your Pull Request...")

	r := g.context.Repo
	wd := path.Join(g.exec.WorkDir(), "src", "github.com", r.Owner, r.Name)
	g.exec = g.exec.WithWorkDir(wd) // XXX: clean gopath, but work in subdir of gopath

	return g.processInWorkDir(ctx)
}
