#!/bin/bash
set -euxo pipefail

GOBINPATH=/app/go/bin

DEP_VERSION=v0.4.1

wget https://github.com/golang/dep/releases/download/${DEP_VERSION}/dep-linux-amd64 -O $GOBINPATH/dep
chmod a+x $GOBINPATH/dep

GODEP_VERSION=v80
wget https://github.com/tools/godep/releases/download/${GODEP_VERSION}/godep_linux_amd64 -O $GOBINPATH/godep
chmod a+x $GOBINPATH/godep

GOVENDOR_VERSION=v1.0.8
wget https://github.com/kardianos/govendor/releases/download/${GOVENDOR_VERSION}/govendor_linux_amd64 -O $GOBINPATH/govendor
chmod a+x $GOBINPATH/govendor

GLIDE_VERSION=v0.13.1
wget https://github.com/Masterminds/glide/releases/download/${GLIDE_VERSION}/glide-${GLIDE_VERSION}-linux-amd64.tar.gz -O - | \
  tar xzvf -
mv linux-amd64/glide $GOBINPATH/

GOMETALINTER_VERSION=v2.0.3
wget https://github.com/alecthomas/gometalinter/releases/download/v2.0.3/gometalinter-v2.0.3-linux-amd64.tar.bz2 -O - | \
  bunzip2 -c - | \
  tar xvf -
mv gometalinter-${GOMETALINTER_VERSION}-linux-amd64/gometalinter $GOBINPATH/
mv gometalinter-${GOMETALINTER_VERSION}-linux-amd64/linters/* $GOBINPATH/
