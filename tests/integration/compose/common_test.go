//go:build docker || compose || integration || integrationComposeInternalAuth || integrationComposeLegacyAuth
// +build docker compose integration integrationComposeInternalAuth integrationComposeLegacyAuth

package integration

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
)

const (
	envFixtureFilePath = "../../fixtures/docker/integration_legacy_auth.env"
)

func validate_endpoint(t *testing.T, url string, expectedBody string, expectedStatus int) {
	maxRetries := 10
	timeBetweenRetries := 3 * time.Second
	http_helper.HttpGetWithRetry(t, url, nil, expectedStatus, expectedBody, maxRetries, timeBetweenRetries)
}

func get_logs(t *testing.T, dockerOptions *docker.Options, container string) string {
	output := docker.RunDockerComposeAndGetStdOut(t, dockerOptions, "logs", container)
	return output
}
