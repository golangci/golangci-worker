package processors

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/fetchers"
	"github.com/golangci/golangci-worker/app/analyze/linters"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/analyze/reporters"
	"github.com/golangci/golangci-worker/app/analyze/state"
	"github.com/golangci/golangci-worker/app/test"
	"github.com/golangci/golangci-worker/app/utils/github"
	gh "github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

var testCtxMatcher = gomock.Any()
var testCtx = analytics.ContextWithEventPropsCollector(context.Background(), analytics.EventPRChecked)

var any = gomock.Any()
var fakeChangedIssue = result.NewIssue("linter2", "F1 issue", "main.go", 10, 11)
var fakeChangedIssues = []result.Issue{
	result.NewIssue("linter2", "F1 issue", "main.go", 9, 10),
	result.NewIssue("linter3", "F1 issue", "main.go", 10, 11),
}

var testSHA = "testSHA"
var testBranch = "testBranch"
var testPR = &gh.PullRequest{
	Head: &gh.PullRequestBranch{
		Ref: gh.String(testBranch),
		SHA: gh.String(testSHA),
	},
	Base: &gh.PullRequestBranch{
		Repo: &gh.Repository{
			Private: gh.Bool(false),
		},
	},
	Number: gh.Int(7),
}
var testAnalysisGUID = "test-guid"

func getFakeLinters(ctrl *gomock.Controller, issues ...result.Issue) []linters.Linter {
	a := linters.NewMockLinter(ctrl)
	a.EXPECT().
		Run(testCtxMatcher, any).
		Return(&result.Result{
			Issues: issues,
		}, nil)
	return []linters.Linter{a}
}

func getNopFetcher(ctrl *gomock.Controller) fetchers.Fetcher {
	f := fetchers.NewMockFetcher(ctrl)
	f.EXPECT().Fetch(testCtxMatcher, "", testBranch, ".", any).Return(nil)
	return f
}

func getNopReporter(ctrl *gomock.Controller) reporters.Reporter {
	r := reporters.NewMockReporter(ctrl)
	r.EXPECT().Report(testCtxMatcher, any, any).AnyTimes().Return(nil)
	return r
}

func getErroredReporter(ctrl *gomock.Controller) reporters.Reporter {
	r := reporters.NewMockReporter(ctrl)
	r.EXPECT().Report(testCtxMatcher, any, any).Return(fmt.Errorf("can't report"))
	return r
}

func getNopState(ctrl *gomock.Controller) state.Storage {
	r := state.NewMockStorage(ctrl)
	r.EXPECT().UpdateState(any, any, any, any, any).AnyTimes().Return(nil)
	r.EXPECT().GetState(any, any, any, any).AnyTimes().Return(&state.State{
		Status: "sent_to_queue",
	}, nil)
	return r
}

func getNopExecutor(ctrl *gomock.Controller) executors.Executor {
	e := executors.NewMockExecutor(ctrl)
	e.EXPECT().WorkDir().Return("").AnyTimes()
	e.EXPECT().WithWorkDir(any).Return(e).AnyTimes()
	e.EXPECT().Run(testCtxMatcher, any, any).Return("", nil).AnyTimes()
	e.EXPECT().Clean().AnyTimes()
	return e
}

func getFakePatch() (string, error) {
	patch, err := ioutil.ReadFile(fmt.Sprintf("test/%d.patch", github.FakeContext.PullRequestNumber))
	return string(patch), err
}

func getFakeStatusGithubClient(ctrl *gomock.Controller, status github.Status, statusDesc string) github.Client {
	c := &github.FakeContext
	gc := github.NewMockClient(ctrl)
	gc.EXPECT().GetPullRequest(testCtxMatcher, c).Return(testPR, nil)

	scsPending := gc.EXPECT().SetCommitStatus(testCtxMatcher, c, testSHA,
		github.StatusPending, "GolangCI is reviewing your Pull Request...", "").
		Return(nil)

	gc.EXPECT().GetPullRequestPatch(any, any).AnyTimes().Return(getFakePatch())

	test.Init()
	url := fmt.Sprintf("%s/r/%s/%s/pulls/%d", os.Getenv("WEB_ROOT"), c.Repo.Owner, c.Repo.Name, testPR.GetNumber())
	gc.EXPECT().SetCommitStatus(testCtxMatcher, c, testSHA, status, statusDesc, url).After(scsPending)

	return gc
}

func getNopGithubClient(ctrl *gomock.Controller) github.Client {
	c := &github.FakeContext

	gc := github.NewMockClient(ctrl)
	gc.EXPECT().CreateReview(any, any, any).AnyTimes()
	gc.EXPECT().GetPullRequest(testCtxMatcher, c).AnyTimes().Return(testPR, nil)
	gc.EXPECT().GetPullRequestPatch(any, any).AnyTimes().Return(getFakePatch())
	gc.EXPECT().SetCommitStatus(any, any, testSHA, any, any, any).AnyTimes()
	return gc
}

func fillWithNops(ctrl *gomock.Controller, cfg *githubGoConfig) {
	if cfg.client == nil {
		cfg.client = getNopGithubClient(ctrl)
	}
	if cfg.exec == nil {
		cfg.exec = getNopExecutor(ctrl)
	}
	if cfg.linters == nil {
		cfg.linters = getFakeLinters(ctrl)
	}
	if cfg.repoFetcher == nil {
		cfg.repoFetcher = getNopFetcher(ctrl)
	}
	if cfg.reporter == nil {
		cfg.reporter = getNopReporter(ctrl)
	}
	if cfg.state == nil {
		cfg.state = getNopState(ctrl)
	}
}

