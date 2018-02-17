package processors

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/fetchers"
	"github.com/golangci/golangci-worker/app/analyze/linters"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/analyze/reporters"
	"github.com/golangci/golangci-worker/app/utils/fsutils"
	"github.com/golangci/golangci-worker/app/utils/github"
	gh "github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

var testCtx = gomock.Any()
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
		Run(testCtx, any).
		Return(&result.Result{
			Issues: issues,
		}, nil)
	return []linters.Linter{a}
}

func getNopFetcher(ctrl *gomock.Controller) fetchers.Fetcher {
	f := fetchers.NewMockFetcher(ctrl)
	f.EXPECT().Fetch(testCtx, "", testBranch, ".", any).Return(nil)
	return f
}

func getFakeReporter(ctrl *gomock.Controller, expIssues ...result.Issue) reporters.Reporter {
	r := reporters.NewMockReporter(ctrl)
	r.EXPECT().Report(testCtx, testSHA, expIssues).Return(nil)
	return r
}

func getNopReporter(ctrl *gomock.Controller) reporters.Reporter {
	r := reporters.NewMockReporter(ctrl)
	r.EXPECT().Report(testCtx, any, any).Return(nil)
	return r
}

func getErroredReporter(ctrl *gomock.Controller) reporters.Reporter {
	r := reporters.NewMockReporter(ctrl)
	r.EXPECT().Report(testCtx, any, any).Return(fmt.Errorf("can't report"))
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
		Run(testCtx, "bash", path.Join(fsutils.GetProjectRoot(), "app", "scripts", "ensure_deps.sh")).
		Return("", nil).After(wwdCall)

	gopathE.EXPECT().Clean().After(runCall)
	return gopathE
}

func getNopExecutor(ctrl *gomock.Controller) executors.Executor {
	e := executors.NewMockExecutor(ctrl)
	e.EXPECT().WorkDir().Return("").AnyTimes()
	e.EXPECT().WithWorkDir(any).Return(e).AnyTimes()
	e.EXPECT().Run(testCtx, any, any).Return("", nil).AnyTimes()
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
	gc.EXPECT().GetPullRequest(testCtx, c).Return(testPR, nil)

	scsPending := gc.EXPECT().SetCommitStatus(testCtx, c, testSHA, github.StatusPending, "GolangCI is reviewing your Pull Request...").
		Return(nil)

	gc.EXPECT().GetPullRequestPatch(any, any).AnyTimes().Return(getFakePatch())

	gc.EXPECT().SetCommitStatus(testCtx, c, testSHA, status, statusDesc).After(scsPending)

	return gc
}

func getNopGithubClient(ctrl *gomock.Controller) github.Client {
	c := &github.FakeContext

	gc := github.NewMockClient(ctrl)
	gc.EXPECT().CreateReview(any, any, any).AnyTimes()
	gc.EXPECT().GetPullRequest(testCtx, c).Return(testPR, nil).AnyTimes()
	gc.EXPECT().GetPullRequestPatch(any, any).AnyTimes().Return(getFakePatch())
	gc.EXPECT().SetCommitStatus(any, any, testSHA, any, any).AnyTimes()
	return gc
}

func fillWithNops(ctrl *gomock.Controller, p *githubGo) {
	if p.client == nil {
		p.client = getNopGithubClient(ctrl)
	}
	if p.context == nil {
		p.context = &github.FakeContext
	}
	if p.exec == nil {
		p.exec = getNopExecutor(ctrl)
	}
	if p.linters == nil {
		p.linters = getFakeLinters(ctrl)
	}
	if p.repoFetcher == nil {
		p.repoFetcher = getNopFetcher(ctrl)
	}
	if p.reporter == nil {
		p.reporter = getNopReporter(ctrl)
	}
}

func testProcessor(t *testing.T, ctrl *gomock.Controller, p *githubGo) {
	fillWithNops(ctrl, p)
	assert.NoError(t, p.Process(context.Background()))
}

func getGithubProcessorWithIssues(t *testing.T, ctrl *gomock.Controller,
	issues, expIssues []result.Issue) *githubGo {

	li := getFakeLinters(ctrl, issues...)
	f := getNopFetcher(ctrl)
	r := getFakeReporter(ctrl, expIssues...)
	e := getFakeExecutor(ctrl)
	gc := getNopGithubClient(ctrl)

	return newGithubGo(f, li, r, e, gc, &github.FakeContext)
}

func TestNewIssuesFiltering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	issues := []result.Issue{
		result.NewIssue("linter1", "F0 issue", "main.go", 6, 0), // must be filtered out because not changed
		fakeChangedIssue,
	}
	p := getGithubProcessorWithIssues(t, ctrl, issues, issues[1:])
	assert.NoError(t, p.Process(context.Background()))
}

func TestOnlyOneIssuePerLine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	issues := []result.Issue{
		result.NewIssue("linter1", "F1 issue", "main.go", 10, 11),
		result.NewIssue("linter2", "F1 another issue", "main.go", 10, 11),
	}
	p := getGithubProcessorWithIssues(t, ctrl, issues, issues[:1])
	assert.NoError(t, p.Process(context.Background()))
}

func TestExcludeGolintCommentsIssues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	issues := []result.Issue{
		result.NewIssue("golint", "exported function SetProcessorFactory should have comment or be unexported", "main.go", 1, 1),
		result.NewIssue("golint", "exported const StatusPending should have comment (or a comment on this block) or be unexported", "main.go", 10, 11),
	}
	p := getGithubProcessorWithIssues(t, ctrl, issues, []result.Issue{})
	assert.NoError(t, p.Process(context.Background()))
}

func TestSetCommitStatusSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, &githubGo{
		linters: getFakeLinters(ctrl),
		client:  getFakeStatusGithubClient(t, ctrl, github.StatusSuccess, "No issues found!"),
	})
}

func TestSetCommitStatusFailureOneIssue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, &githubGo{
		linters: getFakeLinters(ctrl, fakeChangedIssue),
		client:  getFakeStatusGithubClient(t, ctrl, github.StatusFailure, "1 issue found"),
	})
}

func TestSetCommitStatusFailureTwoIssues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testProcessor(t, ctrl, &githubGo{
		linters: getFakeLinters(ctrl, fakeChangedIssues...),
		client:  getFakeStatusGithubClient(t, ctrl, github.StatusFailure, "2 issues found"),
	})
}

func TestSetCommitStatusSuccessOnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := githubGo{
		linters:  getFakeLinters(ctrl, fakeChangedIssue),
		reporter: getErroredReporter(ctrl),
		client:   getFakeStatusGithubClient(t, ctrl, github.StatusSuccess, "No issues found!"),
	}
	fillWithNops(ctrl, &p)
	assert.Error(t, p.Process(context.Background()))
}
