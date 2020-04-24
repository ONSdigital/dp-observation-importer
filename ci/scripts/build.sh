#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-observation-importer
  make build && cp build/dp-observation-importer $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
