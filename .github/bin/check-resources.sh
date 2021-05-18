#!/bin/bash
#
# Copyright (c) 2019-2021 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation

# Checks if repository resources are up to date:
# - CRDs
# - nightly olm bundle

set -e

ROOT_PROJECT_DIR="${GITHUB_WORKSPACE}"
if [ -z "${ROOT_PROJECT_DIR}" ]; then
  SCRIPT=$(readlink -f "${BASH_SOURCE[0]}")
  ROOT_PROJECT_DIR=$(dirname $(dirname $(dirname ${SCRIPT})))
fi

installOperatorSDK() {
  OPERATOR_SDK_BINARY=$(command -v operator-sdk) || true
  if [[ ! -x "${OPERATOR_SDK_BINARY}" ]]; then
    OPERATOR_SDK_TEMP_DIR="$(mktemp -q -d -t "OPERATOR_SDK_XXXXXX" 2>/dev/null || mktemp -q -d)"
    pushd "${OPERATOR_SDK_TEMP_DIR}" || exit
    echo "[INFO] Downloading 'operator-sdk' cli tool..."

    ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
    OS=$(uname | awk '{print tolower($0)}')

    OPERATOR_SDK_VERSION=$(yq -r ".\"operator-sdk\"" "${ROOT_PROJECT_DIR}/REQUIREMENTS")
    OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}
    curl -sLo operator-sdk ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}

    export OPERATOR_SDK_BINARY="${OPERATOR_SDK_TEMP_DIR}/operator-sdk"
    chmod +x "${OPERATOR_SDK_BINARY}"
    echo "[INFO] Downloading completed!"
    echo "[INFO] $(${OPERATOR_SDK_BINARY} version)"
    popd || exit
  fi
}

updateResources() {
  export NO_DATE_UPDATE="true"
  export NO_INCREMENT="true"

  source "${ROOT_PROJECT_DIR}/olm/update-resources.sh"
}

# check_che_types function check first if pkg/apis/org/v1/che_types.go file suffer modifications and
# in case of modification should exist also modifications in deploy/crds/* folder.
checkCRDs() {
    echo "[INFO] Checking CRDs"

    # files to check
    local CRD_V1="config/crd/bases/org_v1_che_crd.yaml"
    local CRD_V1BETA1="config/crd/bases/org_v1_che_crd.yaml"

    source "${ROOT_PROJECT_DIR}/olm/update-resources.sh"

    changedFiles=($(cd ${ROOT_PROJECT_DIR}; git diff --name-only))

    # Check if there are any difference in the crds. If yes, then fail check.
    if [[ " ${changedFiles[*]} " =~ $CRD_V1 ]] || [[ " ${changedFiles[*]} " =~ $CRD_V1BETA1 ]]; then
        echo "[ERROR] CRD file is not up to date: ${BASH_REMATCH}"
        echo "[ERROR] Run 'olm/update-resources.sh' to regenerate CRD files."
        exit 1
    else
        echo "[INFO] CRDs files are up to date."
    fi
}

checkNightlyOlmBundle() {
  # files to check
  local CSV_FILE_KUBERNETES="${ROOT_PROJECT_DIR}/bundle/nightly/eclipse-che-preview-kubernetes/manifests/che-operator.clusterserviceversion.yaml"
  local CSV_FILE_OPENSHIFT="${ROOT_PROJECT_DIR}/bundle/nightly/eclipse-che-preview-openshift/manifests/che-operator.clusterserviceversion.yaml"
  local CRD_FILE_KUBERNETES="${ROOT_PROJECT_DIR}/bundle/nightly/eclipse-che-preview-kubernetes/manifests/org.eclipse.che_checlusters.yaml"
  local CRD_FILE_OPENSHIFT="${ROOT_PROJECT_DIR}/bundle/nightly/eclipse-che-preview-openshift/manifests/org.eclipse.che_checlusters.yaml"

  changedFiles=($(git diff --name-only))
  if [[ " ${changedFiles[*]} " =~ $CSV_FILE_OPENSHIFT ]] || [[ " ${changedFiles[*]} " =~ $CSV_FILE_OPENSHIFT ]] || \
     [[ " ${changedFiles[*]} " =~ $CRD_FILE_KUBERNETES ]] || [[ " ${changedFiles[*]} " =~ $CRD_FILE_OPENSHIFT ]]; then
    echo "[ERROR] Nighlty bundle is not up to date: ${BASH_REMATCH}"
    echo "[ERROR] Run 'olm/update-resources.sh' to regenerate CSV/CRD files."
    exit 1
  else
    echo "[INFO] Nightly bundles are up to date."
  fi
}

checkDockerfile() {
  # files to check
  local Dockerfile="${ROOT_PROJECT_DIR}/Dockerfile"

  changedFiles=($(cd ${ROOT_PROJECT_DIR}; git diff --name-only))
  if [[ " ${changedFiles[*]} " =~ $Dockerfile ]]; then
    echo "[ERROR] Dockerfile is not up to date"
    echo "[ERROR] Run 'olm/update-resources.sh' to update Dockerfile"
    exit 1
  else
    echo "[INFO] Dockerfile is up to date."
  fi
}

checkOperatorYaml() {
  # files to check
  local OperatorYaml="${ROOT_PROJECT_DIR}/config/manager/manager.yaml"

  changedFiles=($(cd ${ROOT_PROJECT_DIR}; git diff --name-only))
  if [[ " ${changedFiles[*]} " =~ $OperatorYaml ]]; then
    echo "[ERROR] $OperatorYaml is not up to date"
    echo "[ERROR] Run 'olm/update-resources.sh' to update $OperatorYaml"
    exit 1
  else
    echo "[INFO] $OperatorYaml is up to date."
  fi
}

installOperatorSDK
updateResources
checkCRDs
checkNightlyOlmBundle
checkDockerfile
checkOperatorYaml

echo "[INFO] Done."
