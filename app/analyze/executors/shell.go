package executors

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/sirupsen/logrus"
)

type shell struct {
	envStore
	wd string
}

func newShell(workDir string) *shell {
	return &shell{
		wd:       workDir,
		envStore: *newEnvStore(),
	}
}

func (s shell) Run(ctx context.Context, name string, args ...string) (string, error) {
	startedAt := time.Now()
	lines := []string{}
	outReader, finish, err := s.RunAsync(ctx, name, args...)
	if err != nil {
		return "", err
	}

	endCh := make(chan struct{})
	defer close(endCh)

	go func() {
		select {
		case <-ctx.Done():
			analytics.Log(ctx).Warnf("Closing shell reader on timeout")
			if cerr := outReader.Close(); cerr != nil {
				analytics.Log(ctx).Warnf("Failed to close shell reader on deadline: %s", cerr)
			}
		case <-endCh:
		}
	}()

	scanner := bufio.NewScanner(outReader)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "should have comment") && !strings.Contains(line, "comment on exported method") {
			// HACK: get less logs
			logrus.Infof("shell[%s]: %s", s.wd, line)
		}
		lines = append(lines, line)
	}
	if err = scanner.Err(); err != nil {
		analytics.Log(ctx).Warnf("Out lines scanning error: %s", err)
	}

	err = finish()

	logrus.Infof("shell[%s]: %s %v executed for %s: %v", s.wd, name, args, time.Since(startedAt), err)

	// XXX: it's important to not change error here, because it holds exit code
	return strings.Join(lines, "\n"), err
}

type finishFunc func() error

func (s shell) RunAsync(ctx context.Context, name string, args ...string) (io.ReadCloser, finishFunc, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = s.env
	cmd.Dir = s.wd

	logrus.Infof("shell[%s]: %s %v executing...", s.wd, name, args)
	startedAt := time.Now()

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("can't make out pipe: %s", err)
	}

	cmd.Stderr = cmd.Stdout // Set the same pipe
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	logrus.Infof("shell[%s]: %s %v started for %s", s.wd, name, args, time.Since(startedAt))

	return outReader, func() error {
		// XXX: it's important to not change error here, because it holds exit code
		return cmd.Wait()
	}, nil
}
