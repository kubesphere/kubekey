#! /bin/bash

docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.14.2-buster go build -v -o kk
