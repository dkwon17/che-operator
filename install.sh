#!/bin/bash

#  make bundle

make docker-build docker-push IMG="quay.io/aandriienko/che-operator:nightly"
make bundle IMG="quay.io/aandriienko/che-operator-bundle:v1.0.2"
make bundle-build bundle-push
operator-sdk run bundle quay.io/aandriienko/che-operator-bundle:v1.0.2

sleep 40

oc apply -f config/samples/org.eclipse.che_v1_checluster.yaml

#  operator-sdk cleanup che-operator
# docker build -t quay.io/aandriienko/che-operator:nightly .
# docker push quay.io/aandriienko/che-operator:nightly
