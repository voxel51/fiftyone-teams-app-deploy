//go:build docker || compose || unit || unitComposeLegacyAuth
// +build docker compose unit unitComposeLegacyAuth

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

const (
	dockerLegacyAuthDir = "../../../docker/legacy-auth"
)

var legacyAuthComposeFile = filepath.Join(dockerLegacyAuthDir, "compose.yaml")
var legacyAuthComposePluginsFile = filepath.Join(dockerLegacyAuthDir, "compose.plugins.yaml")
var legacyAuthComposeDedicatedPluginsFile = filepath.Join(dockerLegacyAuthDir, "compose.dedicated-plugins.yaml")
var legacyAuthComposeDelegatedOperationsFile = filepath.Join(dockerInternalAuthDir, "compose.delegated-operators.yaml")
var legacyAuthEnvTemplateFilePath = filepath.Join(dockerLegacyAuthDir, "env.template")

type commonServicesLegacyAuthDockerComposeTest struct {
	suite.Suite
	composeFilePath string
	projectName     string
	dotEnvFiles     []string
}

func TestDockerComposeLegacyAuth(t *testing.T) {
	t.Parallel()

	_, err := filepath.Abs(dockerLegacyAuthDir)
	require.NoError(t, err)

	suite.Run(t, &commonServicesLegacyAuthDockerComposeTest{
		Suite:           suite.Suite{},
		composeFilePath: dockerLegacyAuthDir,
		projectName:     "fiftyone-compose-test",
		dotEnvFiles: []string{
			legacyAuthEnvTemplateFilePath,
			envFixtureFilePath,
		},
	})
}

func (s *commonServicesLegacyAuthDockerComposeTest) TestServicesNames() {
	testCases := []struct {
		name        string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []string
	}{
		{
			"compose",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"fiftyone-app",
				"teams-api",
				"teams-app",
				"teams-cas",
			},
		},
		{
			"composePlugins",
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			[]string{
				"fiftyone-app",
				"teams-api",
				"teams-app",
				"teams-cas",
			},
		},
		{
			"composeDedicatedPlugins",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"fiftyone-app",
				"teams-api",
				"teams-app",
				"teams-cas",
				"teams-plugins",
			},
		},
		{
			"composeDelegatedOperations",
			[]string{legacyAuthComposeDelegatedOperationsFile},
			s.dotEnvFiles,
			[]string{
				"teams-do",
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
				cli.WithWorkingDirectory(dockerLegacyAuthDir),
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
				logger.Log(subT, line)
			}

			s.Equal(testCase.expected, project.ServiceNames(), fmt.Sprintf("%s - Service Names should be equal", testCase.name))
		})
	}
}

func (s *commonServicesLegacyAuthDockerComposeTest) TestServiceImage() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    string
	}{
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-app:v2.10.0",
		},
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-teams-api:v2.10.0",
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-teams-app:v2.10.0",
		},
		{
			"defaultTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-teams-cas:v2.10.0",
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-app:v2.10.0",
		},
		{
			"delegatedOperationsTeamsDo",
			"teams-do",
			[]string{legacyAuthComposeDelegatedOperationsFile},
			s.dotEnvFiles,
			"voxel51/fiftyone-teams-cv-full:v2.10.0",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			projectOptions, err := cli.NewProjectOptions(
				testCase.configPaths,
				cli.WithWorkingDirectory(dockerLegacyAuthDir),
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
				logger.Log(subT, line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Image, fmt.Sprintf("%s - Image should be equal", testCase.name))
		})
	}
}

