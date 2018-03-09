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

func TestErrcheckIgnoreClose(t *testing.T) {
	const source = `package p

	import "os"

	func ok() error {
		f, err := os.Open("t.go")
		if err != nil {
			return err
		}

		f.Close()
		return nil
	}
`

	test.ExpectIssues(t, errCheck, source, []result.Issue{})
}

// TODO: add cases of non-compiling code
// TODO: don't report issues if got more than 20 issues
