#! /bin/bash

if [ -n "$1" ]; then 
    if [ "$1" == "-p" ] || [ "$1" == "--proxy" ]; then
        # Using the most trusted Go module proxy in China
        docker run --rm -e GO111MODULE=on -e GOPROXY=https://goproxy.cn -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.14.2-buster go build -v -o output/kk
    else
        echo "The option should be '-p' or '--proxy'"
    fi
else
    docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.14.2-buster go build -v -o output/kk
fi
