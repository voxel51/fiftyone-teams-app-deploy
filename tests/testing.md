# Testing

Test types

* Unit
  * Quick
  * Test Helm template logic
  * Validate Helm values result in the expected rendered YAML kubernetes resources
  * Helm template syntax checks
* Integration
  * Slower as they require installation into a Kubernetes cluster
  * Deploy to the rendered YAML to Kubernetes and validate functionality
  * Validate kubernetes resources via queries

Please use Test Driven Development and write tests
before making changes to the Helm chart.

## Tools

[Terratest](https://terratest.gruntwork.io/docs/#getting-started)
is a Go Library for testing Infrastructure as Code.
Terratest supports tests for Helm Charts.

Go test supports
[Subtests](https://go.dev/blog/subtests)
with table-driven tests.

Since we use Go test's parallelism, we use
[terratest_log_parser](https://terratest.gruntwork.io/docs/testing-best-practices/debugging-interleaved-test-output/)
to organize the interleaved test output.
This is instrumented in our make targets.

The library
[Testify](https://github.com/stretchr/testify)
provides suite features for better test organization.

We implemented various patterns from these tools in our tests.

## Environment Initialization

1. Install go

    ```shell
    # from repo root
    make asdf
    ```

## Unit Tests

The unit tests are named after the corresponding Helm templates.

For example, the test
`tests/unit/helm/api-deployment_test.go`
covers the Helm template
`helm/fiftyone-teams-app/templates/api-deployment.yaml`.

### Running Unit Tests

Run tests tagged with `unit` or the more specific
tag (found at the top of the test file).

* Without interleaved test output (good for rapid testing cycle)

    ```shell
    # From repo root
    make test-unit-compose
    make test-unit-helm
    ```

* With interleaved test output (good for CI runs)

    ```shell
    # From repo root
    cd test/unit/helm

    # replace the tag `unit` with any build tag
    go test -count=1 -timeout=3m -v -tags unit
    ```

* To run a specific test function,
  for example within `plugins-service_test.go`,
  matching the regex of the test function name `TestMetadataLabels`

    ```shell
    cd test/unit/helm
    go test \
      -count=1 \
      -timeout=30s \
      -v \
      -tags unit \
      plugins-service_test.go \
      common_test.go \
      chartInfo.go \
      -testify.m '^(TestMetadataLabels|)$'
    ```

    > **Note:** Include the files `common_test.go` and `chartInfo.go`
    > to avoid `undefined` errors for `chartPath` and `chartInfo`
    > **Note:** We pass `-count=1` to disable test caching.

### Writing Unit Tests

To avoid code duplication, consider
adding items to `tests/unit/helm/common_test.go`.
When adding a new build tag, add the new tag to
the test file and to `tests/unit/helm/common_test.go`.
Currently, `common_test.go` contains the
variable `chartPath` used in all of the tests.

For structures (structs), there are two approaches.
Either write

* Go code referencing the type for each field
* JSON (easily converted from YAML) and unmarshall it into the struct

See
[Debugging interleaved test output](https://terratest.gruntwork.io/docs/testing-best-practices/debugging-interleaved-test-output/#installing-the-utility-binaries).

## Integration Tests

### Running Docker Compose Integration Tests

1. Have Docker desktop running
1. Copy the 'Voxel51 GitHub Legacy' license file to `docker/legacy-license.key`
   You can retrieve this from the
   [Voxel51 License Management](https://computer-vision-team.uc.r.appspot.com/)
   UI
1. Copy the 'Voxel51 GitHub Internal' license file to
   `docker/internal-license.key`
   You can retrieve this from the
   [Voxel51 License Management](https://computer-vision-team.uc.r.appspot.com/)
   UI
1. Run tests

    ```shell
    make auth test-integration-compose
    ```

### Running Helm Integration Tests

1. Start minikube

    ```shell
    make auth start
    ```

1. Install cert-manager and mongodb into minikube

    ```shell
    make run-cert-manager run-mongodb
    ```

1. In another terminal, run `minikube tunnel`
   (to expose the services within minikube outside of minikube)

    ```shell
    make tunnel
    ```

    > **NOTE**: This command will prompt for sudo permission
    > on systems where 80 and 443 are privileged ports

1. Copy the 'Voxel51 GitHub Internal' license file and convert it to a
   kubernetes secret

   ```shell
   gcloud storage cp gs://voxel51-test/licenses/299a423b/1/license.key \
     internal-license.key
   kubectl --namespace your-namepace create secret generic fiftyonelicense \
     --from-file=license=./internal-license.key
   ```

1. Run tests

    ```shell
    make test-integration-helm
    ```

## Additional Links

* [Automated Testing for Kubernetes and Helm Charts using Terratest](https://github.com/gruntwork-io/terratest-helm-testing-example)
* [terratest/examples/helm-basic-example](https://github.com/gruntwork-io/terratest/tree/master/examples/helm-basic-example)

* [A Tour of Go](https://go.dev/tour/)
* Kubernetes API Library
  * [apps/v1 Deployment](https://pkg.go.dev/k8s.io/api/apps/v1#Deployment)

## Testing Setup

* [Install delve](https://github.com/go-delve/delve/tree/master/Documentation/installation)

## VSCode Setup

* [Debugging](https://github.com/golang/vscode-go/wiki/debugging)
