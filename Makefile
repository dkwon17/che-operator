# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 1.0.2

CHANNELS = "nightly"

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

DEFAULT_CHANNEL = "nightly"

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# ifeq ($(OPERATOR_SDK_BINARY),)
OPERATOR_SDK_BINARY ?= operator-sdk
# endif

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# quay.io/eclipse/che-operator-bundle:$VERSION and quay.io/eclipse/che-operator-catalog:$VERSION.
IMAGE_TAG_BASE ?= quay.io/aandriienko/che-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= quay.io/aandriienko/che-operator:nightly
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
CRD_BETA_OPTIONS ?= "$(CRD_OPTIONS),crdVersions=v1beta1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
# SHELL = /usr/bin/env bash
# .SHELLFLAGS = -ec
.ONESHELL:

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

download-operator-sdk:
	ARCH=$$(case "$$(uname -m)" in
	x86_64) echo -n amd64 ;;
	aarch64) echo -n arm64 ;;
	*) echo -n $$(uname -m)
	esac)
	OS=$$(uname | awk '{print tolower($$0)}')

	OPERATOR_SDK_VERSION=$$(sed -r 's|operator-sdk:\s*(.*)|\1|' REQUIREMENTS)

	echo "[INFO] ARCH: $$ARCH, OS: $$OS. operator-sdk version: $$OPERATOR_SDK_VERSION"

	if [ -z $(OP_SDK_DIR) ]; then
		OP_SDK_PATH="operator-sdk"
	else
		OP_SDK_PATH="$(OP_SDK_DIR)/operator-sdk"
	fi
	
	echo "[INFO] Downloading operator-sdk..."

	OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/$${OPERATOR_SDK_VERSION}
	curl -sSLo $${OP_SDK_PATH} $${OPERATOR_SDK_DL_URL}/operator-sdk_$${OS}_$${ARCH}

	echo "[INFO] operator-sdk will downloaded to: $${OP_SDK_PATH}"
	echo "[INFO] Set up executable permissions to binary."
	chmod +x $${OP_SDK_PATH}
	echo "[INFO] operator-sdk is ready."

