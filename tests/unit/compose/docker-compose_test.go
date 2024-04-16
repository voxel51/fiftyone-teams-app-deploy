//go:build docker || compose || unit || unitComposeCommonServices
// +build docker compose unit unitComposeCommonServices

package unit

import (
	"context"
	"fmt"
	"strings"

	"path/filepath"

	"testing"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TODO: Move to `common_test.go`?
const (
	dockerDir                   = "../../../docker/"
	composeFile                 = "../../../docker/compose.yaml"
	composePluginsFile          = "../../../docker/compose.plugins.yaml"
	composeDedicatedPluginsFile = "../../../docker/compose.dedicated-plugins.yaml"
	envTemplateFilePath         = "../../../docker/env.template"
	envFixtureFilePath          = "../../fixtures/docker/.env"
)

type commonServicesDockerComposeTest struct {
	suite.Suite
	composeFilePath string
	projectName     string
	dotEnvFiles     []string
}

func TestDockerCompose(t *testing.T) {
	t.Parallel()

	_, err := filepath.Abs(dockerDir)
	require.NoError(t, err)

	suite.Run(t, &commonServicesDockerComposeTest{
		Suite:           suite.Suite{},
		composeFilePath: dockerDir,
		projectName:     "fiftyone-compose-test",
		dotEnvFiles: []string{
			filepath.Join(dockerDir, "env.template"),
			"../../fixtures/docker/.env",
		},
	})
}

func (s *commonServicesDockerComposeTest) TestServicesNames() {
	testCases := []struct {
		name        string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []string
	}{
		{
			"compose",
			[]string{composeFile},
			s.dotEnvFiles,
			[]string{
				"fiftyone-app",
				"teams-api",
				"teams-app",
			},
		},
		{
			"composePlugins",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			[]string{
				"fiftyone-app",
				"teams-api",
				"teams-app",
			},
		},
		{
			"composeDedicatedPlugins",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"fiftyone-app",
				"teams-api",
				"teams-app",
				"teams-plugins",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerDir),
				cli.WithName(s.projectName),
				cli.WithEnvFiles(testCase.envFiles...),
				cli.WithDotEnv,
			)
			s.NoError(err)

			project, err := cli.ProjectFromOptions(context.TODO(), projectOptions)
			s.NoError(err)

			projectYAML, err := project.MarshalYAML()
			s.NoError(err)
			// The only next line only prints timestamp on the first line of the yaml file
			// logger.Log(s.T(), string(projectYAML))
			for _, line := range strings.Split(string(projectYAML), "\n") {
				logger.Log(s.T(), line)
			}

			s.Equal(testCase.expected, project.ServiceNames(), "Service Names should be equal")
		})
	}
}

func (s *commonServicesDockerComposeTest) TestServiceImage() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    string
	}{
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{composeFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-teams-api:v1.5.8",
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{composeFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-teams-app:v1.5.8",
		},
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{composeFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-app:v1.5.8",
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-app:v1.5.8",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerDir),
				cli.WithName(s.projectName),
				cli.WithEnvFiles(testCase.envFiles...),
				cli.WithDotEnv,
			)
			s.NoError(err)

			project, err := cli.ProjectFromOptions(context.TODO(), projectOptions)
			s.NoError(err)

			// Log Output
			projectYAML, err := project.MarshalYAML()
			s.NoError(err)
			// The only next line only prints timestamp on the first line of the yaml file
			// logger.Log(s.T(), string(projectYAML))
			for _, line := range strings.Split(string(projectYAML), "\n") {
				logger.Log(s.T(), line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Image, fmt.Sprintf("%s - Image should be equal", testCase.name))
		})
	}
}

