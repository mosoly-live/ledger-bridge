#!/bin/bash


set -e -o pipefail

if [ -z "$CI_PIPELINE_ID" ]
  then
    echo "No CI_PIPELINE_ID supplied."
    exit 1
fi

if [ -z "$REPO_NAME" ]
  then
    echo "No REPO_NAME supplied."
    exit 1
fi

GOLANG_VERSION=1.12
PACKAGE_NAME=$REPO_NAME
PACKAGE_FULL_PATH=/go/src/$PACKAGE_NAME
VERSION=$(date +"%Y.%m.%d").$CI_PIPELINE_ID

docker run -i --rm \
-v "$PWD":$PACKAGE_FULL_PATH \
-w $PACKAGE_FULL_PATH \
golang:$GOLANG_VERSION /bin/bash << COMMANDS
set -e -o pipefail
HOME=$PACKAGE_FULL_PATH
cd $PACKAGE_FULL_PATH
echo -e "machine gitlab.com\nlogin gitlab-ci-token\npassword ${CI_JOB_TOKEN}" > ~/.netrc
make -j4 VERSION="$VERSION"
curl -sS -L -o ./artifacts/chamber https://github.com/segmentio/chamber/releases/download/v1.13.0/chamber-v1.13.0-linux-amd64
curl -sS -L -o ./artifacts/ca-certificates.crt https://raw.githubusercontent.com/bagder/ca-bundle/master/ca-bundle.crt
chmod +x ./artifacts/chamber
COMMANDS

echo Done.
