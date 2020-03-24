#!/bin/bash -eux


pushd $cwd/dp-search-api
  make test
popd