func (s *commonServicesDockerComposeTest) TestServiceEnvironment() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []string
	}{
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{composeFile},
			s.dotEnvFiles,
			[]string{
				"AUTH0_API_CLIENT_ID=test-auth0-api-client-id",
				"AUTH0_API_CLIENT_SECRET=test-auth0-api-client-secret",
				"AUTH0_AUDIENCE=test-auth0-audience",
				"AUTH0_CLIENT_ID=test-auth0-client-id",
				"AUTH0_DOMAIN=test-auth0-domain",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_ENV=production",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"GRAPHQL_DEFAULT_LIMIT=10",
				"LOGGING_LEVEL=INFO",
				"MONGO_DEFAULT_DB=fiftyone",
			},
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{composeFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"APP_USE_HTTPS=true",
				"AUTH0_AUDIENCE=test-auth0-audience",
				"AUTH0_BASE_URL=https://example.fiftyone.ai",
				"AUTH0_CLIENT_ID=test-auth0-client-id",
				"AUTH0_CLIENT_SECRET=test-auth0-client-secret",
				"AUTH0_ISSUER_BASE_URL=test-auth0-issuer-base-url",
				"AUTH0_ORGANIZATION=test-auth0-organization",
				"AUTH0_SECRET=test-auth0-secret",
				"FIFTYONE_API_URI=https://example-api.fiftyone.ai",
				"FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION=0.15.9",
				"FIFTYONE_SERVER_ADDRESS=",
				"FIFTYONE_SERVER_PATH_PREFIX=/api/proxy/fiftyone-teams",
				"FIFTYONE_TEAMS_PROXY_URL=http://fiftyone-app:5151",
				"NODE_ENV=production",
				"RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED=false",
			},
		},
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{composeFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_DEFAULT_APP_ADDRESS=0.0.0.0",
				"FIFTYONE_DEFAULT_APP_PORT=5151",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_APP_IMAGES=false",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_TEAMS_AUDIENCE=test-auth0-audience",
				"FIFTYONE_TEAMS_CLIENT_ID=test-auth0-client-id",
				"FIFTYONE_TEAMS_DOMAIN=test-auth0-domain",
				"FIFTYONE_TEAMS_ORGANIZATION=test-auth0-organization",
			},
		},
		{
			"pluginsTeamsApi",
			"teams-api",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			[]string{
				"AUTH0_API_CLIENT_ID=test-auth0-api-client-id",
				"AUTH0_API_CLIENT_SECRET=test-auth0-api-client-secret",
				"AUTH0_AUDIENCE=test-auth0-audience",
				"AUTH0_CLIENT_ID=test-auth0-client-id",
				"AUTH0_DOMAIN=test-auth0-domain",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_ENV=production",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_PLUGINS_DIR=/opt/plugins",
				"GRAPHQL_DEFAULT_LIMIT=10",
				"LOGGING_LEVEL=INFO",
				"MONGO_DEFAULT_DB=fiftyone",
			},
		},
		{
			"pluginsTeamsApp",
			"teams-app",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"APP_USE_HTTPS=true",
				"AUTH0_AUDIENCE=test-auth0-audience",
				"AUTH0_BASE_URL=https://example.fiftyone.ai",
				"AUTH0_CLIENT_ID=test-auth0-client-id",
				"AUTH0_CLIENT_SECRET=test-auth0-client-secret",
				"AUTH0_ISSUER_BASE_URL=test-auth0-issuer-base-url",
				"AUTH0_ORGANIZATION=test-auth0-organization",
				"AUTH0_SECRET=test-auth0-secret",
				"FIFTYONE_API_URI=https://example-api.fiftyone.ai",
				"FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION=0.15.9",
				"FIFTYONE_SERVER_ADDRESS=",
				"FIFTYONE_SERVER_PATH_PREFIX=/api/proxy/fiftyone-teams",
				"FIFTYONE_TEAMS_PROXY_URL=http://fiftyone-app:5151",
				"NODE_ENV=production",
				"RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED=false",
			},
		},
		{
			"pluginsFiftyoneApp",
			"fiftyone-app",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_DEFAULT_APP_ADDRESS=0.0.0.0",
				"FIFTYONE_DEFAULT_APP_PORT=5151",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_APP_IMAGES=false",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_PLUGINS_CACHE_ENABLED=true",
				"FIFTYONE_PLUGINS_DIR=/opt/plugins",
				"FIFTYONE_TEAMS_AUDIENCE=test-auth0-audience",
				"FIFTYONE_TEAMS_CLIENT_ID=test-auth0-client-id",
				"FIFTYONE_TEAMS_DOMAIN=test-auth0-domain",
				"FIFTYONE_TEAMS_ORGANIZATION=test-auth0-organization",
			},
		},
		{
			"dedicatedPluginsTeamsApi",
			"teams-api",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"AUTH0_API_CLIENT_ID=test-auth0-api-client-id",
				"AUTH0_API_CLIENT_SECRET=test-auth0-api-client-secret",
				"AUTH0_AUDIENCE=test-auth0-audience",
				"AUTH0_CLIENT_ID=test-auth0-client-id",
				"AUTH0_DOMAIN=test-auth0-domain",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_ENV=production",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"GRAPHQL_DEFAULT_LIMIT=10",
				"LOGGING_LEVEL=INFO",
				"MONGO_DEFAULT_DB=fiftyone",
				"FIFTYONE_PLUGINS_DIR=/opt/plugins",
			},
		},
		{
			"dedicatedPluginsTeamsApp",
			"teams-app",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"APP_USE_HTTPS=true",
				"AUTH0_AUDIENCE=test-auth0-audience",
				"AUTH0_BASE_URL=https://example.fiftyone.ai",
				"AUTH0_CLIENT_ID=test-auth0-client-id",
				"AUTH0_CLIENT_SECRET=test-auth0-client-secret",
				"AUTH0_ISSUER_BASE_URL=test-auth0-issuer-base-url",
				"AUTH0_ORGANIZATION=test-auth0-organization",
				"AUTH0_SECRET=test-auth0-secret",
				"FIFTYONE_API_URI=https://example-api.fiftyone.ai",
				"FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION=0.15.9",
				"FIFTYONE_SERVER_ADDRESS=",
				"FIFTYONE_SERVER_PATH_PREFIX=/api/proxy/fiftyone-teams",
				"FIFTYONE_TEAMS_PLUGIN_URL=http://teams-plugins:5151",
				"FIFTYONE_TEAMS_PROXY_URL=http://fiftyone-app:5151",
				"NODE_ENV=production",
				"RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED=false",
			},
		},
		{
			"dedicatedPluginsFiftyoneApp",
			"fiftyone-app",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_DEFAULT_APP_ADDRESS=0.0.0.0",
				"FIFTYONE_DEFAULT_APP_PORT=5151",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_APP_IMAGES=false",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_TEAMS_AUDIENCE=test-auth0-audience",
				"FIFTYONE_TEAMS_CLIENT_ID=test-auth0-client-id",
				"FIFTYONE_TEAMS_DOMAIN=test-auth0-domain",
				"FIFTYONE_TEAMS_ORGANIZATION=test-auth0-organization",
			},
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"FIFTYONE_PLUGINS_CACHE_ENABLED=true",
				"API_URL=http://teams-api:8000",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_DEFAULT_APP_ADDRESS=0.0.0.0",
				"FIFTYONE_DEFAULT_APP_PORT=5151",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_APP_IMAGES=false",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_PLUGINS_DIR=/opt/plugins",
				"FIFTYONE_TEAMS_AUDIENCE=test-auth0-audience",
				"FIFTYONE_TEAMS_CLIENT_ID=test-auth0-client-id",
				"FIFTYONE_TEAMS_DOMAIN=test-auth0-domain",
				"FIFTYONE_TEAMS_ORGANIZATION=test-auth0-organization",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerDir),
				cli.WithName(s.projectName),
				cli.WithEnvFiles(testCase.envFiles...),
				cli.WithDotEnv,
			)
			s.NoError(err)

			project, err := cli.ProjectFromOptions(context.TODO(), projectOptions)
			s.NoError(err)

			// Log Output
			projectYAML, err := project.MarshalYAML()
			s.NoError(err)
			// The only next line only prints timestamp on the first line of the yaml file
			// logger.Log(s.T(), string(projectYAML))
			for _, line := range strings.Split(string(projectYAML), "\n") {
				logger.Log(s.T(), line)
			}

			s.Equal(types.NewMappingWithEquals(testCase.expected), project.Services[testCase.serviceName].Environment, fmt.Sprintf("%s - Environment should be equal", testCase.name))
		})
	}
}

