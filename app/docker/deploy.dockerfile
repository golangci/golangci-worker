FROM heroku/heroku:16

ENV GO_VERSION=1.9.3
ENV OS=linux
ENV ARCH=amd64

WORKDIR /app
RUN wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GO_VERSION}.$OS-$ARCH.tar.gz

ENV GOPATH=/app/go
ENV GOBINPATH=$GOPATH/bin
ENV PATH=$PATH:/usr/local/go/bin:$GOBINPATH
VOLUME $GOBINPATH

CMD ["/app/run.sh"]
