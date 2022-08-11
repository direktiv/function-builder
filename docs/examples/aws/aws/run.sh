#!/bin/sh

docker build -t aws . && docker run -p 9191:8080 aws