package golinters

import (
	"testing"

	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/test"
)

func TestGolintSimple(t *testing.T) {
	const source = `package p
	var v_1 string`

	test.ExpectIssues(t, golint, source,
		[]result.Issue{test.NewIssue("golint", "don't use underscores in Go names; var v_1 should be v1", 2)})
}
