//go:build docker || compose || unit || unitCompose
// +build docker compose unit unitCompose

package unit

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var inferenceComposeDirs = []string{dockerDir}

var inferenceEnv = []string{
	"INFERENCE_SHARED_TOKEN=test-0123456789abcdef0123456789abcdef",
}

func inferenceProject(t *testing.T, dir string, env []string, profiles []string, extra ...string) (*types.Project, error) {
	files := []string{filepath.Join(dir, "compose.yaml"), filepath.Join(dir, "compose.inference.yaml")}
	for _, file := range extra {
		files = append(files, filepath.Join(dir, file))
	}
	options, err := cli.NewProjectOptions(
		files,
		cli.WithWorkingDirectory(dir),
		cli.WithName("fiftyone-compose-test"),
		cli.WithEnvFiles(filepath.Join(dir, "env.template"), envFixtureFilePath),
		cli.WithDotEnv,
		cli.WithEnv(env),
		cli.WithProfiles(profiles),
	)
	require.NoError(t, err)
	return cli.ProjectFromOptions(context.TODO(), options)
}

func TestDockerComposeInference(t *testing.T) {
	t.Parallel()

	for _, dir := range inferenceComposeDirs {
		project, err := inferenceProject(t, dir, inferenceEnv, nil)
		require.NoError(t, err, dir)

		// No labeler without its profile, and the model server is its own file
		require.Contains(t, project.ServiceNames(), "inference", dir)
		require.NotContains(t, project.ServiceNames(), "model-server", dir)
		require.NotContains(t, project.ServiceNames(), "inference-labeler", dir)
		require.NotContains(t, project.ServiceNames(), "inference-labeler-runs", dir)
		require.NotContains(t, project.Volumes, "model-weights", dir)

		inference, err := project.GetService("inference")
		require.NoError(t, err)
		require.Equal(t, "voxel51/inference:v2.28.0", inference.Image)
		require.Equal(t, "shared-token", *inference.Environment["AUTH_MODE"])
		require.Equal(t, "test-0123456789abcdef0123456789abcdef", *inference.Environment["AUTH_SHARED_TOKEN"])
		require.Equal(t, "license", *inference.Environment["BILLING_MODE"])
		require.Equal(t, "/opt/fiftyone/licenses/license", *inference.Environment["LICENSE_KEY_FILE_PATHS"])
		require.Equal(t, "memory", *inference.Environment["STORE_MODE"])
		require.True(t, strings.Contains(*inference.Environment["BACKENDS"], "baseUrl: http://model-server:8080"), dir)
		require.Equal(t, "/opt/fiftyone/licenses", inference.Volumes[0].Target)
		require.True(t, inference.Volumes[0].ReadOnly, dir)
		require.Equal(t, "8080", *inference.Environment["PORT"])
		require.Equal(t, "67108864", *inference.Environment["MAX_BODY_BYTES"])
		require.Equal(t, "always", inference.Restart)
		require.Equal(t, "6m0s", inference.StopGracePeriod.String())
		// Reached over the project network only
		require.Empty(t, inference.Ports, dir)

		// A model server on another worker
		project, err = inferenceProject(t, dir, append([]string{"MODEL_SERVER_URL=http://10.0.0.5:8080"}, inferenceEnv...), nil)
		require.NoError(t, err, dir)
		inference, err = project.GetService("inference")
		require.NoError(t, err)
		require.True(t, strings.Contains(*inference.Environment["BACKENDS"], "baseUrl: http://10.0.0.5:8080"), dir)

		// On this worker, beside it
		project, err = inferenceProject(t, dir, append([]string{modelServerImageEnv}, inferenceEnv...), nil, "compose.model-server.yaml")
		require.NoError(t, err, dir)
		require.Contains(t, project.ServiceNames(), "model-server", dir)
		require.Contains(t, project.ServiceNames(), "inference", dir)

		project, err = inferenceProject(t, dir, inferenceEnv, []string{"labeler"})
		require.NoError(t, err, dir)
		labeler, err := project.GetService("inference-labeler")
		require.NoError(t, err, dir)
		require.Equal(t, "voxel51/agentic-labeler-cpu:v2.28.0", labeler.Image)
		require.Empty(t, labeler.Ports, dir)
		require.Equal(t, "always", labeler.Restart)
		require.Equal(t, "3m0s", labeler.StopGracePeriod.String())
		require.Equal(t, "1", *labeler.Environment["LABELER_MAX_IN_FLIGHT_TESTS"])
		require.Equal(t, "shared", *labeler.Environment["LABELER_SERVICE_MODE"])
		require.Equal(t, "/app/configs/remote-gpt-6-sol.json", *labeler.Environment["LABELER_CONFIG_FILE"])
		require.Equal(t, "http://inference:8080", *labeler.Environment["FIFTYONE_INFERENCE_URL"])
		require.Equal(t, "test-0123456789abcdef0123456789abcdef", *labeler.Environment["FIFTYONE_INFERENCE_TOKEN"])
		require.Equal(t, "shared-token", *labeler.Environment["LABELER_AUTH_MODE"])
		require.NotContains(t, labeler.Environment, "FIFTYONE_INFERENCE_AUTH")
		require.NotContains(t, labeler.Environment, "LABELER_ENGINE_CONFIG")

		// Runs: a CPU delegated operator worker on the lane the App queues to
		runs, err := project.GetService("inference-labeler-runs")
		require.NoError(t, err, dir)
		require.Equal(t, "voxel51/agentic-labeler-cpu:v2.28.0", runs.Image)
		require.Contains(t, strings.Join(runs.Command, " "), "fiftyone delegated launch -t remote -m -n agentic-labeler", dir)
		require.Empty(t, runs.Ports, dir)
		require.Empty(t, runs.Deploy, dir)
		require.Equal(t, "always", runs.Restart)
		require.Equal(t, "job", *runs.Environment["LABELER_RUN_MODE"])
		require.Equal(t, "none", *runs.Environment["LABELER_INTAKE"])
		require.Equal(t, "env", *runs.Environment["LABELER_TENANCY"])
		require.Equal(t, "/app/configs/remote-gpt-6-sol.json", *runs.Environment["LABELER_CONFIG_FILE"])
		require.Equal(t, "http://inference:8080", *runs.Environment["FIFTYONE_INFERENCE_URL"])
		require.Equal(t, "test-0123456789abcdef0123456789abcdef", *runs.Environment["FIFTYONE_INFERENCE_TOKEN"])
		// Each run names the org and user who started it
		require.NotContains(t, runs.Environment, "FIFTYONE_INFERENCE_TENANT_ID")
		require.Equal(t, "/opt/plugins", *runs.Environment["FIFTYONE_PLUGINS_DIR"])
		// The executor refuses to start without these
		require.Equal(t, "true", *runs.Environment["FIFTYONE_INTERNAL_SERVICE"])
		require.Equal(t, "false", *runs.Environment["FIFTYONE_DATABASE_ADMIN"])
		require.NotEmpty(t, *runs.Environment["FIFTYONE_ENCRYPTION_KEY"], dir)
		require.NotEmpty(t, *runs.Environment["FIFTYONE_DATABASE_URI"], dir)
		require.NotEmpty(t, *runs.Environment["API_URL"], dir)
		// Activity, as on teams-do
		require.Equal(t, "true", *runs.Environment["FIFTYONE_ACTIVITY_ENABLED"])
		require.Equal(t, "redis://fiftyone-mq-redis:6379/0", *runs.Environment["FIFTYONE_MQ_REDIS_URL"])
		require.Contains(t, runs.Environment, "FIFTYONE_ACTIVITY_ORG_ID")
		require.Equal(t, "plugins-vol", runs.Volumes[0].Source)
		require.True(t, runs.Volumes[0].ReadOnly, dir)

		// The service refuses to start without a token
		_, err = inferenceProject(t, dir, nil, nil)
		require.ErrorContains(t, err, "INFERENCE_SHARED_TOKEN is required", dir)
	}
}

