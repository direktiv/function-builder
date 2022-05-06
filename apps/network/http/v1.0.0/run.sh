#!/bin/sh

docker build -t http-request . && docker run -p 8080:8080 http-request