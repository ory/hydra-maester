ifeq ($(OS),Windows_NT)
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		ARCH=amd64
		OS=windows
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OS=linux
		ARCH=amd64
	endif
	ifeq ($(UNAME_S),Darwin)
		OS=darwin
		ARCH=amd64
	endif
endif

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/.bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.0
CONTROLLER_TOOLS_VERSION ?= v0.11.3
ENVTEST_K8S_VERSION = 1.26.1

HELL=/bin/bash -o pipefail
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

run-with-cleanup = $(1) && $(2) || (ret=$$?; $(2) && exit $$ret)

.PHONY: all
all: manager

# Run tests
.PHONY: test
test: manifests generate vet envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

# Start KIND pseudo-cluster
.PHONY: kind-start
kind-start:
	kind create cluster

# Stop KIND pseudo-cluster
.PHONY: kind-stop
kind-stop:
	kind delete cluster

# Deploy on KIND
# Ensures the controller image is built, deploys the image to KIND cluster along with necessary configuration
.PHONY: kind-deploy
kind-deploy: manager manifests docker-build-notest kind-start kustomize
	kubectl config set-context kind-kind
	kind load docker-image controller:latest
	kubectl apply -f config/crd/bases
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# private
.PHONY: kind-test
kind-test: kind-deploy
	kubectl config set-context kind-kind
	go install github.com/onsi/ginkgo/ginkgo@latest
	USE_EXISTING_CLUSTER=true ginkgo -v ./controllers/...

# Run integration tests on local KIND cluster
.PHONY: test-integration
test-integration:
	$(call run-with-cleanup, $(MAKE) kind-test, $(MAKE) kind-stop)

# Build manager binary
.PHONY: manager
manager: generate vet
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -a -o manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate vet
	go run ./main.go --hydra-url ${HYDRA_URL}

# Install CRDs into a cluster
.PHONY: install
install: manifests
	kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: manifests kustomize
	kubectl apply -f config/crd/bases
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Format the source code
format: .bin/ory node_modules
	.bin/ory dev headers copyright --type=open-source
	go fmt ./...
	npm exec -- prettier --write .

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# Generate code
.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
.PHONY: docker-build-notest
docker-build-notest:
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

.PHONY: docker-build
docker-build: test docker-build-notest

# Push the docker image
.PHONY: docker-push
docker-push:
	docker push ${IMG}

## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) --output install_kustomize.sh && bash install_kustomize.sh $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); rm install_kustomize.sh; }

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

## Download envtest-setup locally if necessary.
.PHONY: envtest
envtest: $(ENVTEST)
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.bin/ory: Makefile
	curl https://raw.githubusercontent.com/ory/meta/master/install.sh | bash -s -- -b .bin ory v0.1.48
	touch .bin/ory

licenses: .bin/licenses node_modules  # checks open-source licenses
	.bin/licenses

.bin/licenses: Makefile
	curl https://raw.githubusercontent.com/ory/ci/master/licenses/install | sh

node_modules: package-lock.json
	npm ci
	touch node_modules
