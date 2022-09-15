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

HELL=/bin/bash -o pipefail
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,crdVersions=v1"

run-with-cleanup = $(1) && $(2) || (ret=$$?; $(2) && exit $$ret)

.PHONY: all
all: manager

# Run tests
.PHONY: test
test: generate fmt vet manifests
	go test ./api/... ./controllers/... ./hydra/... ./helpers/... -coverprofile cover.out

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
kind-deploy: manager manifests docker-build-notest kind-start
	kubectl config set-context kind-kind
	kind load docker-image controller:latest
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

# private
.PHONY: kind-test
kind-test: kind-deploy
	kubectl config set-context kind-kind
	go get github.com/onsi/ginkgo/ginkgo
	ginkgo -v ./controllers/...

# Run integration tests on local KIND cluster
.PHONY: test-integration
test-integration:
	$(call run-with-cleanup, $(MAKE) kind-test, $(MAKE) kind-stop)

# Build manager binary
.PHONY: manager
manager: generate fmt vet
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate fmt vet
	go run ./main.go --hydra-url ${HYDRA_URL}

# Install CRDs into a cluster
.PHONY: install
install: manifests
	kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: manifests
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
format: node_modules
	go fmt ./...
	npm exec -- prettier --write .

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# Generate code
.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

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

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0
CONTROLLER_GEN=$(shell which controller-gen)
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Download and setup kubebuilder
.PHONY: kubebuilder
kubebuilder:
	curl -sL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.2/kubebuilder_2.3.2_${OS}_${ARCH}.tar.gz | tar -xz -C /tmp/
	mv /tmp/kubebuilder_2.3.2_${OS}_${ARCH} ${PWD}/.bin/kubebuilder
	export PATH=${PATH}:${PWD}/.bin/kubebuilder/bin

node_modules: package-lock.json
	npm ci
	touch node_modules
