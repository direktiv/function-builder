#!/bin/sh

docker build -t image-magick . && docker run -p 8080:8080 image-magick