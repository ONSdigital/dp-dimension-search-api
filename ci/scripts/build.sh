#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-dimension-search-api
  make build && mv build/dp-dimension-search-api $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
