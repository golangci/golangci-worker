run_dev:
	godotenv go run app/cmd/golangci-worker/golangci-worker.go

gen:
	go generate ./...

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