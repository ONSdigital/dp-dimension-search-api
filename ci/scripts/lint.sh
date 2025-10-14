#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-dimension-search-api
  make lint
popd