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

// builtinServices renders the ConfigMap and parses its payload.
func (s *builtinServicesConfigMapTemplateTest) builtinServices(values map[string]string) []map[string]interface{} {
	s.T().Helper()

	options := &helm.Options{SetValues: disableTelemetry(values)}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	payload, ok := configMap.Data["builtin_services.yaml"]
	s.Require().True(ok, "ConfigMap should carry builtin_services.yaml")

	var services []map[string]interface{}
	err := yaml.Unmarshal([]byte(payload), &services)
	s.Require().NoError(err, "Payload should parse as a YAML list")
	return services
}

// TestGating verifies the ConfigMap renders exactly when at least one
// enabled serviceOrchestrators entry carries a service.
func (s *builtinServicesConfigMapTemplateTest) TestGating() {
	testCases := []struct {
		name     string
		values   map[string]string
		rendered bool
	}{
		{
			// The chart's default gpuServiceOrc ships a service
			"defaultValues",
			nil,
			true,
		},
		{
			"noServiceOrchestrators",
			disableDefaultServiceOrchestrators(nil),
			false,
		},
		{
			// Orchestrators without services have nothing to publish
			"orchestratorsWithoutServices",
			disableDefaultServiceOrchestrators(map[string]string{
				"delegatedOperatorJobTemplates.serviceOrchestrators.cpuServices.unused": "nil",
			}),
			false,
		},
		{
			// A disabled orchestrator does not publish its services
			"disabledOrchestrator",
			map[string]string{
				"delegatedOperatorJobTemplates.serviceOrchestrators.cpuServiceOrc":         "null",
				"delegatedOperatorJobTemplates.serviceOrchestrators.gpuServiceOrc.enabled": "false",
			},
			false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

			if testCase.rendered {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var configMap corev1.ConfigMap
				helm.UnmarshalK8SYaml(subT, output, &configMap)

				s.Equal(
					fmt.Sprintf("%s-fiftyone-teams-app-builtin-services", s.releaseName),
					configMap.ObjectMeta.Name,
				)
			} else {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/service-orchestrator-builtin-services-configmap.yaml in chart")
			}
		})
	}
}

// TestDefaultServiceDerivation verifies the chart's default agentic-labeler
// service: identity fields derived from the map keys, `enabled` mapped
// from `autoStart`, the env map flattened to KEY=VALUE lines, and the
// untagged image defaulting to the chart's appVersion.
func (s *builtinServicesConfigMapTemplateTest) TestDefaultServiceDerivation() {
	services := s.builtinServices(nil)
	// Sorted by service key within the orchestrator
	s.Require().Len(services, 2)
	service := services[0]

	// Identity fields derived from the service and orchestrator keys;
	// the explicit label wins over the derived default
	s.Equal("builtin:agentic-labeler", service["id"])
	s.Equal("agentic-labeler", service["kind"])
	s.Equal("agentic-labeler", service["name"])
	s.Equal("Agentic Labeler", service["label"])
	s.Equal("gpuServiceOrc", service["delegation_target"])

	s.Equal(true, service["builtin"])
	s.Equal(1, service["builtin_version"])
	s.Equal("shared", service["scope"])
	s.Equal("", service["secrets"])

	// autoStart maps to the definition's enabled field
	s.Equal(false, service["enabled"])
	_, hasAutoStart := service["autoStart"]
	s.False(hasAutoStart, "autoStart is a values-side key only")

	// The env map flattens to KEY=VALUE lines
	env, ok := service["env"].(string)
	s.Require().True(ok, "env should flatten to a string")
	s.Contains(env, "LABELER_CONFIG_FILE=/app/configs/gemma4-31B-qat-maxvision.json\n")
	s.Contains(env, "LABELER_TENSOR_PARALLEL_SIZE=1\n")
	s.Contains(env, "LABELER_ENFORCE_EAGER=true\n")

	// The untagged image gets the chart's appVersion
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)
	appVersion, exists := cInfo["appVersion"]
	s.True(exists)
	entrypoint := service["entrypoint"].(map[string]interface{})
	container := entrypoint["container"].(map[string]interface{})
	s.Equal(
		fmt.Sprintf("voxel51/agentic-labeler:%s", appVersion),
		container["image"],
	)
	s.Equal("shell", entrypoint["kind"])
	s.Equal(8000, container["port"])

	// The default annotation-ai (SAM2) service
	service = services[1]
	s.Equal("builtin:annotation-ai", service["id"])
	s.Equal("annotation-ai", service["kind"])
	s.Equal("annotation-ai", service["name"])
	s.Equal("Annotation AI", service["label"])
	s.Equal("gpuServiceOrc", service["delegation_target"])
	s.Equal(false, service["enabled"])

	// An omitted env flattens to the empty string
	s.Equal("", service["env"])

	entrypoint = service["entrypoint"].(map[string]interface{})
	container = entrypoint["container"].(map[string]interface{})
	s.Equal(
		fmt.Sprintf("voxel51/fiftyone-teams-cv-full:%s", appVersion),
		container["image"],
	)

	// The per-task container env list passes through untouched
	containerEnv := container["env"].([]interface{})
	s.Require().Len(containerEnv, 1)
	envVar := containerEnv[0].(map[string]interface{})
	s.Equal("LD_LIBRARY_PATH", envVar["name"])
	s.Equal("/usr/local/nvidia/lib64:/usr/local/nvidia/lib", envVar["value"])
}

// TestServiceOverrides verifies explicitly set fields win over the derived
// defaults, string env passes through, digest images are left alone, and
// only a `builtin: true` service gets the builtin: id prefix.
func (s *builtinServicesConfigMapTemplateTest) TestServiceOverrides() {
	services := s.builtinServices(disableDefaultServiceOrchestrators(map[string]string{
		"delegatedOperatorJobTemplates.serviceOrchestrators.customOrc.services.my-service.id":                         "custom:id",
		"delegatedOperatorJobTemplates.serviceOrchestrators.customOrc.services.my-service.builtin_version":            "7",
		"delegatedOperatorJobTemplates.serviceOrchestrators.customOrc.services.my-service.autoStart":                  "true",
		"delegatedOperatorJobTemplates.serviceOrchestrators.customOrc.services.my-service.env":                        "A=b",
		"delegatedOperatorJobTemplates.serviceOrchestrators.customOrc.services.my-service.entrypoint.container.image": "registry:5000/svc@sha256:abc123",
		"delegatedOperatorJobTemplates.serviceOrchestrators.customOrc.services.prefixed-service.builtin":              "true",
	}))
	s.Require().Len(services, 2)
	service := services[0]

	// Explicit fields win over derived defaults
	s.Equal("custom:id", service["id"])
	s.Equal(7, service["builtin_version"])
	s.Equal(true, service["enabled"])

	// Absent fields still get their derived defaults
	s.Equal("my-service", service["kind"])
	s.Equal("my-service", service["name"])
	s.Equal("my-service", service["label"])
	s.Equal("customOrc", service["delegation_target"])
	s.Equal(false, service["builtin"])
	s.Equal("shared", service["scope"])
	s.Equal("", service["secrets"])

	// String env passes through unchanged
	s.Equal("A=b", service["env"])

	// Digest-pinned images are left alone
	entrypoint := service["entrypoint"].(map[string]interface{})
	container := entrypoint["container"].(map[string]interface{})
	s.Equal("registry:5000/svc@sha256:abc123", container["image"])

	// Only a builtin service gets the prefixed id
	service = services[1]
	s.Equal(true, service["builtin"])
	s.Equal("builtin:prefixed-service", service["id"])
}
