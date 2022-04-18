#!/bin/bash

export PATH=$PATH:/usr/lib/go-1.17/bin    

# args: swagger file
function generate_app() {

    if [ -z "$1" ]; then
        echo "swagger file not provided"
        return 126
    fi

    VERSION=`echo $1 | sed -n "s/^.*_\(.*\).yaml$/\1/p"`

    echo "using swagger file $1, version $VERSION"

    swagger generate server -C templates/server.yaml --target=/tmp/app -f /tmp/app/$1

    cd /tmp/app/ && go mod tidy && \
        go get github.com/go-openapi/runtime && \
        go get github.com/jessevdk/go-flags

}

# args: name
function init_app() {

    if [ -z "$1" ]; then
        echo "name not provided for application"
        return 126
    fi

    mkdir -p /tmp/app
    sed "s/APPNAME/$1/g" /tmp/swagger.yaml > /tmp/app/swagger_v1.0.0.yaml
    cd  /tmp/app/ && go mod init $1

}

echo $1

if [[ "$1" == "init" ]]; then
    init_app $2
elif [[ "$1" == "generate" ]]; then
    generate_app $2 
else
    echo "Strings are not equal."
fi

# echo ${@:2}

