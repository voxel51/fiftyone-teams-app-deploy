//go:build kubeall || helm || unit || unitHpa || unitHpaPlugins
// +build kubeall helm unit unitHpa unitHpaPlugins

package unit

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type horizontalPodAutoscalerPluginsTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestHorizontalPodAutoscalerPluginsTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &horizontalPodAutoscalerPluginsTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/plugins-hpa.yaml"},
	})
}

func (s *horizontalPodAutoscalerPluginsTemplateTest) TestMetadataLabels() {
	// Get chart info (to later obtain the chart's appVersion)
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)

	// Get appVersion from chart info
	chartAppVersion, exists := cInfo["appVersion"]
	s.True(exists, "failed to get app version from chart info")

	// Get version from chart info
	chartVersion, exists := cInfo["version"]
	s.True(exists, "failed to get version from chart info")

	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			nil,
		},
		{
			"defaultValuesPluginsHpaEnabled",
			map[string]string{
				"pluginsSettings.autoscaling.enabled": "true",
				"pluginsSettings.enabled":             "true",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "teams-plugins",
				"app.kubernetes.io/instance":   "fiftyone-test",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
				"pluginsSettings.autoscaling.enabled": "true",
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.service.name":        "test-service-name",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "test-service-name",
				"app.kubernetes.io/instance":   "fiftyone-test",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == nil {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-hpa.yaml in chart")

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Nil(hpa.ObjectMeta.Labels, "Labels should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				for key, value := range testCase.expected {
					foundValue := hpa.ObjectMeta.Labels[key]
					s.Equal(value, foundValue, "Labels should contain all set labels.")
				}
			}
		})
	}
}

func (s *horizontalPodAutoscalerPluginsTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"",
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			"",
		},
		{
			"defaultValuesPluginsHpaEnabled",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
			},
			"teams-plugins",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
				"pluginsSettings.service.name":        "test-service-name",
			},
			"test-service-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == "" {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-hpa.yaml in chart")

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Empty(hpa.ObjectMeta.Name, "Metadata name should be nil")

			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Equal(testCase.expected, hpa.ObjectMeta.Name, "Metadata name should be equal.")
			}
		})
	}
}

func (s *horizontalPodAutoscalerPluginsTemplateTest) TestMetadataNamespace() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"",
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			"",
		},
		{
			"defaultValuesPluginsHpaEnabled",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
			},
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
				"namespace.name":                      "test-namespace-name",
			},
			"test-namespace-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == "" {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-hpa.yaml in chart")

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Empty(hpa.ObjectMeta.Namespace, "Metadata namespace should be nil")

			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Equal(testCase.expected, hpa.ObjectMeta.Namespace, "Namespace name should be equal.")
			}
		})
	}
}

func (s *horizontalPodAutoscalerPluginsTemplateTest) TestScaleTargetRef() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(ref autoscalingv2.CrossVersionObjectReference)
	}{
		{
			"defaultValues",
			nil,
			func(ref autoscalingv2.CrossVersionObjectReference) {
				expectedRefJSON := `{}`
				var expectedRef autoscalingv2.CrossVersionObjectReference
				err := json.Unmarshal([]byte(expectedRefJSON), &expectedRef)
				s.NoError(err)
				s.Equal(expectedRef, ref, "Scale Target Refs should be equal")
			},
		},
		{
			"defaultValuesPluginsHpaEnabled",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
			},
			func(ref autoscalingv2.CrossVersionObjectReference) {
				expectedRefJSON := `{
          "apiVersion": "apps/v1",
          "kind": "Deployment",
          "name": "teams-plugins"
        }`
				var expectedRef autoscalingv2.CrossVersionObjectReference
				err := json.Unmarshal([]byte(expectedRefJSON), &expectedRef)
				s.NoError(err)
				s.Equal(expectedRef, ref, "Scale Target Refs should be equal")
			},
		},
		{
			"overrideServiceName",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
				"pluginsSettings.service.name":        "test-service-name",
			},
			func(ref autoscalingv2.CrossVersionObjectReference) {
				expectedRefJSON := `{
          "apiVersion": "apps/v1",
          "kind": "Deployment",
          "name": "test-service-name"
        }`
				var expectedRef autoscalingv2.CrossVersionObjectReference
				err := json.Unmarshal([]byte(expectedRefJSON), &expectedRef)
				s.NoError(err)
				s.Equal(expectedRef, ref, "Scale Target Refs should be equal")
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.values == nil {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-hpa.yaml in chart")

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Empty(hpa.Spec.ScaleTargetRef, "Scale TargetRef should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				testCase.expected(hpa.Spec.ScaleTargetRef)
			}
		})
	}
}

