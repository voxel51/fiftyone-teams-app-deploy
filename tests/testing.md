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
    make asdf
    ```

## Unit Tests

The unit tests are named after the corresponding Helm templates.

For example, the test
`tests/unit/api-deployment_test.go`
covers the Helm template
`helm/fiftyone-teams-app/templates/api-deployment.yaml`.

### Running Unit Tests

Run tests tagged with `unit` or the more specific
tag (found at the top of the test file).

* Without interleaved test output (good for rapid testing cycle)

    ```shell
    # From repo root
    make test-unit
    ```

* With interleaved test output (good for CI runs)

    ```shell
    # From repo root
    cd test/unit

    # replace `unit` with any build tag
    go test -v -timeout 30m -tags unit
    ```

### Writing Unit Tests

To avoid code duplication, consider
adding items to `tests/unit/common_test.go`.
When adding a new build tag, add the new tag to
the test file and to `tests/unit/common_test.go`.
Currently, `common_test.go` contains the
variable `chartPath` used in all of the tests.

For structures (structs), there are two approaches.
Either write

* Go code referencing the type for each field
* JSON (easily converted from YAML) and unmarshall it into  the struct

See
[Debugging interleaved test output](https://terratest.gruntwork.io/docs/testing-best-practices/debugging-interleaved-test-output/#installing-the-utility-binaries).

## Additional Links

* [Automated Testing for Kubernetes and Helm Charts using Terratest](https://github.com/gruntwork-io/terratest-helm-testing-example)
* [terratest/examples/helm-basic-example](https://github.com/gruntwork-io/terratest/tree/master/examples/helm-basic-example)

* [A Tour of Go](https://go.dev/tour/)
* Kubernetes API Library
  * [apps/v1 Deployment](https://pkg.go.dev/k8s.io/api/apps/v1#Deployment)
