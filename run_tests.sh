#!/bin/sh
set -xe

ROOT_DIR=$PWD

go mod vendor
go test .
for d in */ ; do
  if [ $d = "huobi/" ] || [ $d = "vendor/" ]; then
    continue
  fi
	cd $d
    if [ -f "$d/go.mod" ]; then

	go mod download

    fi
  go test .
	cd $ROOT_DIR
done

# ignore huobi for now. as the test calls directly to the api and some of them are outdated
