#!/bin/sh
set -xe

ROOT_DIR=$PWD

go mod vendor
go vet ./...
go test .
for d in */ ; do
  if [ $d = "huobi/" ] || [ $d = "vendor/" ]; then
    continue
  fi
	cd $d
  if [ -f "$d/go.mod" ]; then
    go mod download
  fi
  go test -cover .
	cd $ROOT_DIR
done

