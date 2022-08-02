#!/bin/sh

docker build -t bash . && docker run -p 9191:8080 bash