func (s *commonServicesLegacyAuthDockerComposeTest) TestServiceEnvironment() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []string
	}{
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_DEFAULT_APP_ADDRESS=0.0.0.0",
				"FIFTYONE_DEFAULT_APP_PORT=5151",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_APP_IMAGES=false",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_SIGNED_URL_EXPIRATION=24",
			},
		},
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"CAS_BASE_URL=http://teams-cas:3000/cas/api",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
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
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"APP_USE_HTTPS=true",
				"FIFTYONE_API_URI=https://example-api.fiftyone.ai",
				"FIFTYONE_APP_ALLOW_MEDIA_EXPORT=true",
				"FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION=2.10.0",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_SERVER_ADDRESS=",
				"FIFTYONE_SERVER_PATH_PREFIX=/api/proxy/fiftyone-teams",
				"FIFTYONE_TEAMS_PROXY_URL=http://fiftyone-app:5151",
				"NODE_ENV=production",
				"RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED=false",
				"FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED=true",
				"FIFTYONE_APP_DEFAULT_QUERY_PERFORMANCE=true",
				"FIFTYONE_APP_ENABLE_QUERY_PERFORMANCE=true",
			},
		},
		{
			"defaultTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"CAS_DATABASE_NAME=fiftyone-cas",
				"CAS_DEFAULT_USER_ROLE=GUEST",
				"CAS_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"CAS_URL=https://example.fiftyone.ai",
				"DEBUG=cas:*,-cas:*:debug",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"LICENSE_KEY_FILE_PATHS=/opt/fiftyone/licenses/license",
				"NEXTAUTH_URL=https://example.fiftyone.ai/cas/api/auth",
				"TEAMS_API_DATABASE_NAME=fiftyone",
				"TEAMS_API_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
			},
		},
		{
			"ffDisableInvitationsFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_DEFAULT_APP_ADDRESS=0.0.0.0",
				"FIFTYONE_DEFAULT_APP_PORT=5151",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_APP_IMAGES=false",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_SIGNED_URL_EXPIRATION=24",
			},
		},
		{
			"ffDisableInvitationsTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"CAS_BASE_URL=http://teams-cas:3000/cas/api",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
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
			"ffDisableInvitationsTeamsApp",
			"teams-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"APP_USE_HTTPS=true",
				"FIFTYONE_API_URI=https://example-api.fiftyone.ai",
				"FIFTYONE_APP_ALLOW_MEDIA_EXPORT=true",
				"FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION=2.10.0",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_SERVER_ADDRESS=",
				"FIFTYONE_SERVER_PATH_PREFIX=/api/proxy/fiftyone-teams",
				"FIFTYONE_TEAMS_PROXY_URL=http://fiftyone-app:5151",
				"NODE_ENV=production",
				"RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED=false",
				"FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED=true",
				"FIFTYONE_APP_DEFAULT_QUERY_PERFORMANCE=true",
				"FIFTYONE_APP_ENABLE_QUERY_PERFORMANCE=true",
			},
		},
		{
			"ffDisableInvitationsTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]string{
				"CAS_DATABASE_NAME=fiftyone-cas",
				"CAS_DEFAULT_USER_ROLE=GUEST",
				"CAS_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"CAS_URL=https://example.fiftyone.ai",
				"DEBUG=cas:*,-cas:*:debug",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"LICENSE_KEY_FILE_PATHS=/opt/fiftyone/licenses/license",
				"NEXTAUTH_URL=https://example.fiftyone.ai/cas/api/auth",
				"TEAMS_API_DATABASE_NAME=fiftyone",
				"TEAMS_API_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
			},
		},
		{
			"pluginsFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
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
				"FIFTYONE_SIGNED_URL_EXPIRATION=24",
			},
		},
		{
			"pluginsTeamsApi",
			"teams-api",
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			[]string{
				"CAS_BASE_URL=http://teams-cas:3000/cas/api",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
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
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"APP_USE_HTTPS=true",
				"FIFTYONE_API_URI=https://example-api.fiftyone.ai",
				"FIFTYONE_APP_ALLOW_MEDIA_EXPORT=true",
				"FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION=2.10.0",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_SERVER_ADDRESS=",
				"FIFTYONE_SERVER_PATH_PREFIX=/api/proxy/fiftyone-teams",
				"FIFTYONE_TEAMS_PROXY_URL=http://fiftyone-app:5151",
				"NODE_ENV=production",
				"RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED=false",
				"FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED=true",
				"FIFTYONE_APP_DEFAULT_QUERY_PERFORMANCE=true",
				"FIFTYONE_APP_ENABLE_QUERY_PERFORMANCE=true",
			},
		},
		{
			"pluginsTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			[]string{
				"CAS_DATABASE_NAME=fiftyone-cas",
				"CAS_DEFAULT_USER_ROLE=GUEST",
				"CAS_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"CAS_URL=https://example.fiftyone.ai",
				"DEBUG=cas:*,-cas:*:debug",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"LICENSE_KEY_FILE_PATHS=/opt/fiftyone/licenses/license",
				"NEXTAUTH_URL=https://example.fiftyone.ai/cas/api/auth",
				"TEAMS_API_DATABASE_NAME=fiftyone",
				"TEAMS_API_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
			},
		},
		{
			"dedicatedPluginsFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_DEFAULT_APP_ADDRESS=0.0.0.0",
				"FIFTYONE_DEFAULT_APP_PORT=5151",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_APP_IMAGES=false",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_SIGNED_URL_EXPIRATION=24",
			},
		},
		{
			"dedicatedPluginsTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"CAS_BASE_URL=http://teams-cas:3000/cas/api",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
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
			"dedicatedPluginsTeamsApp",
			"teams-app",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"APP_USE_HTTPS=true",
				"FIFTYONE_API_URI=https://example-api.fiftyone.ai",
				"FIFTYONE_APP_ALLOW_MEDIA_EXPORT=true",
				"FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION=2.10.0",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_SERVER_ADDRESS=",
				"FIFTYONE_SERVER_PATH_PREFIX=/api/proxy/fiftyone-teams",
				"FIFTYONE_TEAMS_PROXY_URL=http://fiftyone-app:5151",
				"NODE_ENV=production",
				"RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED=false",
				"FIFTYONE_TEAMS_PLUGIN_URL=http://teams-plugins:5151",
				"FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED=true",
				"FIFTYONE_APP_DEFAULT_QUERY_PERFORMANCE=true",
				"FIFTYONE_APP_ENABLE_QUERY_PERFORMANCE=true",
			},
		},
		{
			"dedicatedPluginsTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			[]string{
				"CAS_DATABASE_NAME=fiftyone-cas",
				"CAS_DEFAULT_USER_ROLE=GUEST",
				"CAS_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"CAS_URL=https://example.fiftyone.ai",
				"DEBUG=cas:*,-cas:*:debug",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"LICENSE_KEY_FILE_PATHS=/opt/fiftyone/licenses/license",
				"NEXTAUTH_URL=https://example.fiftyone.ai/cas/api/auth",
				"TEAMS_API_DATABASE_NAME=fiftyone",
				"TEAMS_API_MONGODB_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
			},
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_AUTH_SECRET=test-fiftyone-auth-secret",
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
			},
		},
		{
			"delegatedOperationsTeamsDo",
			"teams-do",
			[]string{legacyAuthComposeDelegatedOperationsFile},
			s.dotEnvFiles,
			[]string{
				"API_URL=http://teams-api:8000",
				"FIFTYONE_DATABASE_ADMIN=false",
				"FIFTYONE_DATABASE_NAME=fiftyone",
				"FIFTYONE_DATABASE_URI=mongodb://root:test-secret@mongodb.local/?authSource=admin",
				"FIFTYONE_ENCRYPTION_KEY=test-fiftyone-encryption-key",
				"FIFTYONE_INTERNAL_SERVICE=true",
				"FIFTYONE_MEDIA_CACHE_SIZE_BYTES=-1",
				"FIFTYONE_PLUGINS_DIR=/opt/plugins",
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
				cli.WithWorkingDirectory(dockerLegacyAuthDir),
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
				logger.Log(subT, line)
			}

			if testCase.expected == nil {
				s.Nil(project.Services[testCase.serviceName].Environment, fmt.Sprintf("%s - Environment should be equal", testCase.name))
			} else {
				s.Equal(types.NewMappingWithEquals(testCase.expected), project.Services[testCase.serviceName].Environment, fmt.Sprintf("%s - Environment should be equal", testCase.name))
			}

		})
	}
}

