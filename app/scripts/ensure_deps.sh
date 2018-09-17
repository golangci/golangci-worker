#!/bin/bash
set -e

go clean -cache
rm -rf /tmp/glide-vendor*
rm -rf /tmp/go-build*
rm -rf $HOME/.glide/cache

function install_by_vendoring_tool() {
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
}

function has_vendor_dir() {
    if [ -d "vendor" ]; then
        if [[ $(find vendor -name "*.go" | head -1) ]]; then
            echo "vendor dir exists with go sources, skip vendoring"
            return 0
        fi
    fi

    return 1
}

echo "GOPATH='$GOPATH'"

if [[ ! -n $(find . -name "*.go") ]]; then
    echo 'no go files in repository'
    exit 0
fi

if ! has_vendor_dir; then
    install_by_vendoring_tool
fi

if [[ "$(golangci-lint run --no-config --disable-all -E typecheck | fgrep 'could not import')" ]]; then
	echo "golangci-lint run --no-config --disable-all -E typecheck | fgrep 'could not import' found something, run 'go get -t ./...'"
	go get -t ./...
fi