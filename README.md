[![CircleCI](https://circleci.com/gh/golangci/golangci-worker.svg?style=svg&circle-token=94e0eb37b49bb5f87364a50592794eba13f0d95d)](https://circleci.com/gh/golangci/golangci-worker)
[![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-worker.svg)](https://golangci.com)

# Worker

This repository contains code of queue worker. Worker runs golangci-lint and reports result to GitHub.

## Development

### Technologies

Go (golang), heroku, circleci, docker, redis, postgres.

### Preparation

In [golangci-api](https://github.com/golangci/golangci-api) repo run:

```bash
docker-compose up -d
```

It runs postgres and redis needed for both api and worker.

### How to run worker

```bash
make run_dev
```

### How to run once on GitHub repo without changing GitHub data: commit status, comments

```bash
REPO={OWNER/NAME} PR={PULL_REQUEST_NUMBER} make test_repo_fake_github
```

e.g. `REPO=golangci/golangci-worker PR=39 make test_repo_fake_github`

### Configuration

Configurate via `.env` file. Dev `.env` may be like this:

```bash
REDIS_URL="redis://localhost:6379"
API_URL="https://api.dev.golangci.com"
WEB_ROOT="https://dev.golangci.com"
USE_CONTAINER_EXECUTOR=1
```

### Executors

TODO

```bash
echo "USE_CONTAINER_EXECUTOR=1" >>.env
```

### API

golangci-api is not needed for running and testing golangci-worker. Not running api can just make log warnings like this:

```bash
level=warning msg="Can't get current state: bad status code 404"
```

### Testing

To run tests:

```bash
make test
```

For more realistic testing than `test_repo_fake_github` use in golangci-api repo GitHub WebHook emulator:

```bash
go run app/scripts/emulate_webhook/main.go --repo golangci/golangci-worker --pr 39 --sha fbd0d7bada8a6cfa7adbc58e5901e0d66f7f65b1
```

## Contributing

See [CONTRIBUTING](https://github.com/golangci/golangci-worker/blob/master/CONTRIBUTING.md).