func getNopedProcessor(t *testing.T, ctrl *gomock.Controller, cfg githubGoConfig) *githubGo {
	fillWithNops(ctrl, &cfg)

	p, err := newGithubGo(testCtx, &github.FakeContext, cfg, testAnalysisGUID)
	assert.NoError(t, err)

	return p
}

func testProcessor(t *testing.T, ctrl *gomock.Controller, cfg githubGoConfig) {
	p := getNopedProcessor(t, ctrl, cfg)

	err := p.Process(testCtx)
	assert.NoError(t, err)
}

func TestSetCommitStatusSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, githubGoConfig{
		linters: getFakeLinters(ctrl),
		client:  getFakeStatusGithubClient(ctrl, github.StatusSuccess, "No issues found!"),
	})
}

func TestSetCommitStatusFailureOneIssue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, githubGoConfig{
		linters: getFakeLinters(ctrl, fakeChangedIssue),
		client:  getFakeStatusGithubClient(ctrl, github.StatusFailure, "1 issue found"),
	})
}

func TestSetCommitStatusFailureTwoIssues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, githubGoConfig{
		linters: getFakeLinters(ctrl, fakeChangedIssues...),
		client:  getFakeStatusGithubClient(ctrl, github.StatusFailure, "2 issues found"),
	})
}

func TestSetCommitStatusOnReportingError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := getNopedProcessor(t, ctrl, githubGoConfig{
		linters:  getFakeLinters(ctrl, fakeChangedIssue),
		reporter: getErroredReporter(ctrl),
		client:   getFakeStatusGithubClient(ctrl, github.StatusError, "can't send pull request comments to github"),
	})
	assert.Error(t, p.Process(testCtx))
}

func getRealisticTestProcessor(ctx context.Context, t *testing.T, ctrl *gomock.Controller) *githubGo {
	c := getTestingRepo(t)
	cloneURL := fmt.Sprintf("git@github.com:%s/%s.git", c.Repo.Owner, c.Repo.Name)
	pr := &gh.PullRequest{
		Head: &gh.PullRequestBranch{
			Ref: gh.String("master"),
			Repo: &gh.Repository{
				SSHURL: gh.String(cloneURL),
			},
		},
	}
	gc := github.NewMockClient(ctrl)
	gc.EXPECT().GetPullRequest(testCtxMatcher, c).Return(pr, nil)
	gc.EXPECT().SetCommitStatus(any, any, any, any, any, any).AnyTimes()

	exec, err := executors.NewTempDirShell("gopath")
	assert.NoError(t, err)

	cfg := githubGoConfig{
		exec:     exec,
		runner:   linters.SimpleRunner{},
		reporter: getNopReporter(ctrl),
		client:   gc,
	}

	p, err := newGithubGo(ctx, c, cfg, testAnalysisGUID)
	assert.NoError(t, err)

	return p
}

func TestProcessorTimeout(t *testing.T) {
	test.Init()

	startedAt := time.Now()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(testCtx, 100*time.Millisecond)
	defer cancel()
	p := getRealisticTestProcessor(ctx, t, ctrl)

	assert.Error(t, p.Process(ctx))
	assert.True(t, time.Since(startedAt) < 300*time.Millisecond)
}

func getTestingRepo(t *testing.T) *github.Context {
	repo := os.Getenv("REPO")
	if repo == "" {
		repo = "golangci/golangci-worker"
	}

	repoParts := strings.Split(repo, "/")
	assert.Len(t, repoParts, 2)

	pr := os.Getenv("PR")
	if pr == "" {
		pr = "1"
	}
	prNumber, err := strconv.Atoi(pr)
	assert.NoError(t, err)

	c := &github.Context{
		Repo: github.Repo{
			Owner: repoParts[0],
			Name:  repoParts[1],
		},
		PullRequestNumber: prNumber,
		GithubAccessToken: os.Getenv("TEST_GITHUB_TOKEN"),
	}

	return c
}

func getTestProcessorWithFakeGithub(ctx context.Context, t *testing.T, ctrl *gomock.Controller) *githubGo {
	c := getTestingRepo(t)

	realGc := github.NewMyClient()
	patch, err := realGc.GetPullRequestPatch(ctx, c)
	assert.NoError(t, err)
	pr, err := realGc.GetPullRequest(ctx, c)
	assert.NoError(t, err)

	gc := github.NewMockClient(ctrl)
	gc.EXPECT().GetPullRequestPatch(any, any).AnyTimes().Return(patch, nil)
	gc.EXPECT().GetPullRequest(testCtxMatcher, c).Return(pr, nil)
	gc.EXPECT().SetCommitStatus(any, any, any, any, any, any).AnyTimes()

	cfg := githubGoConfig{
		reporter: getNopReporter(ctrl),
		client:   gc,
	}

	p, err := newGithubGo(ctx, c, cfg, testAnalysisGUID)
	assert.NoError(t, err)

	return p
}

func TestProcessRepoWithFakeGithub(t *testing.T) {
	test.Init()
	test.MarkAsSlow(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := getTestProcessorWithFakeGithub(testCtx, t, ctrl)
	err := p.Process(testCtx)
	assert.NoError(t, err)
}
