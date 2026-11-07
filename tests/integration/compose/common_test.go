//go:build docker || compose || integration || integrationCompose
// +build docker compose integration integrationCompose

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
	fixtureEnvFilePath        = "../../fixtures/docker/integration.env"
	overrideFile              = "../tests/fixtures/docker/compose.override.yaml"
	mongodbComposeFile        = "../tests/fixtures/docker/compose.override.mongodb.yaml"
	mongodbComposeFilePlugins = "../tests/fixtures/docker/compose.override.mongodb_plugins.yaml"
	mongodbComposeFileDO      = "../tests/fixtures/docker/compose.override.mongodb_do.yaml"
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
	// Workers such as teams-do can take well over 30 seconds
	// to log their startup line on a shared CI runner.
	timeout := 2 * time.Minute
	retryDelay := 5 * time.Second
	deadline := time.Now().Add(timeout)

	var log string

	for {
		log = get_logs(subT, dockerOptions, container)
		if strings.Contains(log, expected) || time.Now().After(deadline) {
			break
		}
		time.Sleep(retryDelay)
	}

	// Final assertion
	subT.Run(fmt.Sprintf("%s - %s", tc, container), func(t *testing.T) {
		if !strings.Contains(log, expected) {
			t.Errorf("%s - %s - log should contain matching entry:\n\t%s", tc, container, expected)
		}
	})
}
