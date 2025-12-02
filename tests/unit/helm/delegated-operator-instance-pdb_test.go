//go:build kubeall || helm || unit || unitPluginsDeployment
// +build kubeall helm unit unitPluginsDeployment

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

type pdbDelegatedOperatorInstanceTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

type pdbDelegatedOperatorInstanceTemplateLabelsExpected struct {
	selectorMatch    map[string]string
	templateMetadata map[string]string
}

func TestPdbDelegatedOperatorInstanceTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &pdbDelegatedOperatorInstanceTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/delegated-operator-instance-pdb.yaml"},
	})
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMetadataLabels() {
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
		expected []map[string]string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"defaultValuesDOTemplatePdbEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
			},
			[]map[string]string{
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do-cpu-default",
					"app.kubernetes.io/instance":   "fiftyone-test",
				},
			},
		},
		{
			"overrideBaseTemplateEnabledAndInstanceDisabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                      "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled": "false",
			},
			nil,
		},
		{
			"overrideBaseTemplateEnabledAndInstanceEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.labels.color":                     "blue",
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.labels.color":       "red",
			},
			[]map[string]string{
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do-cpu-default",
					"app.kubernetes.io/instance":   "fiftyone-test",
					"color":                        "blue",
				},
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do-two",
					"app.kubernetes.io/instance":   "fiftyone-test",
					"color":                        "red",
				},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Nil(pdb.ObjectMeta.Labels, "Labels should be nil")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(len(testCase.expected[i]), len(pdb.ObjectMeta.Labels), "Label maps should have the same length.")

					for key, value := range testCase.expected[i] {
						foundValue := pdb.ObjectMeta.Labels[key]
						s.Equal(value, foundValue, "Labels should contain all set labels.")
					}
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"defaultValuesDOTemplatePdbEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
			},
			[]string{"teams-do-cpu-default"},
		},
		{
			"overrideInstancePdbEnabledTemplateDisabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                      "false",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":                 "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled": "true",
			},
			[]string{"teams-do-cpu-default"},
		},

		{
			"overrideInstancePdbDisabledTemplateEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                      "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled": "false",
			},
			nil,
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":        "nil",
			},
			nil,
		},
		{
			"overrideTemplatePdbEnabledMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":             "nil",
			},
			[]string{"teams-do-cpu-default", "teams-do-two"},
		},
		{
			"overrideTemplatePdbEnabledInstanceDisabledMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":               "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":          "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":               "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled": "false",
			},
			[]string{"teams-do-cpu-default"},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.ObjectMeta.Name, "Name should be nil")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(testCase.expected[i], pdb.ObjectMeta.Name, "Name should be equal.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMetadataNamespace() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":        "nil",
			},
			nil,
		},
		{
			"defaultValuesDOTemplatePdbEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
			},
			[]string{"fiftyone-teams"},
		},
		{
			"overideBaseTemplateMinNamespace",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":             "nil",
				"namespace.name": "test-namespace-name",
			},
			[]string{"test-namespace-name", "test-namespace-name"},
		},
		{
			"overideBaseTemplateAndInstanceNamespace",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":               "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":          "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":               "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled": "false",
				"namespace.name": "test-namespace-name",
			},
			[]string{"test-namespace-name"},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.ObjectMeta.Namespace, "Namespace should be nil")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(testCase.expected[i], pdb.ObjectMeta.Namespace, "Name should be equal.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMinAvailableInt() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []int
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":        "nil",
			},
			nil,
		},
		{
			"defaultValuesDOTemplatePdbEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
			},
			[]int{1},
		},
		{
			"overrideBaseTemplateAndInstanceMinAvailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                           "false",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":                      "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.minAvailable": "5",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled":             "true",
			},
			[]int{5, 1},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MinAvailable, "MinAvailable should be empty.")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(testCase.expected[i], pdb.Spec.MinAvailable.IntValue(), "MinAvailable should be equal.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMinAvailablePercent() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"overrideBaseTemplateMinAvailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "10%",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
			},
			[]string{"10%"},
		},
		{
			"overrideBaseTemplateAndInstanceMinAvailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                           "false",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":                      "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.minAvailable": "50%",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled":             "true",
			},
			[]string{"50%", "1"},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MinAvailable, "MinAvailable should be empty.")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(testCase.expected[i], pdb.Spec.MinAvailable.String(), "MinAvailable should be equal.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMaxUnavailableInt() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []int
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":        "nil",
			},
			nil,
		},
		{
			"defaultValuesDOTemplatePdbEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.maxUnavailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":        "nil",
			},
			[]int{1},
		},
		{
			"overrideBaseTemplateAndInstanceMaxUnavailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                             "false",
				"delegatedOperatorDeployments.template.podDisruptionBudget.maxUnavailable":                      "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.maxUnavailable": "5",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled":               "true",
			},
			[]int{5, 1},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(testCase.expected[i], pdb.Spec.MaxUnavailable.IntValue(), "MaxUnavailable should be equal.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMaxUnavailablePercent() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"overrideBaseTemplateMaxUnavailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.maxUnavailable": "10%",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":        "nil",
			},
			[]string{"10%"},
		},
		{
			"overrideBaseTemplateAndInstanceMaxUnavailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                             "false",
				"delegatedOperatorDeployments.template.podDisruptionBudget.maxUnavailable":                      "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.maxUnavailable": "50%",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled":               "true",
			},
			[]string{"50%", "1"},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(testCase.expected[i], pdb.Spec.MaxUnavailable.String(), "MaxUnavailable should be equal.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestMinAvaiableAndMaxAvailable() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"overrideBaseTemplateMaxUnavailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.maxUnavailable": "80%",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":   "10%",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":        "nil",
			},
			[]string{"10%"},
		},
		{
			"overrideBaseTemplateAndInstanceMaxUnavailable",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                             "false",
				"delegatedOperatorDeployments.template.podDisruptionBudget.maxUnavailable":                      "10%",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":                        "30%",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled":        "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.maxUnavailable": "20%",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.minAvailable":   "40%",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled":               "true",
			},
			[]string{"40%", "30%"},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.MinAvailable, "MinAvailable should be empty.")
				s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")

			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(testCase.expected[i], pdb.Spec.MinAvailable.String(), "MinAvailable should be equal.")
					s.Empty(pdb.Spec.MaxUnavailable, "MaxUnavailable should be empty.")
				}
			}
		})
	}
}

