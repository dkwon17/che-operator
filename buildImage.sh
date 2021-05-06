#!/bin/bash

docker build -t quay.io/aandriienko/che-operator:nightly .

docker push quay.io/aandriienko/che-operator:nightly