// The compose overlay declares the model server with no models: the service
// reads them from it
func TestDockerComposeInferenceModelServer(t *testing.T) {
	t.Parallel()

	for _, dir := range inferenceComposeDirs {
		project, err := inferenceProject(t, dir, inferenceEnv, nil)
		require.NoError(t, err, dir)
		inference, err := project.GetService("inference")
		require.NoError(t, err, dir)
		var backends []map[string]any
		require.NoError(t, yaml.Unmarshal([]byte(*inference.Environment["BACKENDS"]), &backends), dir)

		require.Len(t, backends, 1, dir)
		require.Equal(t, "model-server", backends[0]["id"], dir)
		require.Equal(t, "model-server", backends[0]["kind"], dir)
		require.Equal(t, "http://model-server:8080", backends[0]["baseUrl"], dir)
		// Without a token, an absent auth is none
		require.Nil(t, backends[0]["auth"], dir)
		require.NotContains(t, backends[0], "models", dir)
	}
}

// The model servers' optional token: a bearer on the shipped backend
func TestDockerComposeInferenceModelServerToken(t *testing.T) {
	t.Parallel()

	for _, dir := range inferenceComposeDirs {
		backendsOf := func(env []string) (*types.ServiceConfig, []map[string]any) {
			project, err := inferenceProject(t, dir, env, nil)
			require.NoError(t, err, dir)
			inference, err := project.GetService("inference")
			require.NoError(t, err, dir)
			var backends []map[string]any
			require.NoError(t, yaml.Unmarshal([]byte(*inference.Environment["BACKENDS"]), &backends), dir)
			return &inference, backends
		}

		inference, backends := backendsOf(inferenceEnv)
		require.Equal(t, "", *inference.Environment["MODEL_SERVER_TOKEN"], dir)
		require.Nil(t, backends[0]["auth"], dir)

		inference, backends = backendsOf(append([]string{"MODEL_SERVER_TOKEN=test-model-server-token"}, inferenceEnv...))
		require.Equal(t, "test-model-server-token", *inference.Environment["MODEL_SERVER_TOKEN"], dir)
		require.Equal(t, map[string]any{"mode": "bearer-token", "tokenEnv": "MODEL_SERVER_TOKEN"}, backends[0]["auth"], dir)
	}
}
