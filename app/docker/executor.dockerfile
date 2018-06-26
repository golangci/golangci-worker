# docker build -t golangci_executor -f app/docker/executor.dockerfile .
FROM golang:1.10

WORKDIR /app

ENV GOPATH=/app/go
ENV GOBINPATH=$GOPATH/bin
ENV PATH=$PATH:/usr/local/go/bin:$GOBINPATH

RUN go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
COPY ./app/scripts/ensure_deps.sh /app/ensure_deps.sh
COPY ./app/scripts/forever_run.sh /app/run.sh

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN curl https://glide.sh/get | sh
RUN go get github.com/tools/godep
RUN go get github.com/kardianos/govendor

CMD ["/app/run.sh"]
