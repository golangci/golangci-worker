# docker build -t golangci_executor -f app/docker/executor.dockerfile .
FROM heroku/heroku:16

ENV GO_VERSION=1.9.3
ENV OS=linux
ENV ARCH=amd64

WORKDIR /app
RUN wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz -O - | tar -C /usr/local -xzf -

ENV GOPATH=/app/go
ENV GOBINPATH=$GOPATH/bin
ENV PATH=$PATH:/usr/local/go/bin:$GOBINPATH

#RUN wget https://s3-us-west-2.amazonaws.com/golangci-linters/v1/bin.tar.gz -O - | tar -C $GOBINPATH -xzvf -
COPY bin/bin.tar.gz .
RUN mkdir -p $GOBINPATH && tar -C $GOBINPATH -xzvf bin.tar.gz && rm bin.tar.gz

CMD ["/app/run.sh"]
