package golinters

import (
	"testing"

	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/test"
)

func TestGovetSimple(t *testing.T) {
	const source = `package p

import "os"

func f() error {
  return &os.PathError{"first", "path", os.ErrNotExist}
}
`

	test.ExpectIssues(t, govet, source, []result.Issue{
		test.NewIssue("govet", "os.PathError composite literal uses unkeyed fields", 6),
	})
}