func (s *pdbDelegatedOperatorInstanceTemplateTest) TestSelectorLabels() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []map[string]string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused": "nil",
			},
			nil,
		},
		{
			"defaultValuesDOTemplatePdbEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":      "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable": "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":      "nil",
			},
			[]map[string]string{
				map[string]string{
					"app.kubernetes.io/name":     "teams-do-cpu-default",
					"app.kubernetes.io/instance": "fiftyone-test",
				},
			},
		},
		{
			"overrideBaseTemplateEnabledAndInstanceDisabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":                      "true",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.podDisruptionBudget.enabled": "false",
			},
			nil,
		},
		{
			"overrideBaseTemplateEnabledAndInstanceEnabled",
			map[string]string{
				"delegatedOperatorDeployments.template.podDisruptionBudget.enabled":               "true",
				"delegatedOperatorDeployments.template.podDisruptionBudget.minAvailable":          "1",
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.unused":               "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podDisruptionBudget.enabled": "false",
			},
			[]map[string]string{
				map[string]string{
					"app.kubernetes.io/name":     "teams-do-cpu-default",
					"app.kubernetes.io/instance": "fiftyone-test",
				},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-pdb.yaml in chart")

				var pdb policyv1.PodDisruptionBudget
				helm.UnmarshalK8SYaml(subT, output, &pdb)

				s.Empty(pdb.Spec.Selector, "Selector Labels should be empty.")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				s.Equal(len(testCase.expected), len(allRange[1:]), "Length of expected outputs should match number of found manifests.")

				for i, rawOutput := range allRange[1:] {

					var pdb policyv1.PodDisruptionBudget

					helm.UnmarshalK8SYaml(subT, rawOutput, &pdb)

					s.Equal(len(testCase.expected[i]), len(pdb.Spec.Selector.MatchLabels), "Label maps should have the same length.")

					for key, value := range testCase.expected[i] {
						foundValue := pdb.Spec.Selector.MatchLabels[key]
						s.Equal(value, foundValue, "Labels should contain all set labels.")
					}
				}
			}
		})
	}
}
