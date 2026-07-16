#!/bin/bash

nc -zv -w 2 localhost {{ if .image_registry.auth.plain_http | default false }}{{ .image_registry.harbor.http_port }}{{ else }}{{ .image_registry.harbor.https_port }}{{ end }} > /dev/null 2>&1

if [ $? -eq 0 ]; then
    exit 0
else
    exit 1
fi
