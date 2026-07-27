#!/bin/bash

nc -zv -w 2 localhost {{ if .image_registry.auth.plain_http | default false }}{{ .image_registry.http_port | default 80 }}{{ else }}{{ .image_registry.https_port | default 443 }}{{ end }} > /dev/null 2>&1

if [ $? -eq 0 ]; then
    exit 0
else
    exit 1
fi
