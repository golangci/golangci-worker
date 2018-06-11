package executors

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/sirupsen/logrus"
)

type RemoteShell struct {
	envStore
	tempWorkDir             string
	wd                      string
	user, host, keyFilePath string
}

var _ Executor = &RemoteShell{}

func NewRemoteShell(user, host, keyFilePath string) *RemoteShell {
	return &RemoteShell{
		envStore:    envStore{},
		user:        user,
		host:        host,
		keyFilePath: keyFilePath,
	}
}

func (s *RemoteShell) SetupTempWorkDir(ctx context.Context) error {
	out, err := s.Run(ctx, "mktemp", "-d")
	if err != nil {
		return err
	}

	s.tempWorkDir = strings.TrimSpace(out)
	if s.tempWorkDir == "" {
		return fmt.Errorf("empty temp dir")
	}

	s.wd = s.tempWorkDir

	return nil
}

func quoteArgs(args []string) []string {
	var ret []string
	for _, arg := range args {
		ret = append(ret, strconv.Quote(arg))
	}
	return ret
}

func sprintArgs(args []string) string {
	return strings.Join(quoteArgs(args), " ")
}

func (s RemoteShell) Run(ctx context.Context, name string, srcArgs ...string) (string, error) {
	shellArg := fmt.Sprintf("cd %s; %s %s %s",
		s.wd,
		strings.Join(s.env, " "),
		name, strings.Join(srcArgs, " "))
	args := []string{
		"-i",
		s.keyFilePath,
		fmt.Sprintf("%s@%s", s.user, s.host),
		shellArg,
	}

	cmd := exec.CommandContext(ctx, "ssh", args...)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	out, err := cmd.Output()
	stderr := stderrBuf.String()
	if err != nil {
		return "", fmt.Errorf("can't execute command ssh %s: %s, %s, %s",
			sprintArgs(args), err, string(out), stderr)
	}

	if stderr != "" {
		processStderr(ctx, stderr)
	}

	return string(out), nil
}

func processStderr(ctx context.Context, stderr string) {
	skipPatterns := []string{
		// TODO: it should be reported to user in details of analysis
		"Can't run megacheck because of compilation errors in ",
	}

	lines := strings.Split(stderr, "\n")
	var allowedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		skip := false
		for _, sp := range skipPatterns {
			if strings.Contains(line, sp) {
				skip = true
				break
			}
		}
		if !skip {
			allowedLines = append(allowedLines, line)
		}
	}

	if len(allowedLines) == 0 {
		return
	}

	analytics.Log(ctx).Warnf("golangci-lint warnings: %q", strings.Join(allowedLines, "\n"))
}

func (s RemoteShell) CopyFile(ctx context.Context, dst, src string) error {
	dst = filepath.Join(s.WorkDir(), dst)
	out, err := exec.CommandContext(ctx, "scp",
		"-i", s.keyFilePath,
		src,
		fmt.Sprintf("%s@%s:%s", s.user, s.host, dst),
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("can't copy file %s to %s: %s, %s", src, dst, err, out)
	}

	return nil
}

func (s RemoteShell) Clean() {
	if s.tempWorkDir == "" {
		panic("empty temp work dir")
	}

	out, err := s.Run(context.TODO(), "rm", "-r", s.tempWorkDir)
	if err != nil {
		logrus.Warnf("Can't remove temp work dir in remote shell: %s, %s", err, out)
	}
}

func (s RemoteShell) WithEnv(k, v string) Executor {
	eCopy := s
	eCopy.SetEnv(k, v)
	return &eCopy
}

func (s RemoteShell) WorkDir() string {
	return s.wd
}

func (s *RemoteShell) SetWorkDir(wd string) {
	s.wd = wd
}

func (s RemoteShell) WithWorkDir(wd string) Executor {
	eCopy := s
	eCopy.wd = wd
	return &eCopy
}
