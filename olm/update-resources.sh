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

# Generated CRDs based on pkg/apis/org/v1/che_types.go:
# - deploy/crds/org_v1_che_crd.yaml
# - deploy/crds/org_v1_che_crd-v1beta1.yaml

set -e

unset UBI8_MINIMAL_IMAGE
unset PLUGIN_BROKER_METADATA_IMAGE
unset PLUGIN_BROKER_ARTIFACTS_IMAGE
unset JWT_PROXY_IMAGE

SCRIPT=$(readlink -f "${BASH_SOURCE[0]}")
ROOT_PROJECT_DIR=$(dirname $(dirname ${SCRIPT}))

checkOperatorSDKVersion() {
  if [ -z "${OPERATOR_SDK_BINARY}" ]; then
    OPERATOR_SDK_BINARY=$(command -v operator-sdk)
    if [[ ! -x "${OPERATOR_SDK_BINARY}" ]]; then
      echo "[ERROR] operator-sdk is not installed."
      exit 1
    fi
  fi

  local operatorVersion=$("${OPERATOR_SDK_BINARY}" version)
  REQUIRED_OPERATOR_SDK=$(yq -r ".\"operator-sdk\"" "${ROOT_PROJECT_DIR}/REQUIREMENTS")
  [[ $operatorVersion =~ .*${REQUIRED_OPERATOR_SDK}.* ]] || { echo "operator-sdk ${REQUIRED_OPERATOR_SDK} is required"; exit 1; }

  if [ -z "${GOROOT}" ]; then
    echo "[ERROR] set up '\$GOROOT' env variable to make operator-sdk working"
    exit 1
  fi
}

# Removes `required` attributes for fields to be compatible with OCP 3.11
removeRequiredAttribute() {
  REQUIRED=false
  while IFS= read -r line
  do
      if [[ $REQUIRED == true ]]; then
          if [[ $line == *"- "* ]]; then
              continue
          else
              REQUIRED=false
          fi
      fi

      if [[ $line == *"required:"* ]]; then
          REQUIRED=true
          continue
      fi

      echo  "$line" >> $1.tmp
  done < "$1"
  mv $1.tmp $1
}

