package linters

import (
	"context"
	"fmt"
	"log"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/golangci/golangci-worker/app/analyze/executors"
	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/analyze/linters/result/processors"
)

type Runner interface {
	Run(ctx context.Context, linters []Linter, exec executors.Executor) (*result.Result, error)
}

type SimpleRunner struct {
	Processors []processors.Processor
}

func (r SimpleRunner) Run(ctx context.Context, linters []Linter, exec executors.Executor) (*result.Result, error) {
	results := []result.Result{}
	for _, linter := range linters {
		res, err := linter.Run(ctx, exec)
		if err != nil {
			analytics.Log(ctx).Warnf("Can't run linter %+v: %s", linter, err)
			continue
		}

		// TODO: don't skip when will store all issues, not only new
		if len(res.Issues) == 0 {
			continue
		}

		results = append(results, *res)
	}

	results, err := r.processResults(results)
	if err != nil {
		return nil, fmt.Errorf("can't process results: %s", err)
	}

	return r.mergeResults(results), nil
}

func (r SimpleRunner) processResults(results []result.Result) ([]result.Result, error) {
	if len(r.Processors) == 0 {
		return results, nil
	}

	for _, p := range r.Processors {
		var err error
		results, err = p.Process(results)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (r SimpleRunner) mergeResults(results []result.Result) *result.Result {
	if len(results) == 0 {
		return nil
	}

	if len(results) > 1 {
		log.Fatalf("len(results) can't be more than 1: %+v", results)
	}

	// TODO: support for multiple linters, not only golangci-lint
	return &results[0]
}
