package processors

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/fetchers"
	"github.com/golangci/golangci-worker/app/analyze/linters"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	lp "github.com/golangci/golangci-worker/app/analyze/linters/result/processors"
	"github.com/golangci/golangci-worker/app/analyze/reporters"
	"github.com/golangci/golangci-worker/app/test"
	"github.com/golangci/golangci-worker/app/utils/fsutils"
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
}

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

func getFakeReporter(ctrl *gomock.Controller, expIssues ...result.Issue) reporters.Reporter {
	r := reporters.NewMockReporter(ctrl)
	r.EXPECT().Report(testCtxMatcher, testSHA, expIssues).Return(nil)
	return r
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

func getFakeExecutor(ctrl *gomock.Controller) executors.Executor {
	gopathE := executors.NewMockExecutor(ctrl)
	repoE := executors.NewMockExecutor(ctrl)

	tmpDir := path.Join("/", "tmp", "golangci")
	wdCall := gopathE.EXPECT().WorkDir().Return(tmpDir)

	wwdCall := gopathE.EXPECT().
		WithWorkDir(path.Join(tmpDir, "src", "github.com", "owner", "name")).
		Return(repoE).After(wdCall)

	runCall := repoE.EXPECT().
		Run(testCtxMatcher, "bash", path.Join(fsutils.GetProjectRoot(), "app", "scripts", "ensure_deps.sh")).
		Return("", nil).After(wwdCall)

	gopathE.EXPECT().Clean().After(runCall)
	return gopathE
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

func getFakeStatusGithubClient(t *testing.T, ctrl *gomock.Controller, status github.Status, statusDesc string) github.Client {
	c := &github.FakeContext
	gc := github.NewMockClient(ctrl)
	gc.EXPECT().GetPullRequest(testCtxMatcher, c).Return(testPR, nil)

	scsPending := gc.EXPECT().SetCommitStatus(testCtxMatcher, c, testSHA, github.StatusPending, "GolangCI is reviewing your Pull Request...").
		Return(nil)

	gc.EXPECT().GetPullRequestPatch(any, any).AnyTimes().Return(getFakePatch())

	gc.EXPECT().SetCommitStatus(testCtxMatcher, c, testSHA, status, statusDesc).After(scsPending)

	return gc
}

func getNopGithubClient(ctrl *gomock.Controller) github.Client {
	c := &github.FakeContext

	gc := github.NewMockClient(ctrl)
	gc.EXPECT().CreateReview(any, any, any).AnyTimes()
	gc.EXPECT().GetPullRequest(testCtxMatcher, c).AnyTimes().Return(testPR, nil)
	gc.EXPECT().GetPullRequestPatch(any, any).AnyTimes().Return(getFakePatch())
	gc.EXPECT().SetCommitStatus(any, any, testSHA, any, any).AnyTimes()
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
}

func getNopedProcessor(t *testing.T, ctrl *gomock.Controller, cfg githubGoConfig) *githubGo {
	fillWithNops(ctrl, &cfg)

	p, err := newGithubGo(testCtx, &github.FakeContext, cfg)
	assert.NoError(t, err)

	return p
}

func testProcessor(t *testing.T, ctrl *gomock.Controller, cfg githubGoConfig) {
	p := getNopedProcessor(t, ctrl, cfg)

	err := p.Process(testCtx)
	assert.NoError(t, err)
}

func getGithubProcessorWithIssues(t *testing.T, ctrl *gomock.Controller,
	issues, expIssues []result.Issue) *githubGo {

	cfg := githubGoConfig{
		linters:     getFakeLinters(ctrl, issues...),
		repoFetcher: getNopFetcher(ctrl),
		reporter:    getFakeReporter(ctrl, expIssues...),
		exec:        getFakeExecutor(ctrl),
		client:      getNopGithubClient(ctrl),
	}

	p, err := newGithubGo(testCtx, &github.FakeContext, cfg)
	assert.NoError(t, err)
	return p
}

func TestNewIssuesFiltering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	issues := []result.Issue{
		result.NewIssue("linter1", "F0 issue", "main.go", 6, 0), // must be filtered out because not changed
		fakeChangedIssue,
	}
	p := getGithubProcessorWithIssues(t, ctrl, issues, issues[1:])
	assert.NoError(t, p.Process(testCtx))
}

func TestOnlyOneIssuePerLine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	issues := []result.Issue{
		result.NewIssue("linter1", "F1 issue", "main.go", 10, 11),
		result.NewIssue("linter2", "F1 another issue", "main.go", 10, 11),
	}
	p := getGithubProcessorWithIssues(t, ctrl, issues, issues[:1])
	assert.NoError(t, p.Process(testCtx))
}

