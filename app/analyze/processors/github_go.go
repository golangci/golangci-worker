package processors

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/environments"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/fetchers"
	"github.com/golangci/golangci-worker/app/analyze/linters"
	"github.com/golangci/golangci-worker/app/analyze/linters/golinters"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/analyze/reporters"
	"github.com/golangci/golangci-worker/app/analyze/state"
	"github.com/golangci/golangci-worker/app/utils/errorutils"
	"github.com/golangci/golangci-worker/app/utils/github"
	gh "github.com/google/go-github/github"
)

type githubGoConfig struct {
	repoFetcher fetchers.Fetcher
	linters     []linters.Linter
	runner      linters.Runner
	reporter    reporters.Reporter
	exec        executors.Executor
	client      github.Client
	state       state.Storage
}

type JSONDuration time.Duration

type Timing struct {
	Name     string
	Duration JSONDuration `json:"DurationMs"`
}

type Warning struct {
	Tag  string
	Text string
}

type githubGo struct {
	pr           *gh.PullRequest
	analysisGUID string

	context *github.Context
	githubGoConfig

	timings  []Timing
	warnings []Warning
}

//nolint:gocyclo
func newGithubGo(ctx context.Context, c *github.Context, cfg githubGoConfig, analysisGUID string) (*githubGo, error) {
	if cfg.client == nil {
		cfg.client = github.NewMyClient()
	}

	if cfg.exec == nil {
		patch, err := cfg.client.GetPullRequestPatch(ctx, c)
		if err != nil {
			if !github.IsRecoverableError(err) {
				return nil, err // preserve error
			}
			return nil, fmt.Errorf("can't get patch: %s", err)
		}

		exec, err := makeExecutor(ctx, c, patch)
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

	if cfg.reporter == nil {
		cfg.reporter = reporters.NewGithubReviewer(c, cfg.client)
	}

	if cfg.runner == nil {
		cfg.runner = linters.SimpleRunner{}
	}

	if cfg.state == nil {
		cfg.state = state.NewAPIStorage()
	}

	return &githubGo{
		context:        c,
		githubGoConfig: cfg,
		analysisGUID:   analysisGUID,
	}, nil
}

func makeExecutor(ctx context.Context, c *github.Context, patch string) (executors.Executor, error) {
	repo := c.Repo
	var exec executors.Executor
	useDockerExecutor := os.Getenv("USE_DOCKER_EXECUTOR") == "1"
	if useDockerExecutor {
		var err error
		exec, err = executors.NewDocker(ctx)
		if err != nil {
			return nil, fmt.Errorf("can't build docker executor: %s", err)
		}
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

	f, err := ioutil.TempFile("/tmp", "golangci.diff")
	defer os.Remove(f.Name())

	if err != nil {
		return nil, fmt.Errorf("can't create temp file for patch: %s", err)
	}
	if err = ioutil.WriteFile(f.Name(), []byte(patch), os.ModePerm); err != nil {
		return nil, fmt.Errorf("can't write patch to temp file %s: %s", f.Name(), err)
	}

	if err = exec.CopyFile(ctx, "changes.patch", f.Name()); err != nil {
		return nil, fmt.Errorf("can't copy patch file to remote shell: %s", err)
	}

	gopath := exec.WorkDir()
	wd := path.Join(gopath, "src", "github.com", repo.Owner, repo.Name)
	if out, err := exec.Run(ctx, "mkdir", "-p", wd); err != nil {
		return nil, fmt.Errorf("can't create project dir %q: %s, %s", wd, err, out)
	}

	goEnv := environments.NewGolang(gopath)
	goEnv.Setup(exec)

	return exec, nil
}

func (g *githubGo) prepareRepo(ctx context.Context) error {
	var cloneURL string
	if g.pr.Base.Repo.GetPrivate() {
		cloneURL = fmt.Sprintf("https://%s@github.com/%s/%s.git",
			g.context.GithubAccessToken, // it's already the private token
			g.context.Repo.Owner, g.context.Repo.Name)
	} else {
		cloneURL = g.pr.GetHead().GetRepo().GetCloneURL()
	}
	clonePath := "." // Must be already in needed dir
	ref := g.pr.GetHead().GetRef()

	var err error
	g.trackTiming("Clone", func() {
		err = g.repoFetcher.Fetch(ctx, cloneURL, ref, clonePath, g.exec)
	})
	if err != nil {
		return &errorutils.InternalError{
			PublicDesc:  "can't clone git repo",
			PrivateDesc: fmt.Sprintf("can't clone git repo: %s", err),
		}
	}

	depsPath := path.Join("/app", "ensure_deps.sh")
	var out string
	g.trackTiming("Deps", func() {
		out, err = g.exec.Run(ctx, "bash", depsPath)
	})
	if err != nil {
		g.publicWarn("prepare", "Can't fetch deps")
		analytics.Log(ctx).Warnf("Can't fetch deps: %s, %s", err, out)
	}

	return nil
}

type workerRes struct {
	Timings  []Timing  `json:",omitempty"`
	Warnings []Warning `json:",omitempty"`
	Error    string    `json:",omitempty"`
}

type resultJSON struct {
	Version         int
	GolangciLintRes interface{}
	WorkerRes       workerRes
}

func (g githubGo) updateAnalysisState(ctx context.Context, res *result.Result, status github.Status, publicError string) {
	resJSON := &resultJSON{
		Version: 1,
		WorkerRes: workerRes{
			Timings:  g.timings,
			Warnings: g.warnings,
			Error:    publicError,
		},
	}

	issuesCount := 0
	if res != nil {
		resJSON.GolangciLintRes = res.ResultJSON
		issuesCount = len(res.Issues)
	}
	s := &state.State{
		Status:              "processed/" + string(status),
		ReportedIssuesCount: issuesCount,
		ResultJSON:          resJSON,
	}

	if err := g.state.UpdateState(ctx, g.context.Repo.Owner, g.context.Repo.Name, g.analysisGUID, s); err != nil {
		analytics.Log(ctx).Warnf("Can't set analysis %s status to '%v': %s", g.analysisGUID, s, err)
	}
}

func getGithubStatusForIssues(issues []result.Issue) (github.Status, string) {
	switch len(issues) {
	case 0:
		return github.StatusSuccess, "No issues found!"
	case 1:
		return github.StatusFailure, "1 issue found"
	default:
		return github.StatusFailure, fmt.Sprintf("%d issues found", len(issues))
	}
}

type IgnoredError struct {
	Status        github.Status
	StatusDesc    string
	IsRecoverable bool
}

func (e IgnoredError) Error() string {
	return e.StatusDesc
}

func (g *githubGo) processWithGuaranteedGithubStatus(ctx context.Context) error {
	res, err := g.work(ctx)
	analytics.Log(ctx).Infof("timings: %s", g.timings)

	ctx = context.Background() // no timeout for state and status saving: it must be durable

	var status github.Status
	var statusDesc, publicError string
	if err != nil {
		if serr, ok := err.(*IgnoredError); ok {
			status, statusDesc = serr.Status, serr.StatusDesc
			if !serr.IsRecoverable {
				err = nil
			}
			// already must have warning, don't set publicError
		} else if ierr, ok := err.(*errorutils.InternalError); ok {
			status, statusDesc = github.StatusError, ierr.PublicDesc
			publicError = statusDesc
		} else {
			status, statusDesc = github.StatusError, "Internal error"
			publicError = statusDesc
		}
	} else {
		status, statusDesc = getGithubStatusForIssues(res.Issues)
	}

	// update of state must be before commit status update: user can open details link before: race condition
	g.updateAnalysisState(ctx, res, status, publicError)
	g.setCommitStatus(ctx, status, statusDesc)

	return err
}

func (g *githubGo) work(ctx context.Context) (res *result.Result, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			err = &errorutils.InternalError{
				PublicDesc:  "golangci-worker panic-ed",
				PrivateDesc: fmt.Sprintf("panic occured: %s, %s", rerr, debug.Stack()),
			}
		}
	}()

	prState := strings.ToUpper(g.pr.GetState())
	if prState == "MERGED" || prState == "CLOSED" {
		// branch can be deleted: will be an error; no need to analyze
		g.publicWarn("process", fmt.Sprintf("Pull Request is already %s, skip analysis", prState))
		analytics.Log(ctx).Warnf("Pull Request is already %s, skip analysis", prState)
		return nil, &IgnoredError{
			Status:        github.StatusSuccess,
			StatusDesc:    fmt.Sprintf("Pull Request is already %s", strings.ToLower(prState)),
			IsRecoverable: false,
		}
	}

	if err = g.prepareRepo(ctx); err != nil {
		return nil, err // don't wrap error, need to save it's type
	}

	g.trackTiming("Analysis", func() {
		res, err = g.runner.Run(ctx, g.linters, g.exec)
	})
	if err != nil {
		return nil, err // don't wrap error, need to save it's type
	}

	issues := res.Issues
	analytics.SaveEventProp(ctx, analytics.EventPRChecked, "reportedIssues", len(issues))

	if len(issues) == 0 {
		analytics.Log(ctx).Infof("Linters found no issues")
	} else {
		analytics.Log(ctx).Infof("Linters found %d issues: %+v", len(issues), issues)
	}

	if err = g.reporter.Report(ctx, g.pr.GetHead().GetSHA(), issues); err != nil {
		return nil, &errorutils.InternalError{
			PublicDesc:  "can't send pull request comments to github",
			PrivateDesc: fmt.Sprintf("can't send pull request comments to github: %s", err),
		}
	}

	return res, nil
}

