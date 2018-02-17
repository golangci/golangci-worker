#!/bin/bash

set -x
set -e

BIN_DIR=bin
rm -rf $BIN_DIR
docker build -t bins_builder -f app/docker/deploy.dockerfile .
docker run --rm \
  -v $(pwd)/deploy/docker_make_bins.sh:/app/run.sh \
  -v $(pwd)/$BIN_DIR:/app/go/bin \
  bins_builder

cd $BIN_DIR
ls -l
RES_TAR=bin.tar.gz
rm -f $RES_TAR
tar -zcvf $RES_TAR *

BINS_VERSION=1
aws s3 cp gometalinter.tar.gz s3://golangci-linters/v${BINS_VERSION}/$RES_TAR
# https://s3-us-west-2.amazonaws.com/golangci-linters/v1/bin.tar.gz

cd ..
