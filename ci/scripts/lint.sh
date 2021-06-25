#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-observation-importer
  make lint
popd