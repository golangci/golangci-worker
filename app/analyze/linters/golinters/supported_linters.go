package golinters

import "github.com/golangci/golangci-worker/app/analyze/linters"

const pathLineColMessage = `^(?P<path>.*?\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<message>.*)$`

var errCheck = newLinter("errcheck",
	newLinterConfig(
		"Error return value is not checked",
		pathLineColMessage),
)

var golint = newLinter("golint", newLinterConfig("", pathLineColMessage))

func GetSupportedLinters() []linters.Linter {
	return []linters.Linter{gofmt{}, gofmt{useGoimports: true}, golint, errCheck}
}
