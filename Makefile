run_dev:
	godotenv go run app/cmd/golangci-worker/golangci-worker.go

gen:
	mockgen -package httputils -source ./app/utils/httputils/client.go >./app/utils/httputils/mock.go
	mockgen -package github -source ./app/utils/github/client.go >./app/utils/github/client_mock.go
	mockgen -package linters -source ./app/analyze/linters/linter.go >./app/analyze/linters/linter_mock.go
	mockgen -package fetchers -source ./app/analyze/fetchers/fetcher.go >./app/analyze/fetchers/fetcher_mock.go
	mockgen -package reporters -source ./app/analyze/reporters/reporter.go >./app/analyze/reporters/reporter_mock.go
	mockgen -package executors -source ./app/analyze/executors/executor.go >./app/analyze/executors/executor_mock.go
	mockgen -package state -source ./app/analyze/state/storage.go >./app/analyze/state/storage_mock.go

build:
	go build ./app/cmd/...

test:
	go test -v -count 1 ./...
	golangci-lint run -v

test_repo:
	# set env vars PR, REPO
	SLOW_TESTS_ENABLED=1 go test -v ./app/analyze -run TestAnalyzeRepo

test_repo_fake_github:
	# set env vars PR, REPO
	SLOW_TESTS_ENABLED=1 go test -v ./app/analyze/processors -count=1 -run TestProcessRepoWithFakeGithub
