#!/bin/bash

export PATH=$PATH:/usr/lib/go-1.17/bin    

# args: swagger file
function generate_app() {

    if [ -z "$1" ]; then
        echo "version not provided"
        return 126
    fi

    VERSION=`echo $1 | sed -n "s/^.*_\(.*\).yaml$/\1/p"`

    echo "using version $1"

    if [[ "$2" == "custom" ]]; then
        swagger generate server -C templates/server_custom.yaml --target=/tmp/app/$1 -f /tmp/app/$1/swagger.yaml
    else
        swagger generate server -C templates/server.yaml --target=/tmp/app/$1 -f /tmp/app/$1/swagger.yaml
    fi

    cd /tmp/app/$1 && \
        go get github.com/go-openapi/runtime && \
        go get github.com/jessevdk/go-flags && \
        go get github.com/direktiv/apps/go/pkg/apps && \
        go mod tidy

}

# args: name
function init_app() {

    if [ -z "$1" ]; then
        echo "name not provided for application"
        return 126
    fi

    mkdir -p /tmp/app/v1.0.0
    sed "s/APPNAME/$1/g" templates/Dockerfile > /tmp/app/v1.0.0/Dockerfile

    sed "s/APPNAME/$1/g" templates/run.sh > /tmp/app/v1.0.0/run.sh
    chmod 755 /tmp/app/v1.0.0/run.sh

    sed "s/APPNAME/$1/g" templates/swagger.yaml > /tmp/app/v1.0.0/swagger.yaml
    cd  /tmp/app/v1.0.0/ && go mod init $1

}


echo "runing builder with args $@"

if [[ "$1" == "init" ]]; then
    init_app $2
elif [[ "$1" == "gen-custom" ]]; then
    generate_app $2 custom
elif [[ "$1" == "gen" ]]; then
    generate_app $2 
else
    echo "unknown builder command"
fi

