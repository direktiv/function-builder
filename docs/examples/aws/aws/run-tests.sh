#!/bin/bash

if [[ -z "${DIREKTIV_TEST_URL}" ]]; then
	echo "Test URL is not set, setting it to http://localhost:9191"
	DIREKTIV_TEST_URL="http://localhost:9191"
fi

if [[ -z "${DIREKTIV_SECRET_awsKey}" ]]; then
	echo "Secret awsKey is required, set it with DIREKTIV_SECRET_awsKey"
	exit 1
fi

if [[ -z "${DIREKTIV_SECRET_awsSecret}" ]]; then
	echo "Secret awsSecret is required, set it with DIREKTIV_SECRET_awsSecret"
	exit 1
fi

docker run --network=host -v `pwd`/tests/:/tests direktiv/karate java -DtestURL=${DIREKTIV_TEST_URL} -Dlogback.configurationFile=/logging.xml -DawsKey="${DIREKTIV_SECRET_awsKey}" -DawsSecret="${DIREKTIV_SECRET_awsSecret}"  -jar /karate.jar /tests/v1.0/karate.yaml.test.feature ${*:1}