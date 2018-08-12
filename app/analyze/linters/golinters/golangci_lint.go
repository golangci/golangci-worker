package golinters

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/golangci/golangci-lint/pkg/printers"
	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/lib/errorutils"
	"github.com/golangci/golangci-worker/app/lib/executors"
)

type golangciLint struct {
}

func (g golangciLint) Name() string {
	return "golangci-lint"
}

func (g golangciLint) Run(ctx context.Context, exec executors.Executor) (*result.Result, error) {
	exec = exec.WithEnv("GOLANGCI_COM_RUN", "1")

	// TODO: check golangci-lint warnings in stderr
	out, runErr := exec.Run(ctx,
		g.Name(),
		"run",
		"--out-format=json",
		"--issues-exit-code=0",
		"--print-welcome=false",
		"--deadline=5m",
		"--new-from-patch=../../../../changes.patch",
		filepath.Join(exec.WorkDir(), "..."),
	)
	rawJSON := []byte(out)

	if runErr != nil {
		var res printers.JSONResult
		if jsonErr := json.Unmarshal(rawJSON, &res); jsonErr == nil && res.Report.Error != "" {
			return nil, &errorutils.InternalError{
				PublicDesc:  fmt.Sprintf("can't run golangci-lint: %s", res.Report.Error),
				PrivateDesc: fmt.Sprintf("can't run golangci-lint: %s, %s", res.Report.Error, runErr),
			}
		}

		return nil, &errorutils.InternalError{
			PublicDesc:  "can't run golangci-lint",
			PrivateDesc: fmt.Sprintf("can't run golangci-lint: %s, %s", runErr, out),
		}
	}

	var res printers.JSONResult
	if jsonErr := json.Unmarshal(rawJSON, &res); jsonErr != nil {
		return nil, &errorutils.InternalError{
			PublicDesc:  "can't run golangci-lint: invalid output json",
			PrivateDesc: fmt.Sprintf("can't run golangci-lint: can't parse json output %s: %s", out, jsonErr),
		}
	}

	if res.Report != nil && len(res.Report.Warnings) != 0 {
		analytics.Log(ctx).Warnf("Got golangci-lint warnings: %#v", res.Report.Warnings)
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
		ResultJSON: json.RawMessage(rawJSON),
	}, nil
}
