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

    if [[ "$2" == "custom" ]]; then
        swagger generate server -C templates/server_custom.yaml --target=/tmp/app -f /tmp/app/$1
    else
        swagger generate server -C templates/server.yaml --target=/tmp/app -f /tmp/app/$1
    fi

    cd /tmp/app/ && go mod tidy && \
        go get github.com/go-openapi/runtime && \
        go get github.com/jessevdk/go-flags && \
        go get github.com/direktiv/apps/go/pkg/apps

}

# args: name
function init_app() {

    if [ -z "$1" ]; then
        echo "name not provided for application"
        return 126
    fi

    cp templates/Dockerfile /tmp/app

    mkdir -p /tmp/app
    sed "s/APPNAME/$1/g" /tmp/swagger.yaml > /tmp/app/swagger_v1.0.0.yaml
    cd  /tmp/app/ && go mod init $1

}

if [[ "$1" == "init" ]]; then
    init_app $2
elif [[ "$1" == "gen-custom" ]]; then
    generate_app $2 custom
elif [[ "$1" == "gen" ]]; then
    generate_app $2 
else
    echo "Strings are not equal."
fi

# echo ${@:2}

