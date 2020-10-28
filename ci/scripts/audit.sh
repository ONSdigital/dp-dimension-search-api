#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-dimension-search-api
  make audit
popd  