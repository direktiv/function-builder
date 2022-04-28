#!/bin/sh

docker build -t aws-cli . && docker run -p 8080:8080 aws-cli