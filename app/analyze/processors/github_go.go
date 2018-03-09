package processors

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/environments"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/fetchers"
	"github.com/golangci/golangci-worker/app/analyze/linters"
	"github.com/golangci/golangci-worker/app/analyze/linters/golinters"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	lp "github.com/golangci/golangci-worker/app/analyze/linters/result/processors"
	"github.com/golangci/golangci-worker/app/analyze/reporters"
	"github.com/golangci/golangci-worker/app/utils/fsutils"
	"github.com/golangci/golangci-worker/app/utils/github"
	gh "github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

type githubGoConfig struct {
	repoFetcher fetchers.Fetcher
	linters     []linters.Linter
	runner      linters.Runner
	reporter    reporters.Reporter
	exec        executors.Executor
	client      github.Client
}

type githubGo struct {
	pr *gh.PullRequest

	context *github.Context
	githubGoConfig
}

type analyticsProcessor struct {
	key string
	ctx context.Context
}

func (ap analyticsProcessor) Process(results []result.Result) ([]result.Result, error) {
	issuesCount := 0
	for _, r := range results {
		issuesCount += len(r.Issues)
	}
	analytics.SaveEventProp(ap.ctx, analytics.EventPRChecked, ap.key, issuesCount)
	return results, nil
}

func (ap analyticsProcessor) Name() string {
	return fmt.Sprintf("analytics processor <%s>", ap.key)
}

func getLinterProcessors(ctx context.Context, patch string) []lp.Processor {
	return []lp.Processor{
		lp.NewExcludeProcessor(`(should have comment|comment on exported method)`),
		lp.UniqByLineProcessor{},
		analyticsProcessor{
			key: "totalIssues",
			ctx: ctx,
		},
		lp.NewDiffProcessor(patch),
		lp.MaxLinterIssuesPerFile{},
		analyticsProcessor{
			key: "reportedIssues",
			ctx: ctx,
		},
	}
}

func newGithubGo(ctx context.Context, c *github.Context, cfg githubGoConfig) (*githubGo, error) {
	if cfg.exec == nil {
		exec, err := makeExecutor(c)
		if err != nil {
			return nil, err
		}
		cfg.exec = exec
	}

	if cfg.repoFetcher == nil {
		cfg.repoFetcher = fetchers.Git{}
	}

	if cfg.linters == nil {
		cfg.linters = golinters.GetSupportedLinters()
	}

	if cfg.client == nil {
		cfg.client = github.NewMyClient()
	}

	if cfg.reporter == nil {
		cfg.reporter = reporters.NewGithubReviewer(c, cfg.client)
	}

	if cfg.runner == nil {
		patch, err := cfg.client.GetPullRequestPatch(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("can't get patch: %s", err)
		}

		cfg.runner = linters.SimpleRunner{
			Processors: getLinterProcessors(ctx, patch),
		}
	}

	return &githubGo{
		context:        c,
		githubGoConfig: cfg,
	}, nil
}

func makeExecutor(c *github.Context) (executors.Executor, error) {
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
		analytics.Log(ctx).Warnf("Can't ensure deps: %s, %s", err, out)
	}

	return nil
}

func (g githubGo) runLinters(ctx context.Context) ([]result.Issue, error) {
	return g.runner.Run(ctx, g.linters, g.exec)
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
		analytics.Log(ctx).Warnf("Can't set commit status: %s", err)
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
