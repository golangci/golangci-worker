#!/bin/bash
set -e
set -x

echo "GOPATH='$GOPATH'"

if [[ -f 'Gopkg.toml' ]]; then
  	echo 'Dep was detected'
  	dep ensure -v
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
