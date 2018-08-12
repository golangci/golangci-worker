package processors

import "github.com/golangci/golangci-worker/app/lib/github"

type IgnoredError struct {
	Status        github.Status
	StatusDesc    string
	IsRecoverable bool
}

func (e IgnoredError) Error() string {
	return e.StatusDesc
}