func (g githubGo) setCommitStatus(ctx context.Context, status github.Status, desc string) {
	var url string
	if status == github.StatusFailure || status == github.StatusSuccess || status == github.StatusError {
		c := g.context
		url = fmt.Sprintf("%s/r/%s/%s/pulls/%d",
			os.Getenv("WEB_ROOT"), c.Repo.Owner, c.Repo.Name, g.pr.GetNumber())
	}
	err := g.client.SetCommitStatus(ctx, g.context, g.pr.GetHead().GetSHA(), status, desc, url)
	if err != nil {
		g.publicWarn("github", "Can't set github commit status")
		analytics.Log(ctx).Warnf("Can't set github commit status: %s", err)
	}
}

func fromDBTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
}

func (g githubGo) Process(ctx context.Context) error {
	defer g.exec.Clean()

	var err error
	g.pr, err = g.client.GetPullRequest(ctx, g.context)
	if err != nil {
		return fmt.Errorf("can't get pull request: %s", err)
	}

	g.setCommitStatus(ctx, github.StatusPending, "GolangCI is reviewing your Pull Request...")
	curState, err := g.state.GetState(ctx, g.context.Repo.Owner, g.context.Repo.Name, g.analysisGUID)
	if err != nil {
		analytics.Log(ctx).Warnf("Can't get current state: %s", err)
	} else if curState.Status == "sent_to_queue" {
		g.addTimingFrom("In Queue", fromDBTime(curState.CreatedAt))
		curState.Status = "processing"
		if err = g.state.UpdateState(ctx, g.context.Repo.Owner, g.context.Repo.Name, g.analysisGUID, curState); err != nil {
			analytics.Log(ctx).Warnf("Can't update analysis %s state with setting status to 'processing': %s", g.analysisGUID, err)
		}
	}

	r := g.context.Repo
	wd := path.Join(g.exec.WorkDir(), "src", "github.com", r.Owner, r.Name)
	g.exec = g.exec.WithWorkDir(wd) // XXX: clean gopath, but work in subdir of gopath

	return g.processWithGuaranteedGithubStatus(ctx)
}

func (g *githubGo) trackTiming(name string, f func()) {
	startedAt := time.Now()
	f()
	g.timings = append(g.timings, Timing{
		Name:     name,
		Duration: JSONDuration(time.Since(startedAt)),
	})
}

func (g *githubGo) addTimingFrom(name string, from time.Time) {
	g.timings = append(g.timings, Timing{
		Name:     name,
		Duration: JSONDuration(time.Since(from)),
	})
}

func (g *githubGo) publicWarn(tag string, text string) {
	g.warnings = append(g.warnings, Warning{
		Tag:  tag,
		Text: text,
	})
}

func (d JSONDuration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(time.Duration(d) / time.Millisecond))), nil
}

func (d JSONDuration) String() string {
	return time.Duration(d).String()
}
