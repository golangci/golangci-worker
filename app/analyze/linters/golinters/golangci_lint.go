package golinters

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	lintres "github.com/golangci/golangci-lint/pkg/result"
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

	var issues []lintres.Issue
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "[") {
			continue
		}
		if err = json.Unmarshal([]byte(line), &issues); err != nil {
			return nil, fmt.Errorf("can't parse json output '%s' of %s: %s, %s", line, g.Name(), err, out)
		}
		break
	}

	var retIssues []result.Issue
	for _, i := range issues {
		retIssues = append(retIssues, result.Issue{
			File:       i.FilePath(),
			LineNumber: i.Line(),
			Text:       i.Text,
			FromLinter: i.FromLinter,
			HunkPos:    i.HunkPos,
		})
	}
	return &result.Result{
		Issues: retIssues,
	}, nil
}
