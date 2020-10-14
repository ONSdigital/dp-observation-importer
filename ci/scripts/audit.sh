#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-observation-importer
  make audit
popd