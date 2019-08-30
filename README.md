# Hydra-maester


This project contains a Kubernetes controller that uses Custom Resources to manage Hydra Oauth2 clients.
ORY Hydra Maester watches for instances of `oauth2clients.oathkeeper.ory.sh/v1alpha1` custom resource (CR) and creates, updates, or deletes corresponding OAuth2 clients by communicating with ORY Hydra API.

Visit Hydra-maester's [chart documentation](https://github.com/ory/k8s/blob/master/docs/helm/hydra-maester.md) and view a [sample OAuth2 client resource](./config/samples/hydra_v1alpha1_oauth2client.yaml) to learn more about the `oauth2clients.oathkeeper.ory.sh/v1alpha1` CR. 

The project is based on [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder).

## Prerequisites

- recent version of Go language with support for modules (e.g: 1.12.6)
- make
- kubectl
- kustomize
- [ginkgo](https://onsi.github.io/ginkgo/) for local integration testing
- access to K8s environment: minikube or a remote K8s cluster



## Design

Take a look at [Design Readme](./docs/README.md).

## How to use it

- `make test` to run tests
- `make test-integration` to run integration tests
- `make install` to generate CRD file from go sources and install it on the cluster
- `export HYDRA_URL={HYDRA_SERVICE_URL} && make run` to run the controller

To deploy the controller, edit the value of the ```--hydra-url``` argument in the [manager.yaml](config/manager/manager.yaml) file and run ```make deploy```.

### Command-line flags

| Name            | Required | Description                  | Default value | Example values                                       |
|-----------------|----------|------------------------------|---------------|------------------------------------------------------|
| **hydra-url**   | yes      | ORY Hydra's service address  | -             | ` ory-hydra-admin.ory.svc.cluster.local`             |
| **hydra-port**  | no       | ORY Hydra's service port     | `4445`        | `4445`                                               |