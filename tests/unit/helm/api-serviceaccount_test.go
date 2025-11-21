//go:build kubeall || helm || unit || unitServiceAccount
// +build kubeall helm unit unitServiceAccount

package unit

import (
	// "encoding/json"
	"fmt"
	"path/filepath"

	// "reflect"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corev1 "k8s.io/api/core/v1"
)

type apiServiceAccountTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestAPIServiceAccountTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &apiServiceAccountTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/api-serviceaccount.yaml"},
	})
}

func (s *apiServiceAccountTemplateTest) TestDisabled() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			fmt.Sprintf("%s-fiftyone-teams-app-do-management", s.releaseName),
		},
		{
			"overrideServiceAccountName",
			map[string]string{
				"delegatedOperatorJobTemplates.rbac.create": "false",
			},
			"",
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
				s.ErrorContains(err, "could not find template templates/api-serviceaccount.yaml in chart")

				var serviceAccount corev1.ServiceAccount
				helm.UnmarshalK8SYaml(subT, output, &serviceAccount)

				s.Empty(serviceAccount.ObjectMeta.Name, "Name should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var serviceAccount corev1.ServiceAccount
				helm.UnmarshalK8SYaml(subT, output, &serviceAccount)

				s.Equal(testCase.expected, serviceAccount.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

func (s *apiServiceAccountTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			fmt.Sprintf("%s-fiftyone-teams-app-do-management", s.releaseName),
		},
		{
			"overrideName",
			map[string]string{
				"delegatedOperatorJobTemplates.rbac.serviceAccount.name": "test-service-account-name",
			},
			"test-service-account-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var serviceAccount corev1.ServiceAccount
			helm.UnmarshalK8SYaml(subT, output, &serviceAccount)

			s.Equal(testCase.expected, serviceAccount.ObjectMeta.Name, "Name name should be equal.")
		})
	}
}

func (s *apiServiceAccountTemplateTest) TestMetadataNamespace() {
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
			"overrideName",
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

			var serviceAccount corev1.ServiceAccount
			helm.UnmarshalK8SYaml(subT, output, &serviceAccount)

			s.Equal(testCase.expected, serviceAccount.ObjectMeta.Namespace, "Namespace name should be equal.")
		})
	}
}

func (s *apiServiceAccountTemplateTest) TestMetadataAnnotations() {
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
				"delegatedOperatorJobTemplates.rbac.serviceAccount.annotations.annotation-1": "annotation-1-value",
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

			var serviceAccount corev1.ServiceAccount
			helm.UnmarshalK8SYaml(subT, output, &serviceAccount)

			if testCase.expected == nil {
				s.Nil(serviceAccount.ObjectMeta.Annotations, "Annotations should be nil")
			} else {
				for key, value := range testCase.expected {
					foundValue := serviceAccount.ObjectMeta.Annotations[key]
					s.Equal(value, foundValue, "Annotations should contain all set annotations.")
				}
			}
		})
	}
}

func (s *apiServiceAccountTemplateTest) TestMetadataLabels() {
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
				"app.kubernetes.io/name":       fmt.Sprintf("%s-fiftyone-teams-app-do-management", s.releaseName),
				"app.kubernetes.io/instance":   "fiftyone-test",
				"app.voxel51.com/component":    "on-demand-delegated-operators",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
				"delegatedOperatorJobTemplates.rbac.serviceAccount.name":         "test-service-account-name",
				"delegatedOperatorJobTemplates.rbac.serviceAccount.labels.color": "blue",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "test-service-account-name",
				"app.kubernetes.io/instance":   "fiftyone-test",
				"app.voxel51.com/component":    "on-demand-delegated-operators",
				"color":                        "blue",
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

			var serviceAccount corev1.ServiceAccount
			helm.UnmarshalK8SYaml(subT, output, &serviceAccount)

			for key, value := range testCase.expected {
				foundValue := serviceAccount.ObjectMeta.Labels[key]
				s.Equal(value, foundValue, "Labels should contain all set labels.")
			}
		})
	}
}

func (s *apiServiceAccountTemplateTest) TestAutomountServiceAccountToken() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected bool
	}{
		{
			"defaultValues",
			nil,
			true,
		},
		{
			"overrideAutomount",
			map[string]string{
				"delegatedOperatorJobTemplates.rbac.serviceAccount.automount": "false",
			},
			true, // We currently don't allow for the overriding of AutomountServiceAccountToken. `initContainers` use them to lookup namespaces.
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var serviceAccount corev1.ServiceAccount
			helm.UnmarshalK8SYaml(subT, output, &serviceAccount)

			s.Equal(testCase.expected, *serviceAccount.AutomountServiceAccountToken, "Automount Service Account Token should be equal.")
		})
	}
}
