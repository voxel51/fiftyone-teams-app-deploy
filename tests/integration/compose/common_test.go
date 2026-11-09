//go:build docker || compose || integration || integrationCompose
// +build docker compose integration integrationCompose

package integration

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/shell"
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
	// orchestrator is the delegated operator instance ID
	// that the service must register after it starts.
	orchestrator string
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
	// Some services take well over 30 seconds
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

// orchestratorUpdatedAtScript prints the orchestrator's updated_at
// as Unix seconds, or an empty line when it isn't registered.
const orchestratorUpdatedAtScript = `import datetime, sys
from fiftyone.core.odm import get_db_conn
d = get_db_conn().orchestrators.find_one({"instance_id": sys.argv[1]})
u = d.get("updated_at") if d else None
print(u.replace(tzinfo=datetime.timezone.utc).timestamp() if u else "")`

// checkOrchestratorRegisteredWithRetries waits for a delegated operator worker
// to register its orchestrator in MongoDB after its container starts.
// From v2.26, teams-do no longer logs "Registering executor <id>",
// so the registration record is the reliable signal.
func checkOrchestratorRegisteredWithRetries(subT *testing.T, dockerOptions *docker.Options, composeArgs []string, container string, tc string, instanceID string) {
	timeout := 2 * time.Minute
	retryDelay := 5 * time.Second
	deadline := time.Now().Add(timeout)

	containerName := fmt.Sprintf("%s-%s-1", dockerOptions.ProjectName, container)
	startedAtOutput := shell.RunCommandAndGetOutput(subT, shell.Command{
		Command: "docker",
		Args:    []string{"inspect", "--format", "{{.State.StartedAt}}", containerName},
	})
	startedAt, startedAtErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(startedAtOutput))

	execArgs := append([]string{}, composeArgs...)
	execArgs = append(execArgs, "exec", "-T", container, "python", "-c", orchestratorUpdatedAtScript, instanceID)

	var updatedAt time.Time
	registered := false
	for startedAtErr == nil {
		output, err := docker.RunDockerComposeE(subT, dockerOptions, execArgs...)
		if err == nil {
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if seconds, parseErr := strconv.ParseFloat(strings.TrimSpace(lines[len(lines)-1]), 64); parseErr == nil {
				updatedAt = time.Unix(0, int64(seconds*float64(time.Second))).UTC()
				// Allow for the database timestamp's millisecond precision.
				if !updatedAt.Before(startedAt.Add(-time.Second)) {
					registered = true
				}
			}
		}
		if registered || time.Now().After(deadline) {
			break
		}
		time.Sleep(retryDelay)
	}

	// Final assertion
	subT.Run(fmt.Sprintf("%s - %s - orchestrator %s", tc, container, instanceID), func(t *testing.T) {
		if startedAtErr != nil {
			t.Fatalf("%s - %s - could not read the container start time %q: %v", tc, container, startedAtOutput, startedAtErr)
		}
		if !registered {
			t.Errorf("%s - %s - orchestrator %q should be registered after the container started at %s (last updated_at: %s)", tc, container, instanceID, startedAt, updatedAt)
		}
	})
}
