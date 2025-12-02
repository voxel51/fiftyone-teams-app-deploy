//go:build docker || compose || integration || integrationComposeInternalAuth || integrationComposeLegacyAuth || integrationComposeInternalAuth
// +build docker compose integration integrationComposeInternalAuth integrationComposeLegacyAuth integrationComposeInternalAuth

package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
)

const (
	legacyAuthEnvFixtureFilePath   = "../../fixtures/docker/integration_legacy_auth.env"
	internalAuthFixtureEnvFilePath = "../../fixtures/docker/integration_internal_auth.env"
	overrideFile                   = "../../tests/fixtures/docker/compose.override.yaml"
	mongodbComposeFile             = "../../tests/fixtures/docker/compose.override.mongodb.yaml"
	mongodbComposeFilePlugins      = "../../tests/fixtures/docker/compose.override.mongodb_plugins.yaml"
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

func checkContainerLogsWithRetries(subT *testing.T, dockerOptions *docker.Options, container string, tc string, expected string) {
	maxRetries := 6
	retryDelay := 2 * time.Second

	var log string

	for i := 0; i < maxRetries; i++ {
		log = get_logs(subT, dockerOptions, container)
		if strings.Contains(log, expected) {
			// Log entry found, proceed to next container
			break
		}

		// Log entry not found, wait before retrying
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	// Final assertion
	subT.Run(fmt.Sprintf("%s - %s", tc, container), func(t *testing.T) {
		if !strings.Contains(log, expected) {
			t.Errorf("%s - %s - log should contain matching entry:\n\t%s", tc, container, expected)
		}
	})
}
