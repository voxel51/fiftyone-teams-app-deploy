//go:build docker || compose || integration || integrationComposeLegacyAuth
// +build docker compose integration integrationComposeLegacyAuth

package integration

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"path/filepath"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/compose-spec/compose-go/v2/dotenv"
)

const (
	dockerLegacyAuthDir = "../../../docker/legacy-auth"
)

var legacyAuthComposeFile = "compose.yaml"
var legacyAuthComposePluginsFile = "compose.plugins.yaml"
var legacyAuthComposeDedicatedPluginsFile = "compose.dedicated-plugins.yaml"
var legacyAuthEnvTemplateFilePath = filepath.Join(dockerLegacyAuthDir, "env.template")

type commonServicesLegacyAuthDockerComposeUpTest struct {
	suite.Suite
	composeFilePath string
	dotEnvFiles     []string
	overrideFiles   []string
}

func TestDockerComposeUpLegacyAuth(t *testing.T) {
	t.Parallel()

	_, err := filepath.Abs(dockerLegacyAuthDir)
	require.NoError(t, err)

	// Set the override files used for the tests
	overrideFiles := []string{
		overrideFile,
		mongodbComposeFile,
	}

	// To run the containers on macOS arm64, we need to set the platform
	if runtime.GOOS == "darwin" {
		overrideFiles = append(overrideFiles, darwinOverrideFile)
	}

	suite.Run(t, &commonServicesLegacyAuthDockerComposeUpTest{
		Suite:           suite.Suite{},
		composeFilePath: dockerLegacyAuthDir,
		dotEnvFiles: []string{
			legacyAuthEnvTemplateFilePath,
			legacyAuthEnvFixtureFilePath,
		},
		overrideFiles: overrideFiles,
	})
}

func (s *commonServicesLegacyAuthDockerComposeUpTest) TestDockerComposeUp() {
	testCases := []struct {
		name          string
		composeFile   string
		overrideFiles []string
		envFiles      []string // file paths to ".env" files with additional environment variable data
		expected      []serviceValidations
	}{
		{
			"compose",
			legacyAuthComposeFile,
			s.overrideFiles,
			s.dotEnvFiles,
			[]serviceValidations{
				{
					name:             "teams-api",
					url:              "http://127.0.0.1:8000/health",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              "[INFO] Starting worker",
				},
				{
					name:             "teams-app",
					url:              "http://127.0.0.1:3000/api/hello",
					responsePayload:  `{"name":"John Doe"}`,
					httpResponseCode: 200,
					log:              "Listening on port 3000",
				},
				{
					name:             "teams-cas",
					url:              "http://127.0.0.1:3030/cas/api",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              " ✓ Ready in",
				},
				// ordering this last to avoid test flakes where testing for log before the container is running
				{
					name:             "fiftyone-app",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 200,
					log:              "[INFO] Running on http://0.0.0.0:5151",
				},
			},
		},
		{
			"composePlugins",
			legacyAuthComposePluginsFile,
			s.overrideFiles,
			s.dotEnvFiles,
			[]serviceValidations{
				{
					name:             "teams-api",
					url:              "http://127.0.0.1:8000/health",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              "[INFO] Starting worker",
				},
				{
					name:             "teams-app",
					url:              "http://127.0.0.1:3000/api/hello",
					responsePayload:  `{"name":"John Doe"}`,
					httpResponseCode: 200,
					log:              "Listening on port 3000",
				},
				{
					name:             "teams-cas",
					url:              "http://127.0.0.1:3030/cas/api",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              " ✓ Ready in",
				},
				// ordering this last to avoid test flakes where testing for log before the container is running
				{
					name:             "fiftyone-app",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 200,
					log:              "[INFO] Running on http://0.0.0.0:5151",
				},
			},
		},
		{
			"composeDedicatedPlugins",
			legacyAuthComposeDedicatedPluginsFile,
			s.overrideFiles,
			s.dotEnvFiles,
			[]serviceValidations{
				{
					name:             "teams-api",
					url:              "http://127.0.0.1:8000/health",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              "[INFO] Starting worker",
				},
				{
					name:             "teams-app",
					url:              "http://127.0.0.1:3000/api/hello",
					responsePayload:  `{"name":"John Doe"}`,
					httpResponseCode: 200,
					log:              "Listening on port 3000",
				},
				{
					name:             "teams-cas",
					url:              "http://127.0.0.1:3030/cas/api",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              " ✓ Ready in",
				},
				// ordering this last to avoid test flakes where testing for log before the container is running
				{
					name:             "fiftyone-app",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 0,
					log:              "[INFO] Running on http://0.0.0.0:5151",
				},
				{
					name:             "teams-plugins",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 0,
					log:              "[INFO] Running on http://0.0.0.0:5151", // same as fiftyone-app since plugins uses or is based on the fiftyone-app image
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			// TODO: If we need parallel, dynamically set mongoDB port and configure the env vars with the custom port
			// For now, we cannot perform concurrent runs because mongodb will error on port already in use
			// subT.Parallel()

			// TODO: Should we use `--env-file` instead of the library `dotenv.Read`
			// to more closely align to real world usage?
			// Something like
			//
			// ```shell
			// docker compose \
			//   -f tests/fixtures/docker/compose.override.mongodb.yaml \
			//   --env-file tests/fixtures/docker/integration_legacy_auth.env

			//   up -d
			// ```
			//
			// Use existing function to get map[string]string of environment variables from .env file(s)
			environmentVariables, err := dotenv.Read(s.dotEnvFiles...)
			s.NoError(err)

			dockerOptions := &docker.Options{
				ProjectName: "fiftyone-" + strings.ToLower(random.UniqueId()),
				WorkingDir:  dockerLegacyAuthDir,
				EnvVars:     environmentVariables,
			}

			// In golang, we cannot mix strings with string slice unpacking.
			// Let's create a slice that will later be unpacked and used as an argument
			// to the variadic function `docker.RunDockerCompose` parameter `args`.
			argsUp := []string{}
			argsDown := []string{}
			args := []string{"-f", testCase.composeFile}

			for _, overrideFile := range s.overrideFiles {
				args = append(args, "-f", overrideFile)
			}

			argsUp = append(args, "up", "--detach")
			argsDown = append(args, "down", "--remove-orphans", "--timeout", "2")

			// Run containers
			output := docker.RunDockerCompose(
				subT,
				dockerOptions,
				argsUp...,
			)
			// Delete containers after tests complete
			defer docker.RunDockerCompose(
				subT,
				dockerOptions,
				argsDown...,
			)

			// Validate system health
			for _, expected := range testCase.expected {
				logger.Log(subT, fmt.Sprintf("Validating service %s...", expected.name))
				s.Contains(output, fmt.Sprintf("Container %s-%s-1  Started", dockerOptions.ProjectName, expected.name), fmt.Sprintf("%s - %s - docker compose output should contain service container started", testCase.name, expected.name))
				if expected.url != "" {
					validate_endpoint(subT, expected.url, expected.responsePayload, expected.httpResponseCode)
				}
				s.Contains(get_logs(subT, dockerOptions, expected.name), expected.log, fmt.Sprintf("%s - %s - log should contain matching entry", testCase.name, expected.name))
			}
		})
	}
}
