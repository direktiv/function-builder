#!/bin/sh

docker build -t bash-service . && docker run -p 8080:8080 bash-service