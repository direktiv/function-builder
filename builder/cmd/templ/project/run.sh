#!/bin/sh

docker build -t APPNAME . && docker run -p 9191:8080 APPNAME