package executors

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/golangci/golangci-worker/app/lib/timeutils"
)

type Docker struct {
	image         string
	wd            string
	containerName string
	*envStore
}

func NewDocker(ctx context.Context) (*Docker, error) {
	d := &Docker{
		image:         "golangci_executor",
		envStore:      newEnvStoreNoOS(),
		wd:            "/app/go",
		containerName: "golangci_executor",
	}

	_ = exec.CommandContext(ctx, "docker", "rm", "-f", "-v", d.containerName).Run()

	out, err := exec.CommandContext(ctx, "docker", "run", "-d", "--name", d.containerName, d.image).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("can't run docker: %s, %s", err, out)
	}

	return d, nil
}

var _ Executor = Docker{}

func (d Docker) Run(ctx context.Context, name string, args ...string) (string, error) {
	// XXX: don't use docker sdk because it's too heavyweight: dep ensure takes minutes on it
	dockerArgs := []string{
		"exec",
		"-w", d.wd,
	}

	for _, e := range d.env {
		dockerArgs = append(dockerArgs, "-e", e)
	}

	dockerArgs = append(dockerArgs, d.containerName, name)
	dockerArgs = append(dockerArgs, args...)

	defer timeutils.Track(time.Now(), "docker full execution: docker %v", dockerArgs)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("can't execute command docker %s: %s, %s, %s",
			sprintArgs(dockerArgs), err, string(out), stderrBuf.String())
	}

	return string(out), nil
}

func (d Docker) WithEnv(k, v string) Executor {
	dCopy := d
	dCopy.SetEnv(k, v)
	return dCopy
}

func (d Docker) Clean() {
	_ = exec.Command("docker", "rm", "-f", "-v", d.containerName).Run()
}

func (d Docker) WithWorkDir(wd string) Executor {
	dCopy := d
	dCopy.wd = wd
	return dCopy
}

func (d Docker) WorkDir() string {
	return d.wd
}

func (d Docker) CopyFile(ctx context.Context, dst, src string) error {
	if !filepath.IsAbs(dst) {
		dst = filepath.Join(d.WorkDir(), dst)
	}
	cmd := exec.CommandContext(ctx, "docker", "cp", src, fmt.Sprintf("%s:%s", d.containerName, dst))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("can't docker cp %s to %s: %s, %s", src, dst, err, out)
	}

	return nil
}
