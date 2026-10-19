//go:build docker || compose || unit || unitComposeInternalAuth || unitComposeLegacyAuth
// +build docker compose unit unitComposeInternalAuth unitComposeLegacyAuth

package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/require"
)

var modelServerComposeDirs = []string{"../../../docker/internal-auth", "../../../docker/legacy-auth"}

var modelServerImageEnv = "MODEL_SERVER_IMAGE=registry.example/model-server:v1"

// The project with the app's env files, as on the worker that runs
// FiftyOne Enterprise
func modelServerProject(t *testing.T, dir string, files []string, env []string) (*types.Project, error) {
	return modelServerProjectFrom(t, dir, files, env, cli.WithEnvFiles(filepath.Join(dir, "env.template"), envFixtureFilePath), cli.WithDotEnv)
}

// The project from env alone, as on a worker of its own: no app env file
func aloneProject(t *testing.T, dir string, env []string) (*types.Project, error) {
	return modelServerProjectFrom(t, dir, []string{"compose.model-server.yaml"}, env)
}

func modelServerProjectFrom(t *testing.T, dir string, files []string, env []string, extra ...cli.ProjectOptionsFn) (*types.Project, error) {
	paths := []string{}
	for _, file := range files {
		paths = append(paths, filepath.Join(dir, file))
	}
	options, err := cli.NewProjectOptions(
		paths,
		append([]cli.ProjectOptionsFn{
			cli.WithWorkingDirectory(dir),
			cli.WithName("fiftyone-compose-test"),
			cli.WithEnv(env),
		}, extra...)...,
	)
	require.NoError(t, err)
	return cli.ProjectFromOptions(context.TODO(), options)
}

// Both auth modes ship the same model server files
func TestDockerComposeModelServerFilesMatch(t *testing.T) {
	t.Parallel()

	for _, file := range []string{"compose.model-server.yaml", "compose.model-server.gpu.yaml"} {
		internal, err := os.ReadFile(filepath.Join(modelServerComposeDirs[0], file))
		require.NoError(t, err)
		legacy, err := os.ReadFile(filepath.Join(modelServerComposeDirs[1], file))
		require.NoError(t, err)
		require.Equal(t, string(internal), string(legacy), file)
	}
}

