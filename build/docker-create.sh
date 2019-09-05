#!/bin/bash

set -e
set -o pipefail

if [ -z "$VERSION" ]
  then
    echo "No VERSION supplied."
    exit 1
fi

ARTIFACTS_DIR=artifacts
BINARY_NAME=mosoly-ledger-bridge

DOCKER_IMAGE_FULLNAME=$BINARY_NAME:$VERSION

echo Cleaning Docker context directory...
DOCKER_CONTEXT=$PWD/$ARTIFACTS_DIR/docker-context
mkdir -p $DOCKER_CONTEXT
rm -rf $DOCKER_CONTEXT/*

function cleanUp() {
  set +e

  echo Removing local Docker image...
  docker rmi $DOCKER_IMAGE_FULLNAME

  echo Cleaning Docker context directory...
  rm -rf $DOCKER_CONTEXT

  set -e
}

trap cleanUp EXIT

echo Copying Dockerfile
cp ./Dockerfile $DOCKER_CONTEXT/Dockerfile

echo Copying application files...
mkdir -p $DOCKER_CONTEXT/usr/bin $DOCKER_CONTEXT/etc/ssl/certs
cp ./$ARTIFACTS_DIR/$BINARY_NAME $DOCKER_CONTEXT/usr/bin/$BINARY_NAME
cp ./$ARTIFACTS_DIR/chamber $DOCKER_CONTEXT/usr/bin/
cp ./$ARTIFACTS_DIR/ca-certificates.crt $DOCKER_CONTEXT/etc/ssl/certs/

echo Building image...
docker build --force-rm=true -t $DOCKER_IMAGE_FULLNAME $DOCKER_CONTEXT

echo Exporting docker image full name: $DOCKER_IMAGE_FULLNAME
echo export DOCKER_IMAGE_FULLNAME=$(echo $DOCKER_IMAGE_FULLNAME) > ./$ARTIFACTS_DIR/$DOCKER_IMAGE_FULLNAME

docker save $DOCKER_IMAGE_FULLNAME | gzip > "./${ARTIFACTS_DIR}/${BINARY_NAME}_${VERSION}.tar.gz"

# We succeeded, reset trap and clean up normally.
trap - EXIT
cleanUp
exit 0