removeRequiredAttribute:
	REQUIRED=false

	while IFS= read -r line;
	do
		if [ $${REQUIRED} = true ]; then
			case "$${line}" in
			*"- "*) continue  ;;
			*)      REQUIRED=false
			esac
		fi

		case "$${line}" in
		*"required:"*) REQUIRED=true; continue  ;;
		*)
		esac

		echo  "$${line}" >> $${filePath}.tmp
	done < "$${filePath}"
	mv $${filePath}.tmp $${filePath}

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	# Generate CRDs v1 and v2
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) $(CRD_BETA_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:stdout > config/crd/bases/org_v1_che_crd-v1beta1.yaml

	# Rename and patch CRDs
	cd config/crd/bases

	mv org.eclipse.che_checlusters.yaml org_v1_che_crd.yaml
	sed -i.bak '/---/d' org_v1_che_crd-v1beta1.yaml
	sed -i.bak '/---/d' org_v1_che_crd.yaml
	rm -rf org_v1_che_crd-v1beta1.yaml.bak org_v1_che_crd.yaml.bak

	cd ../../..

	$(MAKE) removeRequiredAttribute "filePath=config/crd/bases/org_v1_che_crd-v1beta1.yaml"

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.6.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

IMAGE_TOOL=docker

docker-build: test ## Build docker image with the manager.
	${IMAGE_TOOL} build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	${IMAGE_TOOL} push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager || true && $(KUSTOMIZE) edit set image controller=${IMG} && cd ../..
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

ENV_FILE="/tmp/che-operator-debug.env"
ECLIPSE_CHE_NAMESPACE="eclipse-che"
ECLIPSE_CHE_CR=config/samples/org.eclipse.che_v1_checluster.yaml
ECLIPSE_CHE_CRD=config/crd/bases/org_v1_che_crd.yaml
DEV_WORKSPACE_CONTROLLER_VERSION="main"
DEV_WORKSPACE_CHE_OPERATOR_VERSION="main"

prepare-templates:
	cp templates/keycloak-provision.sh /tmp/keycloak-provision.sh
	cp templates/delete-identity-provider.sh /tmp/delete-identity-provider.sh
	cp templates/create-github-identity-provider.sh /tmp/create-github-identity-provider.sh
	cp templates/oauth-provision.sh /tmp/oauth-provision.sh
	cp templates/keycloak-update.sh /tmp/keycloak-update.sh

	# Download Dev Workspace operator templates
	echo "[INFO] Downloading Dev Workspace operator templates ..."
	rm -f /tmp/devworkspace-operator.zip
	rm -rf /tmp/devfile-devworkspace-operator-*
	rm -rf /tmp/devworkspace-operator/
	mkdir -p /tmp/devworkspace-operator/templates

	curl -sL https://api.github.com/repos/devfile/devworkspace-operator/zipball/${DEV_WORKSPACE_CONTROLLER_VERSION} > /tmp/devworkspace-operator.zip

	unzip /tmp/devworkspace-operator.zip '*/deploy/deployment/*' -d /tmp
	cp -r /tmp/devfile-devworkspace-operator*/deploy/* /tmp/devworkspace-operator/templates
	echo "[INFO] Downloading Dev Workspace operator templates completed."

	# Download Dev Workspace Che operator templates
	echo "[INFO] Downloading Dev Workspace Che operator templates ..."
	rm -f /tmp/devworkspace-che-operator.zip
	rm -rf /tmp/che-incubator-devworkspace-che-operator-*
	rm -rf /tmp/devworkspace-che-operator/
	mkdir -p /tmp/devworkspace-che-operator/templates

	curl -sL https://api.github.com/repos/che-incubator/devworkspace-che-operator/zipball/${DEV_WORKSPACE_CHE_OPERATOR_VERSION} > /tmp/devworkspace-che-operator.zip

	unzip /tmp/devworkspace-che-operator.zip '*/deploy/deployment/*' -d /tmp
	cp -r /tmp/che-incubator-devworkspace-che-operator*/deploy/* /tmp/devworkspace-che-operator/templates
	echo "[INFO] Downloading Dev Workspace Che operator templates completed."

create-namespace:
	set +e
	kubectl create namespace ${ECLIPSE_CHE_NAMESPACE} || true
	set -e

apply-cr-crd:
	kubectl apply -f ${ECLIPSE_CHE_CRD}
	kubectl apply -f ${ECLIPSE_CHE_CR} -n ${ECLIPSE_CHE_NAMESPACE}

create-env-file: prepare-templates
	rm -rf "${ENV_FILE}"
	touch "${ENV_FILE}"
	CLUSTER_API_URL=$$(oc whoami --show-server=true) || true;
	if [ -n $${CLUSTER_API_URL} ]; then
		echo "CLUSTER_API_URL='$${CLUSTER_API_URL}'" >> "${ENV_FILE}"
		echo "[INFO] Set up cluster api url: $${CLUSTER_API_URL}"
	fi;
	echo "WATCH_NAMESPACE='${ECLIPSE_CHE_NAMESPACE}'" >> "${ENV_FILE}"

create-full-env-file: create-env-file
	cat ./config/default/manager_auth_proxy_patch.yaml | \
	yq -r '.spec.template.spec.containers[1].env[] | "export \(.name)=\"$${\(.name):-\(.value)}\""' \ # Todo "metrics proxy" 0 => 1
	>> ${ENV_FILE}
	echo "[INFO] Env file: ${ENV_FILE}"
	source ${ENV_FILE} ; env | grep CHE_VERSION

debug: prepare-templates create-namespace apply-cr-crd create-env-file manifests kustomize
	echo "[WARN] Make sure that your CR contains valid ingress domain!"
	# dlv has an issue with 'Ctrl-C' termination, that's why we're doing trick with detach.
	dlv debug --listen=:2345 --headless=true --api-version=2 ./main.go -- &
	OPERATOR_SDK_PID=$!
	echo "[INFO] Use 'make uninstall' to remove Che installation after debug"
	wait $$OPERATOR_SDK_PID


CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "[INFO] Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

NIGHTLY_CHANNEL="nightly"

.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	if [ -z "$(platform)" ]; then
		echo "[INFO] You must specify 'platform' macros. For example: `make bundle platform=kubernetes`"
		exit 1
	fi

	BUNDLE_PACKAGE="eclipse-che-preview-$(platform)"
	BUNDLE_DIR="bundle/$(DEFAULT_CHANNEL)/$${BUNDLE_PACKAGE}"
	GENERATED_CSV_NAME=$${BUNDLE_PACKAGE}.clusterserviceversion.yaml
	DESIRED_CSV_NAME=che-operator.clusterserviceversion.yaml
	GENERATED_CRD_NAME=org.eclipse.che_checlusters.yaml
	DESIRED_CRD_NAME=org_v1_che_crd.yaml

	$(OPERATOR_SDK_BINARY) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG) && cd ../..
	$(KUSTOMIZE) build config/platforms/$(platform) | \
	$(OPERATOR_SDK_BINARY) generate bundle \
	-q --overwrite --version $(VERSION) \
	--package $${BUNDLE_PACKAGE} \
	--output-dir $${BUNDLE_DIR} \
	$(BUNDLE_METADATA_OPTS)

	rm -rf bundle.Dockerfile

	cd $${BUNDLE_DIR}/manifests;
	mv $${GENERATED_CSV_NAME} $${DESIRED_CSV_NAME}
	mv $${GENERATED_CRD_NAME} $${DESIRED_CRD_NAME}
	cd $(mkfile_dir)

	$(OPERATOR_SDK_BINARY) bundle validate ./$${BUNDLE_DIR}

bundles:
	$(shell ./olm/update-resources.sh)

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	if [ -z "$(platform)" ]; then
		echo "[INFO] You must specify 'platform' macros. For example: `make bundle platform=kubernetes`"
		exit 1
	fi
	BUNDLE_PACKAGE="eclipse-che-preview-$(platform)"
	BUNDLE_DIR="bundle/$(DEFAULT_CHANNEL)/$${BUNDLE_PACKAGE}"
	docker build -f $${BUNDLE_DIR}/bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=quay.io/eclipse/operator-bundle:v0.1.0,quay.io/eclipse/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=quay.io/eclipse/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add \
	--build-tool $(IMAGE_TOOL) \
	--bundles $(BUNDLE_IMGS) \
	--tag $(CATALOG_IMG) \
	--pull-tool $(IMAGE_TOOL) \
	--binary-image=quay.io/operator-framework/upstream-opm-builder:v1.15.1 \
	--mode semver $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)
