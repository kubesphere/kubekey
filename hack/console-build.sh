#! /bin/bash

set -e

current_dir=$(cd $(dirname $0) && pwd)
cd ${current_dir}/../console
yarn install
yarn build
cd -
rm -rf ${current_dir}/../cmd/kk/cmd/console/router/templates/*
cp -r ${current_dir}/../console/build/* ${current_dir}/../cmd/kk/cmd/console/router/templates/
