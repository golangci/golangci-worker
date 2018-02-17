package golinters

import (
	"testing"

	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/test"
)

func TestErrcheckSimple(t *testing.T) {
	const source = `package p

	func retErr() error {
		return nil
	}

	func missedErrorCheck() {
		retErr()
	}
`

	test.ExpectIssues(t, errCheck, source, []result.Issue{test.NewIssue("errcheck", "Error return value is not checked", 8)})
}

// TODO: add cases of non-compiling code
// TODO: don't report issues if got more than 20 issues
// TODO: dont' report for `defer f.Close()`
