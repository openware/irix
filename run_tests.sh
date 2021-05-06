#!/bin/sh
set -xe

ROOT_DIR=$PWD

for d in */ ; do
    if [ -f "$d/go.mod" ]; then
	cd $d

	go mod download

	cd $ROOT_DIR
    fi
done

# ignore huobi for now. as the test calls directly to the api and some of them are outdated
go mod vendor
go test $(go list ./... | grep -v huobi)
