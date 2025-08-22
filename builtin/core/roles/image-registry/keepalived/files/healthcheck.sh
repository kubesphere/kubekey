#!/bin/bash

nc -zv -w 2 localhost 443 > /dev/null 2>&1

if [ $? -eq 0 ]; then
    exit 0
else
    exit 1
fi