func TestExcludeGolintCommentsIssues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	issues := []result.Issue{
		result.NewIssue("golint", "exported function SetProcessorFactory should have comment or be unexported", "main.go", 1, 1),
		result.NewIssue("golint", "exported const StatusPending should have comment (or a comment on this block) or be unexported", "main.go", 10, 11),
	}
	p := getGithubProcessorWithIssues(t, ctrl, issues, []result.Issue{})
	assert.NoError(t, p.Process(testCtx))
}

func TestSetCommitStatusSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, githubGoConfig{
		linters: getFakeLinters(ctrl),
		client:  getFakeStatusGithubClient(t, ctrl, github.StatusSuccess, "No issues found!"),
	})
}

func TestSetCommitStatusFailureOneIssue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, githubGoConfig{
		linters: getFakeLinters(ctrl, fakeChangedIssue),
		client:  getFakeStatusGithubClient(t, ctrl, github.StatusFailure, "1 issue found"),
	})
}

func TestSetCommitStatusFailureTwoIssues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, githubGoConfig{
		linters: getFakeLinters(ctrl, fakeChangedIssues...),
		client:  getFakeStatusGithubClient(t, ctrl, github.StatusFailure, "2 issues found"),
	})
}

func TestSetCommitStatusSuccessOnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := getNopedProcessor(t, ctrl, githubGoConfig{
		linters:  getFakeLinters(ctrl, fakeChangedIssue),
		reporter: getErroredReporter(ctrl),
		client:   getFakeStatusGithubClient(t, ctrl, github.StatusSuccess, "No issues found!"),
	})
	assert.Error(t, p.Process(testCtx))
}

func TestProcessorTimeout(t *testing.T) {
	startedAt := time.Now()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(testCtx, 100*time.Millisecond)
	defer cancel()
	p := getRealisticTestProcessor(ctx, t, ctrl)

	assert.Error(t, p.Process(ctx))
	assert.True(t, time.Since(startedAt) < 300*time.Millisecond)
}

func getLinterProcessorsExceptDiff() []lp.Processor {
	ret := []lp.Processor{}
	dp := lp.NewDiffProcessor("")
	for _, p := range getLinterProcessors(testCtx, "") {
		if p.Name() != dp.Name() {
			ret = append(ret, p)
		}
	}
	return ret
}

func getTestingRepo(t *testing.T) (*github.Context, string) {
	repo := os.Getenv("REPO")
	if repo == "" {
		repo = "golangci/golangci-worker"
	}

	branch := os.Getenv("BRANCH")
	if branch == "" {
		branch = "master"
	}

	repoParts := strings.Split(repo, "/")
	assert.Len(t, repoParts, 2)

	c := &github.Context{
		Repo: github.Repo{
			Owner: repoParts[0],
			Name:  repoParts[1],
		},
	}

	return c, branch
}

func getRealisticTestProcessor(ctx context.Context, t *testing.T, ctrl *gomock.Controller) *githubGo {
	c, branch := getTestingRepo(t)
	cloneURL := fmt.Sprintf("git@github.com:%s/%s.git", c.Repo.Owner, c.Repo.Name)
	pr := &gh.PullRequest{
		Head: &gh.PullRequestBranch{
			Ref: gh.String(branch),
			Repo: &gh.Repository{
				SSHURL: gh.String(cloneURL),
			},
		},
	}

	gc := github.NewMockClient(ctrl)
	gc.EXPECT().GetPullRequest(testCtxMatcher, c).Return(pr, nil)
	gc.EXPECT().SetCommitStatus(any, any, any, any, any).AnyTimes()

	cfg := githubGoConfig{
		runner: linters.SimpleRunner{
			Processors: getLinterProcessorsExceptDiff(),
		},
		reporter: getNopReporter(ctrl),
		client:   gc,
	}

	p, err := newGithubGo(ctx, c, cfg)
	assert.NoError(t, err)

	return p
}

func TestRunProcessorOnRepo(t *testing.T) {
	test.MarkAsSlow(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := getRealisticTestProcessor(testCtx, t, ctrl)
	err := p.Process(testCtx)
	assert.NoError(t, err)
}
