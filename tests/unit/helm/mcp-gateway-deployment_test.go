//go:build kubeall || helm || unit || unitMcpGatewayDeployment
// +build kubeall helm unit unitMcpGatewayDeployment

package unit

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	appsv1 "k8s.io/api/apps/v1"
)

type deploymentMcpGatewayTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestDeploymentMcpGatewayTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &deploymentMcpGatewayTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/mcp-gateway-deployment.yaml"},
	})
}

// Disabled by default is the whole safety story for this chart's existing
// self-serve deployments: no image is published for them yet, so rendering
// nothing unless explicitly opted in is what keeps this change from
// affecting anyone who doesn't ask for it.
func (s *deploymentMcpGatewayTemplateTest) TestDisabledByDefault() {
	options := &helm.Options{SetValues: nil}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template templates/mcp-gateway-deployment.yaml in chart")
}

// The three required URLs are computed, not hand-typed, specifically so a
// deployer can't mistype one relative to settings they already configured
// elsewhere in this chart (teamsAppSettings.dnsName, casSettings.service,
// apiSettings.service).
func (s *deploymentMcpGatewayTemplateTest) TestContainerEnv() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			map[string]string{
				"mcpGatewaySettings.enabled": "true",
			},
			map[string]string{
				"MCP_GATEWAY_RESOURCE_URL":  "",
				"MCP_GATEWAY_CAS_URL":       "http://teams-cas:80/cas/api",
				"MCP_GATEWAY_TEAMS_API_URL": "http://teams-api:80",
				"SERVICE_AUTH_MODE":         "insecure-dev",
			},
		},
		{
			"overrideDnsNameAndServices",
			map[string]string{
				"mcpGatewaySettings.enabled": "true",
				"teamsAppSettings.dnsName":   "teams-app.fiftyone.ai",
				"casSettings.service.name":   "test-teams-cas",
				"casSettings.service.port":   "81",
				"apiSettings.service.name":   "test-teams-api",
				"apiSettings.service.port":   "82",
			},
			map[string]string{
				"MCP_GATEWAY_RESOURCE_URL":  "https://teams-app.fiftyone.ai/mcp",
				"MCP_GATEWAY_CAS_URL":       "http://test-teams-cas:81/cas/api",
				"MCP_GATEWAY_TEAMS_API_URL": "http://test-teams-api:82",
			},
		},
		{
			"overrideServiceAuthMode",
			map[string]string{
				"mcpGatewaySettings.enabled":               "true",
				"mcpGatewaySettings.env.SERVICE_AUTH_MODE": "gcp-id-token",
			},
			map[string]string{
				"SERVICE_AUTH_MODE": "gcp-id-token",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			envByName := map[string]string{}
			for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
				envByName[env.Name] = env.Value
			}

			for key, value := range testCase.expected {
				s.Equal(value, envByName[key], fmt.Sprintf("%s should be equal", key))
			}
		})
	}
}

func (s *deploymentMcpGatewayTemplateTest) TestContainerImage() {
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)

	chartAppVersion, exists := cInfo["appVersion"]
	s.True(exists, "failed to get app version from chart info")

	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			map[string]string{
				"mcpGatewaySettings.enabled": "true",
			},
			fmt.Sprintf("voxel51/fiftyone-mcp-gateway:%s", chartAppVersion),
		},
		{
			"overrideImageTag",
			map[string]string{
				"mcpGatewaySettings.enabled":   "true",
				"mcpGatewaySettings.image.tag": "testTag",
			},
			"voxel51/fiftyone-mcp-gateway:testTag",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			s.Equal(testCase.expected, deployment.Spec.Template.Spec.Containers[0].Image, "Image values should be equal.")
		})
	}
}

// No HTTP health path exists on this server (it serves only `/mcp`, which
// itself requires a valid bearer token), so the probes are TCP checks on
// the named container port rather than the httpGet probes every other
// service in this chart uses.
func (s *deploymentMcpGatewayTemplateTest) TestContainerProbesAreTCP() {
	options := &helm.Options{SetValues: map[string]string{"mcpGatewaySettings.enabled": "true"}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	container := deployment.Spec.Template.Spec.Containers[0]
	s.NotNil(container.LivenessProbe.TCPSocket, "Liveness probe should be a TCP check")
	s.Nil(container.LivenessProbe.HTTPGet, "Liveness probe should not be an HTTP check")
	s.NotNil(container.ReadinessProbe.TCPSocket, "Readiness probe should be a TCP check")
	s.Nil(container.ReadinessProbe.HTTPGet, "Readiness probe should not be an HTTP check")
}
