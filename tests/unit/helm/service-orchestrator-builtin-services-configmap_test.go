//go:build kubeall || helm || unit || unitServiceOrchestrator
// +build kubeall helm unit unitServiceOrchestrator

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
	"gopkg.in/yaml.v3"

	corev1 "k8s.io/api/core/v1"
)

type builtinServicesConfigMapTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestBuiltinServicesConfigMapTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &builtinServicesConfigMapTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates: []string{
			"templates/service-orchestrator-builtin-services-configmap.yaml",
		},
	})
}

func (s *builtinServicesConfigMapTemplateTest) TestGatingAndName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string // empty => the template must not render
	}{
		{
			"defaultValues",
			nil,
			"",
		},
		{
			"enabledButCreateDisabled",
			map[string]string{
				"serviceOrchestrator.enabled":                          "true",
				"serviceOrchestrator.builtinServices.configMap.create": "false",
			},
			"",
		},
		{
			"enabled",
			map[string]string{
				"serviceOrchestrator.enabled": "true",
			},
			fmt.Sprintf("%s-fiftyone-teams-app-builtin-services", s.releaseName),
		},
		{
			"enabledWithNameOverride",
			map[string]string{
				"serviceOrchestrator.enabled":                        "true",
				"serviceOrchestrator.builtinServices.configMap.name": "custom-builtin-services",
			},
			"custom-builtin-services",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

			if testCase.expected == "" {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/service-orchestrator-builtin-services-configmap.yaml in chart")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var configMap corev1.ConfigMap
				helm.UnmarshalK8SYaml(subT, output, &configMap)

				s.Equal(testCase.expected, configMap.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

func (s *builtinServicesConfigMapTemplateTest) TestServicesPayload() {
	testCases := []struct {
		name          string
		values        map[string]string
		jsonValues    map[string]string
		expectedCount int
		expected      func(services []map[string]interface{})
	}{
		{
			// Empty by default: reconcile deep-merges nothing and the
			// packaged defaults apply unchanged.
			"enabledDefaultsToEmptyList",
			map[string]string{
				"serviceOrchestrator.enabled": "true",
			},
			nil,
			0,
			nil,
		},
		{
			"enabledWithServiceOverrides",
			map[string]string{
				"serviceOrchestrator.enabled": "true",
			},
			map[string]string{
				"serviceOrchestrator.builtinServices.services": `[
          {
            "id": "builtin:image-segmentation-ai",
            "builtin_version": 2,
            "delegation_target": "teams-do"
          },
          {
            "id": "builtin:video-propagation-ai",
            "builtin_version": 2,
            "entrypoint": {
              "container": {
                "image": "registry:5000/sam2-video:tag",
                "command": ["python", "-m", "fiftyone.annotation.endpoints.sam2_video"]
              }
            }
          }
        ]`,
			},
			2,
			func(services []map[string]interface{}) {
				s.Equal("builtin:image-segmentation-ai", services[0]["id"])
				s.Equal("teams-do", services[0]["delegation_target"])
				s.Equal("builtin:video-propagation-ai", services[1]["id"])
				entrypoint := services[1]["entrypoint"].(map[string]interface{})
				container := entrypoint["container"].(map[string]interface{})
				s.Equal("registry:5000/sam2-video:tag", container["image"])
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{
				SetValues:     disableTelemetry(testCase.values),
				SetJsonValues: testCase.jsonValues,
			}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(subT, output, &configMap)

			payload, ok := configMap.Data["builtin_services.yaml"]
			s.True(ok, "ConfigMap should carry builtin_services.yaml")

			var services []map[string]interface{}
			err := yaml.Unmarshal([]byte(payload), &services)
			s.NoError(err, "Payload should parse as a YAML list")
			s.Len(services, testCase.expectedCount)

			if testCase.expected != nil {
				testCase.expected(services)
			}
		})
	}
}