func (s *horizontalPodAutoscalerPluginsTemplateTest) TestMaxReplicas() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int32
	}{
		{
			"defaultValues",
			nil,
			0,
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			0,
		},
		{
			"defaultValuesPluginsHpaEnabled",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
			},
			20,
		},
		{
			"overrideMaxReplicas",
			map[string]string{
				"pluginsSettings.enabled":                 "true",
				"pluginsSettings.autoscaling.enabled":     "true",
				"pluginsSettings.autoscaling.maxReplicas": "19",
			},
			19,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == 0 {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-hpa.yaml in chart")

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Empty(hpa.Spec.MaxReplicas, "maxReplicas should be empty")

			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Equal(testCase.expected, hpa.Spec.MaxReplicas, "maxReplicas name should be equal.")
			}
		})
	}
}

func (s *horizontalPodAutoscalerPluginsTemplateTest) TestMinReplicas() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int32
	}{
		{
			"defaultValues",
			nil,
			0,
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			0,
		},
		{
			"defaultValuesPluginsHpaEnabled",
			map[string]string{
				"pluginsSettings.enabled":             "true",
				"pluginsSettings.autoscaling.enabled": "true",
			},
			2,
		},
		{
			"overrideMinReplicas",
			map[string]string{
				"pluginsSettings.enabled":                 "true",
				"pluginsSettings.autoscaling.enabled":     "true",
				"pluginsSettings.autoscaling.minReplicas": "3",
			},
			3,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == 0 {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-hpa.yaml in chart")

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Empty(hpa.Spec.MinReplicas, "minReplicas should be empty")

			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Equal(testCase.expected, *hpa.Spec.MinReplicas, "minReplicas name should be equal.")
			}
		})
	}
}
func (s *horizontalPodAutoscalerPluginsTemplateTest) TestMetrics() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(metrics []autoscalingv2.MetricSpec)
	}{
		{
			"defaultValues",
			nil,
			func(metrics []autoscalingv2.MetricSpec) {
				s.Empty(metrics, "metricSpec should not be set")
			},
		},
		{
			"defaultValuesPluginsHpaEnabled",
			map[string]string{
				"pluginsSettings.autoscaling.enabled": "true",
				"pluginsSettings.enabled":             "true",
			},
			func(metrics []autoscalingv2.MetricSpec) {
				expectedJSON := `[
          {
            "type": "Resource",
            "resource": {
              "name": "cpu",
              "target": {
                "type": "Utilization",
                "averageUtilization": 80
              }
            }
          },
          {
            "type": "Resource",
            "resource": {
              "name": "memory",
              "target": {
                "type": "Utilization",
                "averageUtilization": 80
              }
            }
          }
        ]`
				var expectedMetrics []autoscalingv2.MetricSpec
				err := json.Unmarshal([]byte(expectedJSON), &expectedMetrics)
				s.NoError(err)
				s.Equal(expectedMetrics, metrics, "Volumes should be equal")
			},
		},
		{
			"overrideTargetCpuAndMemory",
			map[string]string{
				"pluginsSettings.autoscaling.enabled":                           "true",
				"pluginsSettings.autoscaling.targetCPUUtilizationPercentage":    "99",
				"pluginsSettings.autoscaling.targetMemoryUtilizationPercentage": "98",
				"pluginsSettings.enabled":                                       "true",
			},
			func(metrics []autoscalingv2.MetricSpec) {
				expectedJSON := `[
          {
            "type": "Resource",
            "resource": {
              "name": "cpu",
              "target": {
                "type": "Utilization",
                "averageUtilization": 99
              }
            }
          },
          {
            "type": "Resource",
            "resource": {
              "name": "memory",
              "target": {
                "type": "Utilization",
                "averageUtilization": 98
              }
            }
          }
        ]`
				var expectedMetrics []autoscalingv2.MetricSpec
				err := json.Unmarshal([]byte(expectedJSON), &expectedMetrics)
				s.NoError(err)
				s.Equal(expectedMetrics, metrics, "Volumes should be equal")
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.values == nil {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-hpa.yaml in chart")

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				s.Empty(hpa.Spec.Metrics, "metrics should be empty")

			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var hpa autoscalingv2.HorizontalPodAutoscaler
				helm.UnmarshalK8SYaml(subT, output, &hpa)

				testCase.expected(hpa.Spec.Metrics)
			}
		})
	}
}
