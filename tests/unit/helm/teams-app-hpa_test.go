//go:build kubeall || helm || unit || unitHpa || unitTeamsAppHpa
// +build kubeall helm unit unitHpa unitTeamsAppHpa

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

type horizontalPodAutoscalerTeamsAppTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestHorizontalPodAutoscalerTeamsAppTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &horizontalPodAutoscalerTeamsAppTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/teams-app-hpa.yaml"},
	})
}

func (s *horizontalPodAutoscalerTeamsAppTemplateTest) TestMetadataLabels() {
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
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "fiftyone-teams-app",
				"app.kubernetes.io/instance":   "fiftyone-test",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
				// Unlike teams-api, fiftyone-app, and teams-plugins, setting `teamsAppSettings.service.name`
				// does not affect the label `app.kubernetes.io/name` for teams-app.
				// See note in _helpers.tpl.
				"teamsAppSettings.service.name": "test-service-name",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "fiftyone-teams-app",
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
				s.ErrorContains(err, "could not find template templates/teams-app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerTeamsAppTemplateTest) TestMetadataName() {
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
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
			},
			"teams-app",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
				"teamsAppSettings.service.name":        "test-service-name",
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
				s.ErrorContains(err, "could not find template templates/teams-app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerTeamsAppTemplateTest) TestMetadataNamespace() {
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
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
			},
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
				"namespace.name":                       "test-namespace-name",
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
				s.ErrorContains(err, "could not find template templates/teams-app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerTeamsAppTemplateTest) TestScaleTargetRef() {
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
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
			},
			func(ref autoscalingv2.CrossVersionObjectReference) {
				expectedRefJSON := `{
          "apiVersion": "apps/v1",
          "kind": "Deployment",
          "name": "teams-app"
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
				"teamsAppSettings.autoscaling.enabled": "true",
				"teamsAppSettings.service.name":        "test-service-name",
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
				s.ErrorContains(err, "could not find template templates/teams-app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerTeamsAppTemplateTest) TestMaxReplicas() {
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
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
			},
			5,
		},
		{
			"overrideMaxReplicas",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled":     "true",
				"teamsAppSettings.autoscaling.maxReplicas": "9",
			},
			9,
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
				s.ErrorContains(err, "could not find template templates/teams-app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerTeamsAppTemplateTest) TestMinReplicas() {
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
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
			},
			2,
		},
		{
			"overrideMinReplicas",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled":     "true",
				"teamsAppSettings.autoscaling.minReplicas": "3",
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
				s.ErrorContains(err, "could not find template templates/teams-app-hpa.yaml in chart")

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
func (s *horizontalPodAutoscalerTeamsAppTemplateTest) TestMetrics() {
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
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"teamsAppSettings.autoscaling.enabled": "true",
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
				"teamsAppSettings.autoscaling.enabled":                           "true",
				"teamsAppSettings.autoscaling.targetCPUUtilizationPercentage":    "99",
				"teamsAppSettings.autoscaling.targetMemoryUtilizationPercentage": "98",
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
				s.ErrorContains(err, "could not find template templates/teams-app-hpa.yaml in chart")

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
