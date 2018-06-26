[![CircleCI](https://circleci.com/gh/golangci/golangci-worker.svg?style=svg&circle-token=94e0eb37b49bb5f87364a50592794eba13f0d95d)](https://circleci.com/gh/golangci/golangci-worker)
[![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-worker.svg)](https://golangci.com)

## Worker
This repository contains code of queue worker. Worker runs golangci-lint and reports result to GitHub.

## Development
### Preparation
In [golangci-api](https://github.com/golangci/golangci-api) repo run:
```
docker-compose up -d
```
It runs postgres and redis needed for both api and worker.

### How to run worker
```bash
make run_dev
```

### How to run once on GitHub repo without changing GitHub data: commit status, comments
```
REPO={OWNER/NAME} PR={PULL_REQUEST_NUMBER} make test_repo_fake_github
```

### Configuration
Configurate via `.env` file. Dev `.env` may be like this:
```
REDIS_URL="redis://localhost:6379"
API_URL="https://api.dev.golangci.com"
WEB_ROOT="https://dev.golangci.com"
USE_DOCKER_EXECUTOR=1
```

## Executors
Executer is an abstraction over executing shell commands. In production we use remote shell executor (machine by ssh).
For local development it's better to use docker executor:
```
docker build -t golangci_executor -f app/docker/executor.dockerfile .
echo "USE_DOCKER_EXECUTOR=1" >>.env
```


# Contributing
See [CONTRIBUTING](https://github.com/golangci/golangci-worker/blob/master/CONTRIBUTING.md).

