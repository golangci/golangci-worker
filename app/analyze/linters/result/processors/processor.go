package processors

import "github.com/golangci/golangci-worker/app/analyze/linters/result"

type Processor interface {
	Process(results []result.Result) ([]result.Result, error)
	Name() string
}
