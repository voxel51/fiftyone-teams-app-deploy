//go:build kubeall || helm || unit || unitPluginsService
// +build kubeall helm unit unitPluginsService

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

	corev1 "k8s.io/api/core/v1"
)

type servicePluginsTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestServicePluginsTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &servicePluginsTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/plugins-service.yaml"},
	})
}

func (s *servicePluginsTemplateTest) TestMetadataAnnotations() {
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
			"overrideAnnotations",
			map[string]string{
				"pluginsSettings.enabled":                          "true",
				"pluginsSettings.service.annotations.annotation-1": "annotation-1-value",
			},
			map[string]string{
				"annotation-1": "annotation-1-value",
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
				s.ErrorContains(err, "could not find template templates/plugins-service.yaml in chart")

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Nil(service.ObjectMeta.Annotations, "Annotations should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				for key, value := range testCase.expected {
					foundValue := service.ObjectMeta.Annotations[key]
					s.Equal(value, foundValue, "Annotations should contain all set annotations.")
				}
			}
		})
	}
}

func (s *servicePluginsTemplateTest) TestMetadataLabels() {
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
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.name": "test-service-name",
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

			if testCase.values == nil {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-service.yaml in chart")

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Nil(service.ObjectMeta.Labels, "Labels should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				for key, value := range testCase.expected {
					foundValue := service.ObjectMeta.Labels[key]
					s.Equal(value, foundValue, "Labels should contain all set annotations.")
				}
			}
		})
	}
}

func (s *servicePluginsTemplateTest) TestMetadataName() {
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
			"teams-plugins",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.name": "test-service-name",
			},
			"test-service-name",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.name": "test-service-name",
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
				s.ErrorContains(err, "could not find template templates/plugins-service.yaml in chart")

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Empty(service.ObjectMeta.Name, "Name should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Equal(testCase.expected, service.ObjectMeta.Name, "Name name should be equal.")
			}
		})
	}
}

func (s *servicePluginsTemplateTest) TestMetadataNamespace() {
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
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"pluginsSettings.enabled": "true",
				"namespace.name":          "test-namespace-name",
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
				s.ErrorContains(err, "could not find template templates/plugins-service.yaml in chart")

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Empty(service.ObjectMeta.Namespace, "Namespace should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Equal(testCase.expected, service.ObjectMeta.Namespace, "Namespace name should be equal.")
			}
		})
	}
}

func (s *servicePluginsTemplateTest) TestPorts() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(port []corev1.ServicePort)
	}{
		{
			"defaultValues",
			nil,
			func(ports []corev1.ServicePort) {
				expectedPortsJSON := `[]`
				var expectedPorts []corev1.ServicePort
				err := json.Unmarshal([]byte(expectedPortsJSON), &expectedPorts)
				s.NoError(err)
				s.Equal(expectedPorts, ports, "Ports should be equal")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(ports []corev1.ServicePort) {
				expectedPortsJSON := `[
          {
            "port": 80,
            "targetPort": 5151,
            "protocol": "TCP",
            "name": "http"
          }
        ]`
				var expectedPorts []corev1.ServicePort
				err := json.Unmarshal([]byte(expectedPortsJSON), &expectedPorts)
				s.NoError(err)
				s.Equal(expectedPorts, ports, "Ports should be equal")
			},
		},
		{
			"overrideNodePortWithoutPortNumber",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.type": "NodePort",
			},
			func(ports []corev1.ServicePort) {
				expectedPortsJSON := `[
          {
            "port": 80,
            "targetPort": 5151,
            "protocol": "TCP",
            "name": "http"
          }
        ]`
				var expectedPorts []corev1.ServicePort
				err := json.Unmarshal([]byte(expectedPortsJSON), &expectedPorts)
				s.NoError(err)
				s.Equal(expectedPorts, ports, "Ports should be equal")
			},
		},
		{
			"overrideNodePortWithPortNumber",
			map[string]string{
				"pluginsSettings.enabled":          "true",
				"pluginsSettings.service.type":     "NodePort",
				"pluginsSettings.service.nodePort": "9999",
			},
			func(ports []corev1.ServicePort) {
				expectedPortsJSON := `[
          {
            "port": 80,
            "targetPort": 5151,
            "protocol": "TCP",
            "name": "http",
            "nodePort": 9999
          }
        ]`
				var expectedPorts []corev1.ServicePort
				err := json.Unmarshal([]byte(expectedPortsJSON), &expectedPorts)
				s.NoError(err)
				s.Equal(expectedPorts, ports, "Ports should be equal")
			},
		},
		{
			"overrideServiceServicePortValues",
			map[string]string{
				"pluginsSettings.enabled":               "true",
				"pluginsSettings.service.containerPort": "8001",
				"pluginsSettings.service.port":          "88",
			},
			func(ports []corev1.ServicePort) {
				expectedPortsJSON := `[
          {
            "port": 88,
            "targetPort": 8001,
            "protocol": "TCP",
            "name": "http"
          }
        ]`
				var expectedPorts []corev1.ServicePort
				err := json.Unmarshal([]byte(expectedPortsJSON), &expectedPorts)
				s.NoError(err)
				s.Equal(expectedPorts, ports, "Ports should be equal")
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
				s.ErrorContains(err, "could not find template templates/plugins-service.yaml in chart")

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Nil(service.Spec.Ports, "Ports should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				testCase.expected(service.Spec.Ports)
			}
		})
	}
}

func (s *servicePluginsTemplateTest) TestType() {
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
			"ClusterIP",
		},
		{
			"overrideSelectorLabels",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.type": "NodePort",
			},
			"NodePort",
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
				s.ErrorContains(err, "could not find template templates/plugins-service.yaml in chart")

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Empty(service.ObjectMeta.Name, "Type should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Equal(testCase.expected, string(service.Spec.Type), "Type should be equal.")
			}
		})
	}
}

func (s *servicePluginsTemplateTest) TestSelectorLabels() {
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
			map[string]string{
				"app.kubernetes.io/name":     "teams-plugins",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
		},
		{
			"overrideSelectorLabels",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.name": "test-service-name",
			},
			map[string]string{
				"app.kubernetes.io/name":     "test-service-name",
				"app.kubernetes.io/instance": "fiftyone-test",
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
				s.ErrorContains(err, "could not find template templates/plugins-service.yaml in chart")

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				s.Nil(service.Spec.Selector, "Selector labels should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var service corev1.Service
				helm.UnmarshalK8SYaml(subT, output, &service)

				for key, value := range testCase.expected {
					foundValue := service.Spec.Selector[key]
					s.Equal(value, foundValue, "Selector labels should contain all set labels.")
				}
			}
		})
	}
}
