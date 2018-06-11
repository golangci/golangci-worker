#!/bin/bash
set -e

echo "GOPATH='$GOPATH'"

if [[ ! -n $(find . -name "*.go") ]]; then
    echo 'no go files in repository'
    exit 0
fi

if [ -d "vendor" ]; then
  if [[ $(find vendor -name "*.go" | head -1) ]]; then
    echo "vendor dir exists with go sources, skip vendoring"
    exit 0
  fi
fi

if [[ -f 'Gopkg.toml' ]]; then
  	echo 'Dep was detected'
    dep ensure
elif [[ -f 'glide.yaml' ]]; then
  	echo 'Glide was detected'
  	glide install
elif [[ -f 'vendor/vendor.json' ]]; then
    echo 'Govendor was detected'
    govendor sync
elif [[ -f 'Godeps/Godeps.json' ]]; then
    echo 'Godep was detected'
    godep restore
else
    echo 'Vendoring wasnt found: use go get'
    go get -t ./...
fi
