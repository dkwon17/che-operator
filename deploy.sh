#!/bin/bash
#
# Copyright (c) 2012-2018 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation

set -e
set -x

BASE_DIR=$(cd "$(dirname "$0")"; pwd)

NAMESPACE="eclipse-che"
CHE_OPERATOR_IMAGE="quay.io/eclipse/che-operator:nightly"

while [[ "$#" -gt 0 ]]; do
  case $1 in
    '--namespace'|'-n') NAMESPACE=$2; shift 1;;
    '--che-operator-image') CHE_OPERATOR_IMAGE=$2; shift 1;;
    esac
    shift 1
done

oc apply -f ${BASE_DIR}/deploy/service_account.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/role.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/role_binding.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/cluster_role.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/cluster_role_binding.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/proxy_cluster_role.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/proxy_cluster_role_binding.yaml -n $NAMESPACE

oc apply -f ${BASE_DIR}/deploy/crds/org_v1_che_crd.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/dwco/chemanagers.che.eclipse.org.CustomResourceDefinition.yaml -n $NAMESPACE
oc apply -f ${BASE_DIR}/deploy/dwco/devworkspaceroutings.controller.devfile.io.CustomResourceDefinition.yaml -n $NAMESPACE

# sometimes the operator cannot get CRD right away
sleep 5

# patch and apply operator.yaml
cp ${BASE_DIR}/deploy/operator.yaml /tmp/operator.yaml
yq -riyY "( .spec.template.spec.containers[] | select(.name == \"che-operator\") | .image ) = \"${CHE_OPERATOR_IMAGE}\"" /tmp/operator.yaml
oc apply -f /tmp/operator.yaml -n $NAMESPACE

# create CR
oc apply -f ${BASE_DIR}/deploy/crds/org_v1_che_cr.yaml -n $NAMESPACE
