#!/bin/bash

export PATH=$PATH:/usr/lib/go-1.17/bin    

function generate_app() {

    if [[ "$2" == "custom" ]]; then
        swagger generate server -C templates/server_custom.yaml --target=/tmp/app -f /tmp/app/swagger.yaml
    else
        swagger generate server -C templates/server.yaml --target=/tmp/app -f /tmp/app/swagger.yaml
    fi

    cd /tmp/app/$1 && go mod tidy && \
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

    mkdir -p /tmp/app
    
    sed "s/APPNAME/$1/g" templates/Dockerfile > /tmp/app/Dockerfile

    sed "s/APPNAME/$1/g" templates/run.sh > /tmp/app/run.sh
    chmod 755 /tmp/app/run.sh

    sed "s/APPNAME/$1/g" templates/swagger.yaml > /tmp/app/swagger.yaml
    cd  /tmp/app/ && go mod init $1

}

generate_docs() {
    swagger generate markdown -f /tmp/app/swagger.yaml --output=/tmp/app/readme.md -t /tmp/app/ --template-dir=templates/ --with-flatten=full
    mv /tmp/app/readme.md /tmp/app/README.md
}

echo "runing builder with args $@"

if [[ "$1" == "init" ]]; then
    init_app $2
elif [[ "$1" == "gen-custom" ]]; then
    generate_app $2 custom
elif [[ "$1" == "gen" ]]; then
    generate_app $2 
elif [[ "$1" == "docs" ]]; then
    generate_docs
else
    echo "unknown builder command"
fi

