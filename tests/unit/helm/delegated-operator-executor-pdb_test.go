//go:build kubeall || helm || unit || unitApiService
// +build kubeall helm unit unitApiService

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

	policyv1 "k8s.io/api/policy/v1"
)

type pdbDelegatedOperatorExecutorTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestPdbDelegatedOperatorExecutorTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &pdbDelegatedOperatorExecutorTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/delegated-operator-executor-pdb.yaml"},
	})
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestMetadataLabels() {
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
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			nil,
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "teams-do",
				"app.kubernetes.io/instance":   "fiftyone-test",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorExecutorSettings.name":                             "test-service-name",
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Nil(pdb.ObjectMeta.Labels, "Labels should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				for key, value := range testCase.expected {
					foundValue := pdb.ObjectMeta.Labels[key]
					s.Equal(value, foundValue, "Labels should contain all set labels.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestMetadataName() {
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
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			"",
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
			},
			"teams-do",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorExecutorSettings.name":                             "test-service-name",
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.ObjectMeta.Name, "Name should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Equal(testCase.expected, pdb.ObjectMeta.Name, "Name should be equal.")
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestMetadataNamespace() {
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
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			"",
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
			},
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
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
			if testCase.expected == "" {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.ObjectMeta.Name, "Namespace should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Equal(testCase.expected, pdb.ObjectMeta.Namespace, "Namespace name should be equal.")
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestMinAvailableInt() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int
	}{
		{
			"defaultValues",
			nil,
			-1,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			-1,
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
			},
			1,
		},
		{
			"overrideMinAvaialble",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "5",
			},
			5,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected < 0 {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MinAvailable, "MinAvailable should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Equal(testCase.expected, pdb.Spec.MinAvailable.IntValue(), "MinAvailable should be equal.")
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestMinAvailablePercent() {
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
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			"",
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "10%",
			},
			"10%",
		},
		{
			"overrideMinAvaialble",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "50%",
			},
			"50%",
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MinAvailable, "MinAvailable should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Equal(testCase.expected, pdb.Spec.MinAvailable.String(), "MinAvailable should be equal.")
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestAxUnavailableInt() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int
	}{
		{
			"defaultValues",
			nil,
			-1,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			-1,
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                            "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.maxUnavailable": "1",
			},
			1,
		},
		{
			"overrideMaxUnavailable",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                            "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.maxUnavailable": "5",
			},
			5,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected < 0 {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Equal(testCase.expected, pdb.Spec.MaxUnavailable.IntValue(), "MaxUnavailable should be equal.")
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestMaxUnavailablePercent() {
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
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			"",
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                            "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.maxUnavailable": "10%",
			},
			"10%",
		},
		{
			"overrideMaxUnavaialble",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                            "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.maxUnavailable": "50%",
			},
			"50%",
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Equal(testCase.expected, pdb.Spec.MaxUnavailable.String(), "MaxUnavailable should be equal.")
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestMinAvaiableAndMaxAvailable() {
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
			"overrideMinAvailableAndMaxUnavailable",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                            "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable":   "10%",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.maxUnavailable": "50%",
			},
			"10%",
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MinAvailable, "MinAvailable should be empty.")
				s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Equal(testCase.expected, pdb.Spec.MinAvailable.String(), "MinAvailable should be equal.")
				s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")
			}
		})
	}
}

func (s *pdbDelegatedOperatorExecutorTemplateTest) TestSelectorLabels() {
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
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled": "true",
			},
			nil,
		},
		{
			"defaultValuesDOPdbEnabled",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
			},
			map[string]string{
				"app.kubernetes.io/name":     "teams-do",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
		},
		{
			"overrideSelectorLabels",
			map[string]string{
				"delegatedOperatorExecutorSettings.enabled":                          "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorExecutorSettings.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorExecutorSettings.name":                             "test-service-name",
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

			if testCase.expected == nil {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/delegated-operator-executor-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.Selector, "Selector Labels should be empty.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				for key, value := range testCase.expected {
					foundValue := pdb.Spec.Selector.MatchLabels[key]
					s.Equal(value, foundValue, "Selector labels should contain all set labels.")
				}
			}
		})
	}
}
