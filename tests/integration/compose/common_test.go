//go:build docker || compose || integration || integrationComposeInternalAuth || integrationComposeLegacyAuth || integrationComposeInternalAuth
// +build docker compose integration integrationComposeInternalAuth integrationComposeLegacyAuth integrationComposeInternalAuth

package integration

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
)

const (
	legacyAuthEnvFixtureFilePath   = "../../fixtures/docker/integration_legacy_auth.env"
	internalAuthFixtureEnvFilePath = "../../fixtures/docker/integration_internal_auth.env"
	// Override FIFTYONE_DATABASE_ADMIN to true (until we can override via environment variable)
	overrideFile = "../../tests/fixtures/docker/compose.override.yaml"
	// To run the containers on macOS arm64, we need to set the platform
	darwinOverrideFile = "../../tests/fixtures/docker/compose.override.darwin.yaml"
	mongodbComposeFile = "../../tests/fixtures/docker/compose.override.mongodb.yaml"
)

type serviceValidations struct {
	name             string
	url              string
	responsePayload  string
	httpResponseCode int
	log              string
}

func validate_endpoint(t *testing.T, url string, expectedBody string, expectedStatus int) {
	maxRetries := 10
	timeBetweenRetries := 3 * time.Second
	http_helper.HttpGetWithRetry(t, url, nil, expectedStatus, expectedBody, maxRetries, timeBetweenRetries)
}

func get_logs(t *testing.T, dockerOptions *docker.Options, container string) string {
	output := docker.RunDockerComposeAndGetStdOut(t, dockerOptions, "logs", container)
	return output
}