func (s *commonServicesDockerComposeTest) TestServicePorts() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []types.ServicePortConfig
	}{
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{composeFile},
			s.dotEnvFiles,
			[]types.ServicePortConfig{
				{
					Mode:       "ingress",
					HostIP:     "127.0.0.1",
					Target:     8000,
					Published:  "8000",
					Protocol:   "tcp",
					Extensions: nil,
				},
			},
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{composeFile},
			s.dotEnvFiles,
			[]types.ServicePortConfig{
				{
					Mode:       "ingress",
					HostIP:     "127.0.0.1",
					Target:     3000,
					Published:  "3000",
					Protocol:   "tcp",
					Extensions: nil,
				},
			},
		},
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{composeFile},
			s.dotEnvFiles,
			[]types.ServicePortConfig{
				{
					Mode:       "ingress",
					HostIP:     "127.0.0.1",
					Target:     5151,
					Published:  "5151",
					Protocol:   "tcp",
					Extensions: nil,
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerDir),
				cli.WithName(s.projectName),
				cli.WithEnvFiles(testCase.envFiles...),
				cli.WithDotEnv,
			)
			s.NoError(err)

			project, err := cli.ProjectFromOptions(context.TODO(), projectOptions)
			s.NoError(err)

			// Log Output
			projectYAML, err := project.MarshalYAML()
			s.NoError(err)
			// The only next line only prints timestamp on the first line of the yaml file
			// logger.Log(s.T(), string(projectYAML))
			for _, line := range strings.Split(string(projectYAML), "\n") {
				logger.Log(s.T(), line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Ports, fmt.Sprintf("%s - Ports should be equal", testCase.name))
		})
	}
}

