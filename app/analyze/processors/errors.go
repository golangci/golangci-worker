package processors

import (
	"strings"

	"github.com/golangci/golangci-worker/app/lib/github"
)

type IgnoredError struct {
	Status        github.Status
	StatusDesc    string
	IsRecoverable bool
}

func (e IgnoredError) Error() string {
	return e.StatusDesc
}

func escapeErrorText(text string, secrets map[string]string) string {
	ret := text
	for secret, replacement := range secrets {
		ret = strings.Replace(ret, secret, replacement, -1)
	}

	return ret
}
