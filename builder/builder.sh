#!/bin/bash

set -e

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

    mkdir -p /tmp/app/test/data
    mkdir -p /tmp/app/test/secrets/data
    echo -n "My Name" > /tmp/app/test/data/example.dat
    cp templates/testing/test.feature /tmp/app/test/
    cp templates/testing/karate-config.js /tmp/app/test/
    cp templates/testing/log-config.xml /tmp/app/test/

    # setup project
    cp templates/project/gitignore /tmp/app/.gitignore
    cp templates/project/LICENSE /tmp/app/LICENSE

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

test() {
    # upload data directory
    old=`pwd`
    if find /tmp/app/test/data -mindepth 1 -maxdepth 1 | read; then
        rm -Rf /tmp/app/test/data.tar.gz
        cd /tmp/app/test/data && tar -cvzf ../data.tar.gz *
        curl -XPOST --data-binary @../data.tar.gz http://192.168.0.177:9292
    fi

    if find /tmp/app/test/secrets/data -mindepth 1 -maxdepth 1 | read; then
        rm -Rf /tmp/app/test/secrets-data.tar
        cd /tmp/app/test/secrets/data && tar -cvzf ../../secrets-data.tar.gz *
        curl -XPOST --data-binary @../../secrets-data.tar.gz http://192.168.0.177:9292
    fi

    if [ -d "/tmp/app/test/secrets" ]; then
        cd /tmp/app/test/secrets   
        rm -Rf ../secrets-data.tar
        tar -cvzf ../secrets-data-all.tar.gz *
    fi 
    cd $old

    if [[ -z "${REPORT_DIR}" ]]; then
        mkdir -p /tmp/app/test/reports
        REPORT_DIR="/tmp/app/test/reports"
    else
        REPORT_DIR="${REPORT_DIR}"
    fi

    java ${@} -Dlogback.configurationFile=/tmp/app/test/log-config.xml -jar karate-1.2.0.jar --format=~html,cucumber:json -C --output=${REPORT_DIR} --configdir /tmp/app/test/  /tmp/app/test/test.feature
}

echo "running builder with args $@"

if [[ "$1" == "init" ]]; then
    init_app $2
elif [[ "$1" == "gen-custom" ]]; then
    generate_app $2 custom
elif [[ "$1" == "gen" ]]; then
    generate_app $2 
elif [[ "$1" == "docs" ]]; then
    generate_docs
elif [[ "$1" == "test" ]]; then
    test ${@:2}
else
    echo "unknown builder command"
fi