func (s *commonServicesDockerComposeTest) TestServiceRestart() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    string
	}{
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{composeFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{composeFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{composeFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerDir),
				cli.WithName(s.projectName),
				cli.WithEnvFiles(testCase.envFiles...),
				cli.WithDotEnv,
			)
			s.NoError(err)

			project, err := cli.ProjectFromOptions(context.TODO(), projectOptions)
			s.NoError(err)

			// Log Output
			projectYAML, err := project.MarshalYAML()
			s.NoError(err)
			// The only next line only prints timestamp on the first line of the yaml file
			// logger.Log(s.T(), string(projectYAML))
			for _, line := range strings.Split(string(projectYAML), "\n") {
				logger.Log(s.T(), line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Restart, fmt.Sprintf("%s - Restart should be equal", testCase.name))
		})
	}
}

func (s *commonServicesDockerComposeTest) TestServiceVolumes() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []types.ServiceVolumeConfig
	}{
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{composeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{composeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{composeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"pluginsTeamsApi",
			"teams-api",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			[]types.ServiceVolumeConfig{
				{
					Type:     "volume",
					Source:   "plugins-vol",
					Target:   "/opt/plugins",
					ReadOnly: false,
					Volume:   &types.ServiceVolumeVolume{},
				},
			},
		},
		{
			"pluginsTeamsApp",
			"teams-app",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"pluginsFiftyoneApp",
			"fiftyone-app",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			[]types.ServiceVolumeConfig{
				{
					Type:     "volume",
					Source:   "plugins-vol",
					Target:   "/opt/plugins",
					ReadOnly: true,
					Volume:   &types.ServiceVolumeVolume{},
				},
			},
		},
		{
			"dedicatedPluginsTeamsApi",
			"teams-api",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]types.ServiceVolumeConfig{
				{
					Type:     "volume",
					Source:   "plugins-vol",
					Target:   "/opt/plugins",
					ReadOnly: false,
					Volume:   &types.ServiceVolumeVolume{},
				},
			},
		},
		{
			"dedicatedPluginsTeamsApp",
			"teams-app",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"dedicatedPluginsFiftyoneApp",
			"fiftyone-app",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			// []types.ServiceVolumeConfig{},
			nil,
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]types.ServiceVolumeConfig{
				{
					Type:     "volume",
					Source:   "plugins-vol",
					Target:   "/opt/plugins",
					ReadOnly: true,
					Volume:   &types.ServiceVolumeVolume{},
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerDir),
				cli.WithName(s.projectName),
				cli.WithEnvFiles(testCase.envFiles...),
				cli.WithDotEnv,
			)
			s.NoError(err)

			project, err := cli.ProjectFromOptions(context.TODO(), projectOptions)
			s.NoError(err)

			// Log Output
			projectYAML, err := project.MarshalYAML()
			s.NoError(err)
			// The only next line only prints timestamp on the first line of the yaml file
			// logger.Log(s.T(), string(projectYAML))
			for _, line := range strings.Split(string(projectYAML), "\n") {
				logger.Log(s.T(), line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Volumes, fmt.Sprintf("%s - Service Volumes should be equal", testCase.name))
		})
	}
}
func (s *commonServicesDockerComposeTest) TestVolumes() {
	testCases := []struct {
		name        string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    types.Volumes
	}{
		{
			"default",
			[]string{composeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"plugins",
			[]string{composePluginsFile},
			s.dotEnvFiles,
			types.Volumes{
				"plugins-vol": {
					Name: "fiftyone-compose-test_plugins-vol",
				},
			},
		},
		{
			"dedicatedPlugins",
			[]string{composeDedicatedPluginsFile},
			s.dotEnvFiles,
			types.Volumes{
				"plugins-vol": {
					Name: "fiftyone-compose-test_plugins-vol",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerDir),
				cli.WithName(s.projectName),
				cli.WithEnvFiles(testCase.envFiles...),
				cli.WithDotEnv,
			)
			s.NoError(err)

			project, err := cli.ProjectFromOptions(context.TODO(), projectOptions)
			s.NoError(err)

			// Log Output
			projectYAML, err := project.MarshalYAML()
			s.NoError(err)
			// The only next line only prints timestamp on the first line of the yaml file
			// logger.Log(s.T(), string(projectYAML))
			for _, line := range strings.Split(string(projectYAML), "\n") {
				logger.Log(s.T(), line)
			}

			s.Equal(testCase.expected, project.Volumes, fmt.Sprintf("%s - Volumes should be equal", testCase.name))
		})
	}
}
