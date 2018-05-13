package golinters

import "github.com/golangci/golangci-worker/app/analyze/linters"

func GetSupportedLinters() []linters.Linter {
	return []linters.Linter{golangciLint{}}
}
