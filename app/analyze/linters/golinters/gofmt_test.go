package golinters

import (
	"testing"

	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/test"
)

func TestGofmtIssueFound(t *testing.T) {
	const source = `package p

func noFmt() error {
return nil
}
`

	test.ExpectIssues(t, gofmt{}, source, []result.Issue{test.NewIssue("gofmt", "File is not gofmt-ed with -s", 4)})
}

func TestGofmtNoIssue(t *testing.T) {
	const source = `package p

func fmted() error {
	return nil
}
`

	test.ExpectIssues(t, gofmt{}, source, []result.Issue{})
}

func TestGoimportsIssueFound(t *testing.T) {
	const source = `package p
func noFmt() error {return nil}
`

	lint := gofmt{useGoimports: true}
	test.ExpectIssues(t, lint, source, []result.Issue{test.NewIssue("goimports", "File is not goimports-ed", 2)})
}

func TestGoimportsNoIssue(t *testing.T) {
	const source = `package p

func fmted() error {
	return nil
}
`

	lint := gofmt{useGoimports: true}
	test.ExpectIssues(t, lint, source, []result.Issue{})
}
