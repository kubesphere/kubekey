#! /bin/bash

#docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.14.2-buster go build -v -o kk

# Using the most trusted Go module proxy in China mainland
docker run --rm -e GO111MODULE=on -e GOPROXY=https://goproxy.cn -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.14.2-buster go build -v -o kk