func (s *commonServicesLegacyAuthDockerComposeTest) TestServicePorts() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []types.ServicePortConfig
	}{
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeFile},
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
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeFile},
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
			[]string{legacyAuthComposeFile},
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
			"defaultTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]types.ServicePortConfig{
				{
					Mode:       "ingress",
					HostIP:     "127.0.0.1",
					Target:     3000,
					Published:  "3030",
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
				cli.WithWorkingDirectory(dockerLegacyAuthDir),
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
				logger.Log(subT, line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Ports, fmt.Sprintf("%s - Ports should be equal", testCase.name))
		})
	}
}

func (s *commonServicesLegacyAuthDockerComposeTest) TestServiceRestart() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    string
	}{
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
		{
			"defaultTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			"always",
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			types.RestartPolicyAlways,
		},
		{
			"delegatedOperationsTeamsDo",
			"teams-do",
			[]string{legacyAuthComposeDelegatedOperationsFile},
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
				cli.WithWorkingDirectory(dockerLegacyAuthDir),
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
				logger.Log(subT, line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Restart, fmt.Sprintf("%s - Restart should be equal", testCase.name))
		})
	}
}

func (s *commonServicesLegacyAuthDockerComposeTest) TestServiceVolumes() {
	testCases := []struct {
		name        string
		serviceName string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    []types.ServiceVolumeConfig
	}{
		{
			"defaultFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"defaultTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"defaultTeamsApp",
			"teams-app",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"defaultTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			[]types.ServiceVolumeConfig{
				{
					Type:        "bind",
					Source:      "/some/directory/with/licenses/",
					Target:      "/opt/fiftyone/licenses",
					ReadOnly:    true,
					Consistency: "",
				},
			},
		},
		{
			"pluginsFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposePluginsFile},
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
			"pluginsTeamsApi",
			"teams-api",
			[]string{legacyAuthComposePluginsFile},
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
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"pluginsTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			[]types.ServiceVolumeConfig{
				{
					Type:        "bind",
					Source:      "/some/directory/with/licenses/",
					Target:      "/opt/fiftyone/licenses",
					ReadOnly:    true,
					Consistency: "",
				},
			},
		},
		{
			"dedicatedPluginsFiftyoneApp",
			"fiftyone-app",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			// []types.ServiceVolumeConfig{},
			nil,
		},
		{
			"dedicatedPluginsTeamsApi",
			"teams-api",
			[]string{legacyAuthComposeDedicatedPluginsFile},
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
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"dedicatedPluginsTeamsCas",
			"teams-cas",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			[]types.ServiceVolumeConfig{
				{
					Type:        "bind",
					Source:      "/some/directory/with/licenses/",
					Target:      "/opt/fiftyone/licenses",
					ReadOnly:    true,
					Consistency: "",
				},
			},
		},
		{
			"dedicatedPluginsTeamsPlugins",
			"teams-plugins",
			[]string{legacyAuthComposeDedicatedPluginsFile},
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
			"delegatedOperationsTeamsDo",
			"teams-do",
			[]string{legacyAuthComposeDelegatedOperationsFile},
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
				cli.WithWorkingDirectory(dockerLegacyAuthDir),
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
				logger.Log(subT, line)
			}

			s.Equal(testCase.expected, project.Services[testCase.serviceName].Volumes, fmt.Sprintf("%s - Service Volumes should be equal", testCase.name))
		})
	}
}
func (s *commonServicesLegacyAuthDockerComposeTest) TestVolumes() {
	testCases := []struct {
		name        string
		configPaths []string // file paths to one or more Compose files.
		envFiles    []string // file paths to ".env" files with additional environment variable data
		expected    types.Volumes
	}{
		{
			"default",
			[]string{legacyAuthComposeFile},
			s.dotEnvFiles,
			nil,
		},
		{
			"plugins",
			[]string{legacyAuthComposePluginsFile},
			s.dotEnvFiles,
			types.Volumes{
				"plugins-vol": {
					Name: "fiftyone-compose-test_plugins-vol",
				},
			},
		},
		{
			"dedicatedPlugins",
			[]string{legacyAuthComposeDedicatedPluginsFile},
			s.dotEnvFiles,
			types.Volumes{
				"plugins-vol": {
					Name: "fiftyone-compose-test_plugins-vol",
				},
			},
		},
		{
			"delegatedOperations",
			[]string{legacyAuthComposeDelegatedOperationsFile},
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
				cli.WithWorkingDirectory(dockerLegacyAuthDir),
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
				logger.Log(subT, line)
			}

			s.Equal(testCase.expected, project.Volumes, fmt.Sprintf("%s - Volumes should be equal", testCase.name))
		})
	}
}