func TestDockerComposeModelServer(t *testing.T) {
	t.Parallel()

	for _, dir := range modelServerComposeDirs {
		// Alone, on a worker of its own, with its image the only setting
		project, err := aloneProject(t, dir, []string{modelServerImageEnv})
		require.NoError(t, err, dir)
		require.Equal(t, []string{"model-server"}, project.ServiceNames(), dir)

		service, err := project.GetService("model-server")
		require.NoError(t, err)
		// Nothing it reads is an app variable left empty. An empty token is
		// no token
		for name, value := range service.Environment {
			require.NotNil(t, value, dir+": "+name)
			if name != "MODEL_SERVER_TOKEN" {
				require.NotEmpty(t, *value, dir+": "+name)
			}
		}
		require.Equal(t, "", *service.Environment["MODEL_SERVER_TOKEN"])
		require.Equal(t, "registry.example/model-server:v1", service.Image)
		require.Equal(t, "always", service.Restart)
		require.Equal(t, `["siglip-base-patch16-224-torch"]`, *service.Environment["MODEL_SERVER_PRELOAD"])
		require.Equal(t, "8080", *service.Environment["MODEL_SERVER_PORT"])
		// The server reads an empty list as any, and fails on an empty string
		require.Equal(t, "[]", *service.Environment["MODEL_SERVER_ALLOWED_MODELS"])
		require.Equal(t, "[]", *service.Environment["MODEL_SERVER_ALLOWED_MODEL_KWARGS"])
		require.Equal(t, "[]", *service.Environment["MODEL_SERVER_ALLOWED_MEDIA_PREFIXES"])
		require.Equal(t, "6", *service.Environment["MODEL_SERVER_MAX_LOADED_MODELS"])
		// Weights land on the weights volume, which the server writes
		require.Equal(t, "/models/zoo", *service.Environment["FIFTYONE_MODEL_ZOO_DIR"])
		require.Equal(t, "/models/hf", *service.Environment["HF_HOME"])
		require.Equal(t, "/models/torch", *service.Environment["TORCH_HOME"])
		require.Len(t, service.Volumes, 1)
		require.Equal(t, "model-weights", service.Volumes[0].Source)
		require.Equal(t, "/models", service.Volumes[0].Target)
		require.False(t, service.Volumes[0].ReadOnly)
		require.Contains(t, project.Volumes, "model-weights")
		// CPU: no device reservation
		require.Nil(t, service.Deploy)

		// It checks no token: only this host reaches it by default
		require.Len(t, service.Ports, 1)
		require.Equal(t, "127.0.0.1", service.Ports[0].HostIP)
		require.Equal(t, "8080", service.Ports[0].Published)
		require.Equal(t, uint32(8080), service.Ports[0].Target)

		// A worker of its own binds its address, and may fence what it serves
		project, err = aloneProject(t, dir, []string{
			modelServerImageEnv,
			"MODEL_SERVER_BIND_ADDRESS=10.0.0.5",
			"MODEL_SERVER_BIND_PORT=9000",
			`MODEL_SERVER_ALLOWED_MODELS=["siglip-base-patch16-224-torch"]`,
			`MODEL_SERVER_ALLOWED_MODEL_KWARGS=["classes"]`,
			`MODEL_SERVER_ALLOWED_MEDIA_PREFIXES=["https://minio.internal"]`,
			"MODEL_SERVER_MAX_LOADED_MODELS=2",
			"MODEL_SERVER_TOKEN=test-model-server-token",
		})
		require.NoError(t, err, dir)
		service, err = project.GetService("model-server")
		require.NoError(t, err)
		require.Equal(t, "10.0.0.5", service.Ports[0].HostIP)
		require.Equal(t, "9000", service.Ports[0].Published)
		require.Equal(t, `["siglip-base-patch16-224-torch"]`, *service.Environment["MODEL_SERVER_ALLOWED_MODELS"])
		require.Equal(t, `["classes"]`, *service.Environment["MODEL_SERVER_ALLOWED_MODEL_KWARGS"])
		require.Equal(t, `["https://minio.internal"]`, *service.Environment["MODEL_SERVER_ALLOWED_MEDIA_PREFIXES"])
		require.Equal(t, "2", *service.Environment["MODEL_SERVER_MAX_LOADED_MODELS"])
		require.Equal(t, "test-model-server-token", *service.Environment["MODEL_SERVER_TOKEN"])

		// Beside FiftyOne Enterprise, in the app's -f set
		project, err = modelServerProject(t, dir, []string{"compose.yaml", "compose.model-server.yaml"}, []string{modelServerImageEnv})
		require.NoError(t, err, dir)
		require.Contains(t, project.ServiceNames(), "model-server", dir)
		require.Contains(t, project.ServiceNames(), "fiftyone-app", dir)

		// No image, no start
		_, err = aloneProject(t, dir, nil)
		require.ErrorContains(t, err, "MODEL_SERVER_IMAGE is required", dir)
	}
}

func TestDockerComposeModelServerGpu(t *testing.T) {
	t.Parallel()

	for _, dir := range modelServerComposeDirs {
		project, err := modelServerProject(t, dir, []string{"compose.model-server.yaml", "compose.model-server.gpu.yaml"}, []string{
			modelServerImageEnv,
			"NVIDIA_GPU_COUNT=2",
		})
		require.NoError(t, err, dir)
		service, err := project.GetService("model-server")
		require.NoError(t, err)

		devices := service.Deploy.Resources.Reservations.Devices
		require.Len(t, devices, 1, dir)
		require.Equal(t, "nvidia", devices[0].Driver)
		require.Equal(t, types.DeviceCount(2), devices[0].Count)
		require.Equal(t, []string{"gpu"}, devices[0].Capabilities)
		require.Equal(t, "all", *service.Environment["NVIDIA_VISIBLE_DEVICES"])
		require.Equal(t, "compute,utility", *service.Environment["NVIDIA_DRIVER_CAPABILITIES"])
		// The CPU file's settings stay
		require.Equal(t, "registry.example/model-server:v1", service.Image)
		require.Equal(t, "127.0.0.1", service.Ports[0].HostIP)

		// One GPU unless told otherwise
		project, err = modelServerProject(t, dir, []string{"compose.model-server.yaml", "compose.model-server.gpu.yaml"}, []string{modelServerImageEnv})
		require.NoError(t, err, dir)
		service, err = project.GetService("model-server")
		require.NoError(t, err)
		require.Equal(t, types.DeviceCount(1), service.Deploy.Resources.Reservations.Devices[0].Count, dir)
	}
}
