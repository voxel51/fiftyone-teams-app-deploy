//go:build kubeall || helm || unit || unitAppHpa
// +build kubeall helm unit unitAppHpa

package unit

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type horizontalPodAutoscalerAppTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestHorizontalPodAutoscalerAppTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &horizontalPodAutoscalerAppTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/app-hpa.yaml"},
	})
}

func (s *horizontalPodAutoscalerAppTemplateTest) TestMetadataLabels() {
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
				"appSettings.autoscaling.enabled": "true",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "fiftyone-app",
				"app.kubernetes.io/instance":   "fiftyone-test",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
				"appSettings.autoscaling.enabled": "true",
				"appSettings.service.name":        "test-service-name",
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
				s.ErrorContains(err, "could not find template templates/app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerAppTemplateTest) TestMetadataName() {
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
				"appSettings.autoscaling.enabled": "true",
			},
			"fiftyone-app",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"appSettings.autoscaling.enabled": "true",
				"appSettings.service.name":        "test-service-name",
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
				s.ErrorContains(err, "could not find template templates/app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerAppTemplateTest) TestMetadataNamespace() {
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
				"appSettings.autoscaling.enabled": "true",
			},
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"appSettings.autoscaling.enabled": "true",
				"namespace.name":                  "test-namespace-name",
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
				s.ErrorContains(err, "could not find template templates/app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerAppTemplateTest) TestScaleTargetRef() {
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
				s.True(reflect.DeepEqual(expectedRef, ref), "Scale Target Refs should be equal")
			},
		},
		{
			"defaultValuesAppHpaEnabled",
			map[string]string{
				"appSettings.autoscaling.enabled": "true",
			},
			func(ref autoscalingv2.CrossVersionObjectReference) {
				expectedRefJSON := `{
          "apiVersion": "apps/v1",
          "kind": "Deployment",
          "name": "fiftyone-app"
        }`
				var expectedRef autoscalingv2.CrossVersionObjectReference
				err := json.Unmarshal([]byte(expectedRefJSON), &expectedRef)
				s.NoError(err)
				s.True(reflect.DeepEqual(expectedRef, ref), "Scale Target Refs should be equal")
			},
		},
		{
			"overrideServiceName",
			map[string]string{
				"appSettings.autoscaling.enabled": "true",
				"appSettings.service.name":        "test-service-name",
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
				s.True(reflect.DeepEqual(expectedRef, ref), "Scale Target Refs should be equal")
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
				s.ErrorContains(err, "could not find template templates/app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerAppTemplateTest) TestMaxReplicas() {
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
				"appSettings.autoscaling.enabled": "true",
			},
			20,
		},
		{
			"overrideMaxReplicas",
			map[string]string{
				"appSettings.autoscaling.enabled":     "true",
				"appSettings.autoscaling.maxReplicas": "19",
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
				s.ErrorContains(err, "could not find template templates/app-hpa.yaml in chart")

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

func (s *horizontalPodAutoscalerAppTemplateTest) TestMinReplicas() {
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
				"appSettings.autoscaling.enabled": "true",
			},
			2,
		},
		{
			"overrideMinReplicas",
			map[string]string{
				"appSettings.autoscaling.enabled":     "true",
				"appSettings.autoscaling.minReplicas": "3",
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
				s.ErrorContains(err, "could not find template templates/app-hpa.yaml in chart")

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
func (s *horizontalPodAutoscalerAppTemplateTest) TestMetrics() {
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
				"appSettings.autoscaling.enabled": "true",
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
				s.True(reflect.DeepEqual(expectedMetrics, metrics), "Volumes should be equal")
			},
		},
		{
			"overrideTargetCpuAndMemory",
			map[string]string{
				"appSettings.autoscaling.enabled":                           "true",
				"appSettings.autoscaling.targetCPUUtilizationPercentage":    "99",
				"appSettings.autoscaling.targetMemoryUtilizationPercentage": "98",
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
				s.True(reflect.DeepEqual(expectedMetrics, metrics), "Volumes should be equal")
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
				s.ErrorContains(err, "could not find template templates/app-hpa.yaml in chart")

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
