package executors

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type TempDirShell struct {
	shell
}

var _ Executor = &TempDirShell{}

var tmpRoot string

func init() {
	var err error
	tmpRoot, err = filepath.EvalSymlinks("/tmp")
	if err != nil {
		log.Fatalf("can't eval symlinks on /tmp: %s", err)
	}
}

func NewTempDirShell(tag string) (*TempDirShell, error) {
	wd, err := ioutil.TempDir(tmpRoot, fmt.Sprintf("golangci.%s", tag))
	if err != nil {
		return nil, fmt.Errorf("can't make temp dir: %s", err)
	}

	return &TempDirShell{
		shell: *newShell(wd),
	}, nil
}

func (s TempDirShell) WorkDir() string {
	return s.wd
}

func (s *TempDirShell) SetWorkDir(wd string) {
	s.wd = wd
}

func (s TempDirShell) Clean() {
	if err := os.RemoveAll(s.wd); err != nil {
		logrus.Warnf("Can't remove temp dir %s: %s", s.wd, err)
	}
}

func (s TempDirShell) WithEnv(k, v string) Executor {
	eCopy := s
	eCopy.SetEnv(k, v)
	return &eCopy
}

func (s TempDirShell) WithWorkDir(wd string) Executor {
	eCopy := s
	eCopy.wd = wd
	return &eCopy
}
