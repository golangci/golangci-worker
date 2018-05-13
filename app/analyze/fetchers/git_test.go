package fetchers

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/stretchr/testify/assert"
)

func TestGitOnTestRepo(t *testing.T) {
	g := Git{}
	ref := "test-branch"
	cloneURL := "git@github.com:golangci/test.git"

	exec, err := executors.NewTempDirShell("test.git")
	assert.NoError(t, err)
	defer exec.Clean()

	err = g.Fetch(context.Background(), cloneURL, ref, "src", exec)
	assert.NoError(t, err)

	files, err := ioutil.ReadDir(path.Join(exec.WorkDir(), "src"))
	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Equal(t, ".git", files[0].Name())
	assert.Equal(t, "README.md", files[1].Name())
	assert.Equal(t, "main.go", files[2].Name())
}
