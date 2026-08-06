<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Ory Hydra Maester](#ory-hydra-maester)
  - [Prerequisites](#prerequisites)
  - [Design](#design)
  - [How to use it](#how-to-use-it)
    - [Command-line flags](#command-line-flags)
    - [Environmental Variables](#environmental-variables)
    - [Externally managed secrets](#externally-managed-secrets)
  - [Development](#development)
    - [Testing](#testing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Ory Hydra Maester

⚠️ ⚠️ ⚠️

> Ory Hydra Maester is developed by the Ory community and is not actively
> maintained by Ory core maintainers due to lack of resources, time, and
> knolwedge. As such please be aware that there might be issues with the system.
> If you have ideas for better testing and development principles please open an
> issue or PR!

⚠️ ⚠️ ⚠️

This project contains a Kubernetes controller that uses Custom Resources (CR) to
manage Hydra Oauth2 clients. ORY Hydra Maester watches for instances of
`oauth2clients.hydra.ory.sh/v1alpha1` CR and creates, updates, or deletes
corresponding OAuth2 clients by communicating with ORY Hydra's API.

Visit Hydra-maester's
[chart documentation](https://github.com/ory/k8s/blob/master/docs/helm/hydra-maester.md)
and view [sample OAuth2 client resources](config/samples) to learn more about
the `oauth2clients.hydra.ory.sh/v1alpha1` CR.

The project is based on
[Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder).

## Prerequisites

- recent version of Go language with support for modules (e.g: 1.12.6)
- make
- kubectl
- kustomize
- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) for running
  tests
- [ginkgo](https://onsi.github.io/ginkgo/) for local integration testing
- access to K8s environment: minikube or a remote K8s cluster
- [mockery](https://github.com/vektra/mockery) to generate mocks for testing
  purposes

## Design

Take a look at [Design Readme](./docs/README.md).

## How to use it

- `make test` to run tests
- `make test-integration` to run integration tests
- `make install` to generate CRD file from go sources and install it on the
  cluster
- `export HYDRA_URL={HYDRA_SERVICE_URL} && make run` to run the controller

To deploy the controller, edit the value of the `--hydra-url` argument in the
[manager.yaml](config/manager/manager.yaml) file and run `make deploy`.

### Command-line flags

| Name                         | Required | Description                                                                                                                                 | Default value | Example values                           |
| ---------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ------------- | ---------------------------------------- |
| **hydra-url**                | yes      | ORY Hydra's service address                                                                                                                 | -             | ` ory-hydra-admin.ory.svc.cluster.local` |
| **hydra-port**               | no       | ORY Hydra's service port                                                                                                                    | `4445`        | `4445`                                   |
| **tls-trust-store**          | no       | TLS cert path for hydra client                                                                                                              | `""`          | `/etc/ssl/certs/ca-certificates.crt`     |
| **insecure-skip-verify**     | no       | Skip http client insecure verification                                                                                                      | `false`       | `true` or `false`                        |
| **namespace**                | no       | Namespace in which the controller should operate. Setting this will make the controller ignore other namespaces.                            | `""`          | `"my-namespace"`                         |
| **leader-elector-namespace** | no       | Leader elector namespace where controller should be set.                                                                                    | `""`          | `"my-namespace"`                         |
| **require-existing-secret**  | no       | Wait for the secret referenced by an OAuth2Client instead of generating one. See [Externally managed secrets](#externally-managed-secrets). | `false`       | `true` or `false`                        |

### Environmental Variables

| Variable name                 | Default value       | Example value         |
| :---------------------------- | ------------------- | --------------------- |
| `**CLIENT_ID_KEY**`           | `**CLIENT_ID**`     | `**MY_SECRET_NAME**`  |
| `**CLIENT_SECRET_KEY**`       | `**CLIENT_SECRET**` | `**MY_SECRET_VALUE**` |
| `**REQUIRE_EXISTING_SECRET**` | `**false**`         | `**true**`            |

### Externally managed secrets

By default, when the secret referenced by `spec.secretName` does not exist, the
controller registers a new OAuth2 client with a Hydra-generated secret and
creates that Kubernetes secret itself, owned by the `OAuth2Client` resource.

That is the wrong behaviour when the secret is provisioned by something else,
for example the External Secrets Operator, sealed-secrets, or any other
controller syncing it from an external vault. If the `OAuth2Client` reconciles
before the secret has been materialized, hydra-maester wins the race and the two
controllers end up fighting over the same secret.

Set `--require-existing-secret` (or `REQUIRE_EXISTING_SECRET=true`) to make the
controller wait for the secret instead. While the secret is missing, the client
is not registered in Hydra, no secret is created, and the resource reports:

```yaml
status:
  conditions:
    - type: Ready
      status: "False"
  reconciliationError:
    statusCode: SECRET_NOT_FOUND
    description: secret my-namespace/hydra-client-oauth2-secret does not exist
```

The resource is requeued with an exponential backoff, starting at 15 seconds and
capped at 5 minutes, and reconciles normally as soon as the secret appears.

Individual resources can override the controller-wide default through
`spec.requireExistingSecret`:

```yaml
apiVersion: hydra.ory.sh/v1alpha1
kind: OAuth2Client
metadata:
  name: my-oauth2-client
spec:
  secretName: hydra-client-oauth2-secret
  requireExistingSecret: true
  # ...
```

## Development

### Testing

Use mockery to generate mock types that implement existing interfaces. To
generate a mock type for an interface, navigate to the directory containing that
interface and run this command:

```
mockery -name={INTERFACE_NAME}
```
