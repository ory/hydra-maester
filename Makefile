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
		ifeq ($(shell uname -m),x86_64)
			ARCH=amd64
		endif
		ifeq ($(shell uname -m),arm64)
			ARCH=arm64
		endif
	endif
endif
##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/.bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

SHELL=/bin/bash -euo pipefail

export PATH := .bin:${PATH}
export PWD := $(shell pwd)
export K3SIMAGE := docker.io/rancher/k3s:v1.26.1-k3s1
## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.11.3
ENVTEST_K8S_VERSION = 1.26.1

# Image URL to use all building/pushing image targets
IMG ?= controller:latest

run-with-cleanup = $(1) && $(2) || (ret=$$?; $(2) && exit $$ret)

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
	curl https://raw.githubusercontent.com/ory/meta/master/install.sh | bash -s -- -b .bin ory
	touch .bin/ory

.bin/kubectl: Makefile
	@URL=$$(.bin/ory dev ci deps url -o ${OS} -a ${ARCH} -c .deps/kubectl.yaml); \
	echo "Downloading 'kubectl' $${URL}...."; \
	curl -Lo .bin/kubectl $${URL}; \
	chmod +x .bin/kubectl;

.bin/kustomize: Makefile
	@URL=$$(.bin/ory dev ci deps url -o ${OS} -a ${ARCH} -c .deps/kustomize.yaml); \
	echo "Downloading 'kustomize' $${URL}...."; \
	curl -L $${URL} | tar -xmz -C .bin kustomize; \
	chmod +x .bin/kustomize;

.bin/k3d: Makefile
	@URL=$$(.bin/ory dev ci deps url -o ${OS} -a ${ARCH} -c .deps/k3d.yaml); \
	echo "Downloading 'k3d' $${URL}...."; \
	curl -Lo .bin/k3d $${URL}; \
	chmod +x .bin/k3d;

.PHONY: deps
deps: .bin/ory .bin/k3d .bin/kubectl .bin/kustomize

.PHONY: all
all: manager

# Run tests
.PHONY: test
test: manifests generate vet envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: k3d-up
k3d-up:
	k3d cluster create --image $${K3SIMAGE} ory \
		--k3s-arg=--kube-apiserver-arg="enable-admission-plugins=NodeRestriction,ServiceAccount@server:0" \
		--k3s-arg=feature-gates="NamespaceDefaultLabelName=true@server:0";

.PHONY: k3d-down
k3d-down:
	k3d cluster delete ory || true

.PHONY: k3d-deploy
k3d-deploy: manager manifests docker-build-notest k3d-up
	kubectl config set-context k3d-ory
	k3d image load controller:latest -c ory
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

.PHONY: k3d-test
k3d-test: k3d-deploy
	kubectl config set-context k3d-ory
	go install github.com/onsi/ginkgo/ginkgo@latest
	USE_EXISTING_CLUSTER=true ginkgo -v ./controllers/...

# Run integration tests on local cluster
.PHONY: test-integration
test-integration:
	$(call run-with-cleanup, $(MAKE) k3d-test, $(MAKE) k3d-down)

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
deploy: manifests
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

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

licenses: .bin/licenses node_modules  # checks open-source licenses
	.bin/licenses

.bin/licenses: Makefile
	curl https://raw.githubusercontent.com/ory/ci/master/licenses/install | sh

node_modules: package-lock.json
	npm ci
	touch node_modules