detectImages() {
  ubiMinimal8Version=$(skopeo inspect docker://registry.access.redhat.com/ubi8-minimal:latest | jq -r '.Labels.version')
  ubiMinimal8Release=$(skopeo inspect docker://registry.access.redhat.com/ubi8-minimal:latest | jq -r '.Labels.release')
  UBI8_MINIMAL_IMAGE="registry.access.redhat.com/ubi8-minimal:"$ubiMinimal8Version"-"$ubiMinimal8Release
  skopeo inspect docker://$UBI8_MINIMAL_IMAGE > /dev/null

  wget https://raw.githubusercontent.com/eclipse/che/master/assembly/assembly-wsmaster-war/src/main/webapp/WEB-INF/classes/che/che.properties -q -O /tmp/che.properties
  PLUGIN_BROKER_METADATA_IMAGE=$(cat /tmp/che.properties| grep "che.workspace.plugin_broker.metadata.image" | cut -d = -f2)
  PLUGIN_BROKER_ARTIFACTS_IMAGE=$(cat /tmp/che.properties | grep "che.workspace.plugin_broker.artifacts.image" | cut -d = -f2)
  JWT_PROXY_IMAGE=$(cat /tmp/che.properties | grep "che.server.secure_exposer.jwtproxy.image" | cut -d = -f2)

  echo "[INFO] UBI base image               : $UBI8_MINIMAL_IMAGE"
  echo "[INFO] Plugin broker metadata image : $PLUGIN_BROKER_METADATA_IMAGE"
  echo "[INFO] Plugin broker artifacts image: $PLUGIN_BROKER_ARTIFACTS_IMAGE"
  echo "[INFO] Plugin broker jwt proxy image: $JWT_PROXY_IMAGE"
}

updateOperatorYaml() {
  # todo fight for old yaml name...
  OPERATOR_YAML="${ROOT_PROJECT_DIR}/config/manager/manager.yaml"
  yq -riY "( .spec.template.spec.containers[] | select(.name == \"che-operator\").env[] | select(.name == \"RELATED_IMAGE_pvc_jobs\") | .value ) = \"${UBI8_MINIMAL_IMAGE}\"" ${OPERATOR_YAML}
  yq -riY "( .spec.template.spec.containers[] | select(.name == \"che-operator\").env[] | select(.name == \"RELATED_IMAGE_che_workspace_plugin_broker_metadata\") | .value ) = \"${PLUGIN_BROKER_METADATA_IMAGE}\"" ${OPERATOR_YAML}
  yq -riY "( .spec.template.spec.containers[] | select(.name == \"che-operator\").env[] | select(.name == \"RELATED_IMAGE_che_workspace_plugin_broker_artifacts\") | .value ) = \"${PLUGIN_BROKER_ARTIFACTS_IMAGE}\"" ${OPERATOR_YAML}
  yq -riY "( .spec.template.spec.containers[] | select(.name == \"che-operator\").env[] | select(.name == \"RELATED_IMAGE_che_server_secure_exposer_jwt_proxy_image\") | .value ) = \"${JWT_PROXY_IMAGE}\"" ${OPERATOR_YAML}
  addLicenseHeader $OPERATOR_YAML
}

updateDockerfile() {
  DOCKERFILE="${ROOT_PROJECT_DIR}/Dockerfile"
  sed -i 's|registry.access.redhat.com/ubi8-minimal:.*|'${UBI8_MINIMAL_IMAGE}'|g' $DOCKERFILE
}

updateNighltyBundle() {
  source ${ROOT_PROJECT_DIR}/olm/olm.sh

  for platform in 'kubernetes' 'openshift'
  do
    if [ -z "${NO_INCREMENT}" ]; then
      source "${ROOT_PROJECT_DIR}/olm/incrementNightlyBundles.sh"
      incrementNightlyVersion "${platform}"
    fi

    echo "[INFO] Updating OperatorHub bundle for platform '${platform}'"

    pushd "${ROOT_PROJECT_DIR}" || true

    NIGHTLY_BUNDLE_PATH=$(getBundlePath "${platform}" "nightly")
    bundleCSVName="che-operator.clusterserviceversion.yaml"
    NEW_CSV=${NIGHTLY_BUNDLE_PATH}/manifests/${bundleCSVName}
    newNightlyBundleVersion=$(yq -r ".spec.version" "${NEW_CSV}")
    echo "[INFO] Creation new nightly bundle version: ${newNightlyBundleVersion}"

    pushd "${ROOT_PROJECT_DIR}" || true
    make bundle "platform=${platform}" "VERSION=${newNightlyBundleVersion}"
    popd || true

    containerImage=$(sed -n 's|^ *image: *\([^ ]*/che-operator:[^ ]*\) *|\1|p' ${NEW_CSV})
    echo "[INFO] Updating new package version fields:"
    echo "[INFO]        - containerImage => ${containerImage}"
    sed -e "s|containerImage:.*$|containerImage: ${containerImage}|" "${NEW_CSV}" > "${NEW_CSV}.new"
    mv "${NEW_CSV}.new" "${NEW_CSV}"

    if [ -z "${NO_DATE_UPDATE}" ]; then
      createdAt=$(date -u +%FT%TZ)
      echo "[INFO]        - createdAt => ${createdAt}"
      sed -e "s/createdAt:.*$/createdAt: \"${createdAt}\"/" "${NEW_CSV}" > "${NEW_CSV}.new"
      mv "${NEW_CSV}.new" "${NEW_CSV}"
    fi

    if [ -z "${NO_INCREMENT}" ]; then
      incrementNightlyVersion "${platform}"
    fi

    addLicenseHeader "${ROOT_PROJECT_DIR}/config/crd/bases/org_v1_che_crd-v1beta1.yaml"
    addLicenseHeader "${ROOT_PROJECT_DIR}/config/crd/bases/org_v1_che_crd.yaml"

    # templateCRD="${ROOT_PROJECT_DIR}/config/crd/bases/org_v1_che_crd.yaml"
    # platformCRD="${NIGHTLY_BUNDLE_PATH}/manifests/org.eclipse.che_checlusters.yaml"

    # cp -rf $templateCRD $platformCRD
    # if [[ $platform == "openshift" ]]; then
    #   yq -riSY  '.spec.preserveUnknownFields = false' $platformCRD
    #   eval head -10 $templateCRD | cat - ${platformCRD} > tmp.crd && mv tmp.crd ${platformCRD}
    # fi

    echo "Done for ${platform}"

    if [[ -n "$TAG" ]]; then
      echo "[INFO] Set tags in nightly OLM files"
      sed -ri "s/(.*:\s?)${RELEASE}([^-])?$/\1${TAG}\2/" "${NEW_CSV}"
    fi

    YAML_CONTENT=$(cat "${NEW_CSV}")
    if [[ $platform == "kubernetes" ]]; then
      clusterPermLength=$(echo "${YAML_CONTENT}" | yq -r ".spec.install.spec.clusterPermissions[0].rules | length")
      for (( i=0; i < ${clusterPermLength}; i++ )); do
        apiGroupLength=$(echo "${YAML_CONTENT}" | yq -r '.spec.install.spec.clusterPermissions[0].rules['${i}'].apiGroups | length')
        if [ "${apiGroupLength}" -gt 0 ]; then
          for (( j=0; j < ${apiGroupLength}; j++ )); do
            if [[ $(echo "${YAML_CONTENT}" | yq -r '.spec.install.spec.clusterPermissions[0].rules['${i}'].apiGroups['${j}']') =~ openshift.io$ ]]; then
              YAML_CONTENT=$(echo "${YAML_CONTENT}" | yq -rY 'del(.spec.install.spec.clusterPermissions[0].rules['${i}'])' )
              j=${j}-1
              i=${i}-1
              break
            fi
          done
        fi
      done
    
      clusterPermLength=$(echo "${YAML_CONTENT}" | yq -r ".spec.install.spec.permissions[0].rules | length")
      for (( i=0; i < ${clusterPermLength}; i++ )); do
        apiGroupLength=$(echo "${YAML_CONTENT}" | yq -r '.spec.install.spec.permissions[0].rules['${i}'].apiGroups | length')
        if [ "${apiGroupLength}" -gt 0 ]; then
          for (( j=0; j < ${apiGroupLength}; j++ )); do
            if [[ $(echo "${YAML_CONTENT}" | yq -r '.spec.install.spec.permissions[0].rules['${i}'].apiGroups['${j}']') =~ openshift.io$ ]]; then
              YAML_CONTENT=$(echo "${YAML_CONTENT}" | yq -rY 'del(.spec.install.spec.permissions[0].rules['${i}'])' )
              j=${j}-1
              i=${i}-1
              break
            fi
          done
        fi
      done
    fi
    echo "${YAML_CONTENT}" > "${NEW_CSV}"

    if [[ $platform == "openshift" ]]; then
      # Removes che-tls-secret-creator
      index=0
      while [[ $index -le 30 ]]
      do
        if [[ $(cat ${NEW_CSV} | yq -r '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env['$index'].name') == "RELATED_IMAGE_che_tls_secrets_creation_job" ]]; then
          yq -rYSi 'del(.spec.install.spec.deployments[0].spec.template.spec.containers[0].env['$index'])' ${NEW_CSV}
          break
        fi
        index=$((index+1))
      done
    fi

    # Fix sample
    if [ "${platform}" == "openshift" ]; then
      echo "[INFO] Fix openshift sample"
      sample=$(yq -r ".metadata.annotations.\"alm-examples\"" "${NEW_CSV}")
      fixedSample=$(echo "${sample}" | yq -r ".[0] | del(.spec.k8s) | [.]" | sed -r 's/"/\\"/g')
      # Update sample in the CSV
      yq -rY " (.metadata.annotations.\"alm-examples\") = \"${fixedSample}\"" "${NEW_CSV}" > "${NEW_CSV}.old"
      mv "${NEW_CSV}.old" "${NEW_CSV}"
    fi
    if [ "${platform}" == "kubernetes" ]; then
      echo "[INFO] Fix kubernetes sample"
      sample=$(yq -r ".metadata.annotations.\"alm-examples\"" "${NEW_CSV}")
      fixedSample=$(echo "${sample}" | yq -r ".[0] | (.spec.k8s.ingressDomain) = \"\" | del(.spec.auth.openShiftoAuth) | [.]" | sed -r 's/"/\\"/g')
      # Update sample in the CSV
      yq -rY " (.metadata.annotations.\"alm-examples\") = \"${fixedSample}\"" "${NEW_CSV}" > "${NEW_CSV}.old"
      mv "${NEW_CSV}.old" "${NEW_CSV}"
    fi

    # set `app.kubernetes.io/managed-by` label
    yq -riSY  '(.spec.install.spec.deployments[0].spec.template.metadata.labels."app.kubernetes.io/managed-by") = "olm"' "${NEW_CSV}"

    # Format code.
    yq -rY "." "${NEW_CSV}" > "${NEW_CSV}.old"
    mv "${NEW_CSV}.old" "${NEW_CSV}"

    popd || true
  done
}

addLicenseHeader() {
echo -e "#
#  Copyright (c) 2019-2021 Red Hat, Inc.
#    This program and the accompanying materials are made
#    available under the terms of the Eclipse Public License 2.0
#    which is available at https://www.eclipse.org/legal/epl-2.0/
#
#  SPDX-License-Identifier: EPL-2.0
#
#  Contributors:
#    Red Hat, Inc. - initial API and implementation
$(cat $1)" > $1
}

checkOperatorSDKVersion
detectImages
# update CRDs and autogenerated go files...
make generate; make manifests
updateOperatorYaml
updateDockerfile
updateNighltyBundle
