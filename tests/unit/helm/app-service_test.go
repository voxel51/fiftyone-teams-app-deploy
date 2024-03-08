//go:build kubeall || helm || unit || unitAppService
// +build kubeall helm unit unitAppService

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

type serviceAppTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestServiceAppTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &serviceAppTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/app-service.yaml"},
	})
}

func (s *serviceAppTemplateTest) TestMetadataAnnotations() {
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
			"overrideAnnotations",
			map[string]string{
				"appSettings.service.annotations.annotation-1": "annotation-1-value",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var service corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &service)

			if testCase.expected == nil {
				s.Nil(service.ObjectMeta.Annotations, "Annotations should be nil")
			} else {
				for key, value := range testCase.expected {
					foundValue := service.ObjectMeta.Annotations[key]
					s.Equal(value, foundValue, "Annotations should contain all set annotations.")
				}
			}
		})
	}
}

func (s *serviceAppTemplateTest) TestMetadataLabels() {
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
				"appSettings.service.name": "test-service-name",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var service corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &service)

			for key, value := range testCase.expected {
				foundValue := service.ObjectMeta.Labels[key]
				s.Equal(value, foundValue, "Labels should contain all set labels.")
			}
		})
	}
}

func (s *serviceAppTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-app",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"appSettings.service.name": "test-service-name",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var service corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &service)

			s.Equal(testCase.expected, service.ObjectMeta.Name, "Name should be equal.")
		})
	}
}

func (s *serviceAppTemplateTest) TestMetadataNamespace() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"namespace.name": "test-namespace-name",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var service corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &service)

			s.Equal(testCase.expected, service.ObjectMeta.Namespace, "Namespace name should be equal.")
		})
	}
}

func (s *serviceAppTemplateTest) TestPorts() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(port []corev1.ServicePort)
	}{
		{
			"defaultValues",
			nil,
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
				"appSettings.service.type": "NodePort",
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
				"appSettings.service.type":     "NodePort",
				"appSettings.service.nodePort": "9999",
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
				"appSettings.service.containerPort": "8001",
				"appSettings.service.port":          "88",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var service corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &service)

			testCase.expected(service.Spec.Ports)
		})
	}
}

func (s *serviceAppTemplateTest) TestType() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"ClusterIP",
		},
		{
			"overrideSelectorLabels",
			map[string]string{
				"appSettings.service.type": "NodePort",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var service corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &service)

			s.Equal(testCase.expected, string(service.Spec.Type), "Type should be equal.")
		})
	}
}

func (s *serviceAppTemplateTest) TestSelectorLabels() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			nil,
			map[string]string{
				"app.kubernetes.io/name":     "fiftyone-app",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
		},
		{
			"overrideSelectorLabels",
			map[string]string{
				"appSettings.service.name": "test-service-name",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var service corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &service)

			for key, value := range testCase.expected {
				foundValue := service.Spec.Selector[key]
				s.Equal(value, foundValue, "Selector labels should contain all set labels.")
			}
		})
	}
}
