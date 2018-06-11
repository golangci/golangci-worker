package golinters

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golangci/golangci-lint/pkg/printers"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
)

type golangciLint struct {
}

func (g golangciLint) Name() string {
	return "golangci-lint"
}

func (g golangciLint) Run(ctx context.Context, exec executors.Executor) (*result.Result, error) {
	exec = exec.WithEnv("GOLANGCI_COM_RUN", "1")

	// TODO: check golangci-lint warnings in stderr
	out, err := exec.Run(ctx,
		g.Name(),
		"run",
		"--out-format=json",
		"--issues-exit-code=0",
		"--print-welcome=false",
		"--deadline=5m",
		"--new-from-patch=../../../../changes.patch",
		filepath.Join(exec.WorkDir(), "..."),
	)
	if err != nil {
		return nil, fmt.Errorf("can't run %s: %s, %s", g.Name(), err, out)
	}

	var res printers.JSONResult
	rawJSON := []byte(out)
	if strings.HasPrefix(out, "[") {
		// old format
		if err = json.Unmarshal(rawJSON, &res.Issues); err != nil {
			return nil, fmt.Errorf("can't parse json output '%s' of %s: %s", out, g.Name(), err)
		}
	} else {
		if err = json.Unmarshal(rawJSON, &res); err != nil {
			return nil, fmt.Errorf("can't parse json output '%s' of %s: %s", out, g.Name(), err)
		}
	}

	var retIssues []result.Issue
	for _, i := range res.Issues {
		retIssues = append(retIssues, result.Issue{
			File:       i.FilePath(),
			LineNumber: i.Line(),
			Text:       i.Text,
			FromLinter: i.FromLinter,
			HunkPos:    i.HunkPos,
		})
	}
	return &result.Result{
		Issues:     retIssues,
		ResultJSON: res,
	}, nil
}
