#!/bin/sh

docker build -t crypto . && docker run -p 9191:8080 crypto