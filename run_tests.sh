#!/bin/sh
set -xe

ROOT_DIR=$PWD

go ./test... -cover -race
go vet ./test...

for d in */ ; do
    if [ -f "$d/go.mod" ]; then
	cd $d

	go mod download
	go test ./... -cover -race
	go vet ./...

	cd $ROOT_DIR
    fi
done
