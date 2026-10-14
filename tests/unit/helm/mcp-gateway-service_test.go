//go:build kubeall || helm || unit || unitMcpGatewayService
// +build kubeall helm unit unitMcpGatewayService

package unit

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corev1 "k8s.io/api/core/v1"
)

type serviceMcpGatewayTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestServiceMcpGatewayTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &serviceMcpGatewayTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/mcp-gateway-service.yaml"},
	})
}

func (s *serviceMcpGatewayTemplateTest) TestDisabledByDefault() {
	options := &helm.Options{SetValues: nil}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template templates/mcp-gateway-service.yaml in chart")
}

// The ingress's `/mcp` path hardcodes `serviceName: mcp-gateway` in
// `values.yaml` (it can't reference this service by a computed name the
// way in-cluster env vars do). This default has to stay literally
// "mcp-gateway", or the route silently points at nothing.
func (s *serviceMcpGatewayTemplateTest) TestMetadataNameDefaultsToMcpGateway() {
	options := &helm.Options{SetValues: map[string]string{"mcpGatewaySettings.enabled": "true"}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var service corev1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &service)

	s.Equal("mcp-gateway", service.ObjectMeta.Name, "Service name should default to mcp-gateway")
}

func (s *serviceMcpGatewayTemplateTest) TestSelectorMatchesDeploymentPodLabels() {
	options := &helm.Options{SetValues: map[string]string{"mcpGatewaySettings.enabled": "true"}}

	serviceOutput := helm.RenderTemplate(
		s.T(), options, s.chartPath, s.releaseName, []string{"templates/mcp-gateway-service.yaml"},
	)
	var service corev1.Service
	helm.UnmarshalK8SYaml(s.T(), serviceOutput, &service)

	deploymentOutput := helm.RenderTemplate(
		s.T(), options, s.chartPath, s.releaseName, []string{"templates/mcp-gateway-deployment.yaml"},
	)
	var deployment struct {
		Spec struct {
			Template struct {
				Metadata struct {
					Labels map[string]string `yaml:"labels"`
				} `yaml:"metadata"`
			} `yaml:"template"`
		} `yaml:"spec"`
	}
	helm.UnmarshalK8SYaml(s.T(), deploymentOutput, &deployment)

	for key, value := range service.Spec.Selector {
		s.Equal(
			value, deployment.Spec.Template.Metadata.Labels[key],
			"Service selector should match the Deployment's pod labels, or traffic never reaches a pod",
		)
	}
}
