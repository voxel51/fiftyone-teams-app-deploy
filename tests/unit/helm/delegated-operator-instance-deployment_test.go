//go:build kubeall || helm || unit || unitPluginsDeployment
// +build kubeall helm unit unitPluginsDeployment

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
	"k8s.io/apimachinery/pkg/api/resource"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type deploymentDelegatedOperatorInstanceTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

type deploymentDelegatedOperatorInstanceTemplateLabelsExpected struct {
	selectorMatch    map[string]string
	templateMetadata map[string]string
}

func TestDeploymentDelegatedOperatorInstanceTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &deploymentDelegatedOperatorInstanceTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/delegated-operator-instance-deployment.yaml"},
	})
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestMetadataLabels() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]map[string]string{
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do",
					"app.kubernetes.io/instance":   "fiftyone-test",
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]map[string]string{
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do",
					"app.kubernetes.io/instance":   "fiftyone-test",
				},
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do-two",
					"app.kubernetes.io/instance":   "fiftyone-test",
				},
			},
		},
		{
			"overrideName",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoNewName.unused": "nil",
			},
			[]map[string]string{
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do-new-name",
					"app.kubernetes.io/instance":   "fiftyone-test",
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.ObjectMeta.Labels, "Labels should be nil")
			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {

					var deployment appsv1.Deployment

					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					for key, value := range testCase.expected[i] {
						foundValue := deployment.ObjectMeta.Labels[key]
						s.Equal(value, foundValue, "Labels should contain all set labels.")
					}
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestMetadataName() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]string{"teams-do"},
		},
		{
			"multipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]string{"teams-do", "teams-do-two"},
		},
		{
			"overrideMetadataName",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoNewName.unused": "nil",
			},
			[]string{"teams-do-new-name"},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Empty(deployment.ObjectMeta.Name, "Metadata name should be nil")

			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {

					var deployment appsv1.Deployment

					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], deployment.ObjectMeta.Name, "Deployment name should be equal.")

				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestMetadataNamespace() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]string{"fiftyone-teams"},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]string{"fiftyone-teams", "fiftyone-teams"},
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
				"namespace.name": "test-namespace-name",
			},
			[]string{"test-namespace-name"},
		},
		{
			"overrideNamespaceNameMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"namespace.name": "test-namespace-name",
			},
			[]string{"test-namespace-name", "test-namespace-name"},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Empty(deployment.ObjectMeta.Namespace, "Metadata namespace should be nil")

			} else {

				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {

					var deployment appsv1.Deployment

					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], deployment.ObjectMeta.Namespace, "Namespace name should be equal.")

				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestReplicas() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []int32
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]int32{3},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]int32{3, 3},
		},
		{
			"overrideBaseTemplateReplicaCount",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.replicaCount":         "2",
			},
			[]int32{2, 2},
		},
		{
			"overrideInstanceReplicaCount",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.replicaCount":    "2",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.replicaCount": "6",
			},
			[]int32{2, 6},
		},
		{
			"overrideBaseTemplateAndInstanceReplicaCount",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.replicaCount":    "2",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.replicaCount": "6",
				"delegatedOperatorDeployments.template.replicaCount":               "4",
			},
			[]int32{2, 6},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Empty(&deployment.Spec.Replicas, "Replica count should be nil.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], *deployment.Spec.Replicas, "Replica count should be equal.")
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestTopologySpreadConstraints() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(constraint []corev1.TopologySpreadConstraint)
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(constraint []corev1.TopologySpreadConstraint){
				func(constraint []corev1.TopologySpreadConstraint) {
					s.Empty(constraint, "Constrains should be empty")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(constraint []corev1.TopologySpreadConstraint){
				func(constraint []corev1.TopologySpreadConstraint) {
					s.Empty(constraint, "Constrains should be empty")
				},
				func(constraint []corev1.TopologySpreadConstraint) {
					s.Empty(constraint, "Constrains should be empty")
				},
			},
		},
		{
			"overrideBaseTemplateTopologyConstraints",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":                               "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                            "nil",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].matchLabelKeys[0]":  "pod-template-hash",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].maxSkew":            "1",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].minDomains":         "1",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].nodeAffinityPolicy": "Honor",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].nodeTaintsPolicy":   "Honor",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].topologyKey":        "kubernetes.io/hostname",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].whenUnsatisfiable":  "DoNotSchedule",
			},
			[]func(constraint []corev1.TopologySpreadConstraint){
				func(constraint []corev1.TopologySpreadConstraint) {
					var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
					expectedTopologySpreadConstraintJSON := `[
          {
            "matchLabelKeys": [
              "pod-template-hash"
            ],
            "maxSkew": 1,
            "minDomains": 1,
            "nodeAffinityPolicy": "Honor",
            "nodeTaintsPolicy": "Honor",
            "topologyKey": "kubernetes.io/hostname",
            "whenUnsatisfiable": "DoNotSchedule",
            "labelSelector": {
              "matchLabels": {
                "app.kubernetes.io/name": "teams-do",
                "app.kubernetes.io/instance": "fiftyone-test"
              }
            }
          }
        ]`
					err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
					s.NoError(err)
					s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
				},
				func(constraint []corev1.TopologySpreadConstraint) {
					var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
					expectedTopologySpreadConstraintJSON := `[
          {
            "matchLabelKeys": [
              "pod-template-hash"
            ],
            "maxSkew": 1,
            "minDomains": 1,
            "nodeAffinityPolicy": "Honor",
            "nodeTaintsPolicy": "Honor",
            "topologyKey": "kubernetes.io/hostname",
            "whenUnsatisfiable": "DoNotSchedule",
            "labelSelector": {
              "matchLabels": {
                "app.kubernetes.io/name": "teams-do-two",
                "app.kubernetes.io/instance": "fiftyone-test"
              }
            }
          }
        ]`
					err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
					s.NoError(err)
					s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
				},
			},
		},
		{
			"overrideInstanceTopologyConstraints",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].matchLabelKeys[0]":  "pod-template-hash",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].maxSkew":            "2",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].minDomains":         "2",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].nodeAffinityPolicy": "Ignore",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].nodeTaintsPolicy":   "Ignore",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].topologyKey":        "kubernetes.io/hostname",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].whenUnsatisfiable":  "DoNotSchedule",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                                       "nil",
			},
			[]func(constraint []corev1.TopologySpreadConstraint){
				func(constraint []corev1.TopologySpreadConstraint) {
					var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
					expectedTopologySpreadConstraintJSON := `[
          {
            "matchLabelKeys": [
              "pod-template-hash"
            ],
            "maxSkew": 2,
            "minDomains": 2,
            "nodeAffinityPolicy": "Ignore",
            "nodeTaintsPolicy": "Ignore",
            "topologyKey": "kubernetes.io/hostname",
            "whenUnsatisfiable": "DoNotSchedule",
            "labelSelector": {
              "matchLabels": {
                "app.kubernetes.io/name": "teams-do",
                "app.kubernetes.io/instance": "fiftyone-test"
              }
            }
          }
        ]`
					err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
					s.NoError(err)
					s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
				},
				func(constraint []corev1.TopologySpreadConstraint) {
					s.Empty(constraint, "Constrains should be empty")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceTopologyConstraints",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].matchLabelKeys[0]":  "pod-template-hash",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].maxSkew":            "2",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].minDomains":         "2",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].nodeAffinityPolicy": "Ignore",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].nodeTaintsPolicy":   "Ignore",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].topologyKey":        "kubernetes.io/hostname",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].whenUnsatisfiable":  "DoNotSchedule",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                                       "nil",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].matchLabelKeys[0]":             "pod-template-hash",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].maxSkew":                       "1",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].minDomains":                    "1",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].nodeAffinityPolicy":            "Honor",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].nodeTaintsPolicy":              "Honor",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].topologyKey":                   "kubernetes.io/hostname",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].whenUnsatisfiable":             "DoNotSchedule",
			},
			[]func(constraint []corev1.TopologySpreadConstraint){
				func(constraint []corev1.TopologySpreadConstraint) {
					var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
					expectedTopologySpreadConstraintJSON := `[
          {
            "matchLabelKeys": [
              "pod-template-hash"
            ],
            "maxSkew": 2,
            "minDomains": 2,
            "nodeAffinityPolicy": "Ignore",
            "nodeTaintsPolicy": "Ignore",
            "topologyKey": "kubernetes.io/hostname",
            "whenUnsatisfiable": "DoNotSchedule",
            "labelSelector": {
              "matchLabels": {
                "app.kubernetes.io/name": "teams-do",
                "app.kubernetes.io/instance": "fiftyone-test"
              }
            }
          }
        ]`
					err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
					s.NoError(err)
					s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
				},
				func(constraint []corev1.TopologySpreadConstraint) {
					var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
					expectedTopologySpreadConstraintJSON := `[
          {
            "matchLabelKeys": [
              "pod-template-hash"
            ],
            "maxSkew": 1,
            "minDomains": 1,
            "nodeAffinityPolicy": "Honor",
            "nodeTaintsPolicy": "Honor",
            "topologyKey": "kubernetes.io/hostname",
            "whenUnsatisfiable": "DoNotSchedule",
            "labelSelector": {
              "matchLabels": {
                "app.kubernetes.io/name": "teams-do-two",
                "app.kubernetes.io/instance": "fiftyone-test"
              }
            }
          }
        ]`
					err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
					s.NoError(err)
					s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceTopologyConstraintsLabelSelectors",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].matchLabelKeys[0]":             "pod-template-hash",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].maxSkew":                       "2",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].minDomains":                    "2",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].nodeAffinityPolicy":            "Ignore",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].nodeTaintsPolicy":              "Ignore",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].topologyKey":                   "kubernetes.io/hostname",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].whenUnsatisfiable":             "DoNotSchedule",
				"delegatedOperatorDeployments.deployments.teamsDo.topologySpreadConstraints[0].labelSelector.matchLabels.app": "instance-label-override",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                                                  "nil",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].matchLabelKeys[0]":                        "pod-template-hash",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].maxSkew":                                  "1",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].minDomains":                               "1",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].nodeAffinityPolicy":                       "Honor",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].nodeTaintsPolicy":                         "Honor",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].topologyKey":                              "kubernetes.io/hostname",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].whenUnsatisfiable":                        "DoNotSchedule",
				"delegatedOperatorDeployments.template.topologySpreadConstraints[0].labelSelector.matchLabels.app":            "template-label-override",
			},
			[]func(constraint []corev1.TopologySpreadConstraint){
				func(constraint []corev1.TopologySpreadConstraint) {
					var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
					expectedTopologySpreadConstraintJSON := `[
          {
            "matchLabelKeys": [
              "pod-template-hash"
            ],
            "maxSkew": 2,
            "minDomains": 2,
            "nodeAffinityPolicy": "Ignore",
            "nodeTaintsPolicy": "Ignore",
            "topologyKey": "kubernetes.io/hostname",
            "whenUnsatisfiable": "DoNotSchedule",
            "labelSelector": {
              "matchLabels": {
                "app": "instance-label-override"
              }
            }
          }
        ]`
					err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
					s.NoError(err)
					s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
				},
				func(constraint []corev1.TopologySpreadConstraint) {
					var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
					expectedTopologySpreadConstraintJSON := `[
          {
            "matchLabelKeys": [
              "pod-template-hash"
            ],
            "maxSkew": 1,
            "minDomains": 1,
            "nodeAffinityPolicy": "Honor",
            "nodeTaintsPolicy": "Honor",
            "topologyKey": "kubernetes.io/hostname",
            "whenUnsatisfiable": "DoNotSchedule",
            "labelSelector": {
              "matchLabels": {
                "app": "template-label-override"
              }
            }
          }
        ]`
					err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
					s.NoError(err)
					s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
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

			if testCase.values == nil {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Empty(deployment.Spec.Template.Spec.TopologySpreadConstraints, "Topology constraints should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.TopologySpreadConstraints)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerCount() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]int{1},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]int{1, 1},
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(0, len(deployment.Spec.Template.Spec.Containers), "Container count should be equal.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], len(deployment.Spec.Template.Spec.Containers), "Container count should be equal.")
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerEnv() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(envVars []corev1.EnvVar)
	}{
		{
			"defaultValues",
			nil,
			[]func(envVars []corev1.EnvVar){
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(envVars []corev1.EnvVar){
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_DATABASE_ADMIN",
            "value": "false"
          },
          {
            "name": "FIFTYONE_DATABASE_NAME",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneDatabaseName"
              }
            }
          },
          {
            "name": "FIFTYONE_DATABASE_URI",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "mongodbConnectionString"
              }
            }
          },
          {
            "name": "FIFTYONE_ENCRYPTION_KEY",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "encryptionKey"
              }
            }
          },
          {
            "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
            "value": ""
          },
          {
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
            "value": "-1"
          }
        ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(envVars []corev1.EnvVar){
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_DATABASE_ADMIN",
            "value": "false"
          },
          {
            "name": "FIFTYONE_DATABASE_NAME",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneDatabaseName"
              }
            }
          },
          {
            "name": "FIFTYONE_DATABASE_URI",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "mongodbConnectionString"
              }
            }
          },
          {
            "name": "FIFTYONE_ENCRYPTION_KEY",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "encryptionKey"
              }
            }
          },
          {
            "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
            "value": ""
          },
          {
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
            "value": "-1"
          }
        ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_DATABASE_ADMIN",
            "value": "false"
          },
          {
            "name": "FIFTYONE_DATABASE_NAME",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneDatabaseName"
              }
            }
          },
          {
            "name": "FIFTYONE_DATABASE_URI",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "mongodbConnectionString"
              }
            }
          },
          {
            "name": "FIFTYONE_ENCRYPTION_KEY",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "encryptionKey"
              }
            }
          },
          {
            "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
            "value": ""
          },
          {
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
            "value": "-1"
          }
        ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateEnv",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":                             "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                          "nil",
				"delegatedOperatorDeployments.template.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH":     "gs://template",
				"delegatedOperatorDeployments.template.env.TEST_KEY":                                  "TEMPLATE_TEST_VALUE",
				"delegatedOperatorDeployments.template.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretName": "template-existing-secret", // pragma: allowlist secret
				"delegatedOperatorDeployments.template.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretKey":  "templateAnExistingKey",    // pragma: allowlist secret
			},
			[]func(envVars []corev1.EnvVar){
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_DATABASE_ADMIN",
            "value": "false"
          },
          {
            "name": "FIFTYONE_DATABASE_NAME",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneDatabaseName"
              }
            }
          },
          {
            "name": "FIFTYONE_DATABASE_URI",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "mongodbConnectionString"
              }
            }
          },
          {
            "name": "FIFTYONE_ENCRYPTION_KEY",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "encryptionKey"
              }
            }
          },
          {
            "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
            "value": "gs://template"
          },
          {
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
            "value": "-1"
          },
          {
            "name": "TEST_KEY",
            "value": "TEMPLATE_TEST_VALUE"
          },
          {
            "name": "AN_ADDITIONAL_SECRET_ENV",
            "valueFrom": {
              "secretKeyRef": {
                "name": "template-existing-secret",
                "key": "templateAnExistingKey"
              }
            }
          }
        ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_DATABASE_ADMIN",
            "value": "false"
          },
          {
            "name": "FIFTYONE_DATABASE_NAME",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneDatabaseName"
              }
            }
          },
          {
            "name": "FIFTYONE_DATABASE_URI",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "mongodbConnectionString"
              }
            }
          },
          {
            "name": "FIFTYONE_ENCRYPTION_KEY",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "encryptionKey"
              }
            }
          },
          {
            "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
            "value": "gs://template"
          },
          {
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
            "value": "-1"
          },
          {
            "name": "TEST_KEY",
            "value": "TEMPLATE_TEST_VALUE"
          },
          {
            "name": "AN_ADDITIONAL_SECRET_ENV",
            "valueFrom": {
              "secretKeyRef": {
                "name": "template-existing-secret",
                "key": "templateAnExistingKey"
              }
            }
          }
        ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
			},
		},
		{
			"overrideInstanceEnv",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH":     "gs://foo.com",
				"delegatedOperatorDeployments.deployments.teamsDo.env.TEST_KEY":                                  "INSTANCE_TEST_VALUE",
				"delegatedOperatorDeployments.deployments.teamsDo.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretName": "instance-existing-secret", // pragma: allowlist secret
				"delegatedOperatorDeployments.deployments.teamsDo.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretKey":  "instanceAnExistingKey",    // pragma: allowlist secret
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                                     "nil",
			},
			[]func(envVars []corev1.EnvVar){
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
                  {
                    "name": "API_URL",
                    "value": "http://teams-api:80"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_ADMIN",
                    "value": "false"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_NAME",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "fiftyoneDatabaseName"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DATABASE_URI",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "mongodbConnectionString"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_ENCRYPTION_KEY",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "encryptionKey"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
                    "value": "gs://foo.com"
                  },
                  {
                    "name": "FIFTYONE_INTERNAL_SERVICE",
                    "value": "true"
                  },
                  {
                    "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
                    "value": "-1"
                  },
                  {
                    "name": "TEST_KEY",
                    "value": "INSTANCE_TEST_VALUE"
                  },
                  {
                    "name": "AN_ADDITIONAL_SECRET_ENV",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "instance-existing-secret",
                        "key": "instanceAnExistingKey"
                      }
                    }
                  }
                ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
                  {
                    "name": "API_URL",
                    "value": "http://teams-api:80"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_ADMIN",
                    "value": "false"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_NAME",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "fiftyoneDatabaseName"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DATABASE_URI",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "mongodbConnectionString"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_ENCRYPTION_KEY",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "encryptionKey"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
                    "value": ""
                  },
                  {
                    "name": "FIFTYONE_INTERNAL_SERVICE",
                    "value": "true"
                  },
                  {
                    "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
                    "value": "-1"
                  }
                ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceEnv",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH":     "gs://foo.com",
				"delegatedOperatorDeployments.deployments.teamsDo.env.TEST_KEY":                                  "INSTANCE_TEST_VALUE",
				"delegatedOperatorDeployments.deployments.teamsDo.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretName": "instance-existing-secret", // pragma: allowlist secret
				"delegatedOperatorDeployments.deployments.teamsDo.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretKey":  "instanceAnExistingKey",    // pragma: allowlist secret
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                                     "nil",
				"delegatedOperatorDeployments.template.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH":                "gs://template",
				"delegatedOperatorDeployments.template.env.TEST_KEY":                                             "TEMPLATE_TEST_VALUE",
				"delegatedOperatorDeployments.template.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretName":            "template-existing-secret", // pragma: allowlist secret
				"delegatedOperatorDeployments.template.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretKey":             "templateAnExistingKey",    // pragma: allowlist secret
			},
			[]func(envVars []corev1.EnvVar){
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
                  {
                    "name": "API_URL",
                    "value": "http://teams-api:80"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_ADMIN",
                    "value": "false"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_NAME",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "fiftyoneDatabaseName"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DATABASE_URI",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "mongodbConnectionString"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_ENCRYPTION_KEY",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "encryptionKey"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
                    "value": "gs://foo.com"
                  },
                  {
                    "name": "FIFTYONE_INTERNAL_SERVICE",
                    "value": "true"
                  },
                  {
                    "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
                    "value": "-1"
                  },
                  {
                    "name": "TEST_KEY",
                    "value": "INSTANCE_TEST_VALUE"
                  },
                  {
                    "name": "AN_ADDITIONAL_SECRET_ENV",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "instance-existing-secret",
                        "key": "instanceAnExistingKey"
                      }
                    }
                  }
                ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
				func(envVars []corev1.EnvVar) {
					expectedEnvVarJSON := `[
                  {
                    "name": "API_URL",
                    "value": "http://teams-api:80"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_ADMIN",
                    "value": "false"
                  },
                  {
                    "name": "FIFTYONE_DATABASE_NAME",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "fiftyoneDatabaseName"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DATABASE_URI",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "mongodbConnectionString"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_ENCRYPTION_KEY",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "fiftyone-teams-secrets",
                        "key": "encryptionKey"
                      }
                    }
                  },
                  {
                    "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
                    "value": "gs://template"
                  },
                  {
                    "name": "FIFTYONE_INTERNAL_SERVICE",
                    "value": "true"
                  },
                  {
                    "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
                    "value": "-1"
                  },
                  {
                    "name": "TEST_KEY",
                    "value": "TEMPLATE_TEST_VALUE"
                  },
                  {
                    "name": "AN_ADDITIONAL_SECRET_ENV",
                    "valueFrom": {
                      "secretKeyRef": {
                        "name": "template-existing-secret",
                        "key": "templateAnExistingKey"
                      }
                    }
                  }
                ]`
					var expectedEnvVars []corev1.EnvVar
					err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
					s.NoError(err)
					s.Equal(expectedEnvVars, envVars, "Envs should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].Env)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerImage() {

	// Get chart info (to later obtain the chart's appVersion)
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)

	// Get appVersion from chart info
	chartAppVersion, exists := cInfo["appVersion"]
	s.True(exists, "failed to get app version from chart info")

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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]string{fmt.Sprintf("voxel51/fiftyone-teams-cv-full:%s", chartAppVersion)},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]string{
				fmt.Sprintf("voxel51/fiftyone-teams-cv-full:%s", chartAppVersion),
				fmt.Sprintf("voxel51/fiftyone-teams-cv-full:%s", chartAppVersion),
			},
		},
		// Base Template Tests
		{
			"overrideBaseTemplateImageTag",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.image.tag":            "testTag",
			},
			[]string{
				"voxel51/fiftyone-teams-cv-full:testTag",
				"voxel51/fiftyone-teams-cv-full:testTag",
			},
		},
		{
			"overrideBaseTemplateImageRepository",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.image.repository":     "ghcr.io/fiftyone-teams-cv-full",
			},
			[]string{
				fmt.Sprintf("ghcr.io/fiftyone-teams-cv-full:%s", chartAppVersion),
				fmt.Sprintf("ghcr.io/fiftyone-teams-cv-full:%s", chartAppVersion),
			},
		},
		{
			"overrideBaseTemplateImageRepository",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.image.tag":            "testTag",
				"delegatedOperatorDeployments.template.image.repository":     "ghcr.io/fiftyone-teams-cv-full",
			},
			[]string{
				"ghcr.io/fiftyone-teams-cv-full:testTag",
				"ghcr.io/fiftyone-teams-cv-full:testTag",
			},
		},
		// Instance Tests
		{
			"overrideInstanceImageTag",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.tag":    "foo",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.tag": "bar",
			},
			[]string{
				"voxel51/fiftyone-teams-cv-full:foo",
				"voxel51/fiftyone-teams-cv-full:bar",
			},
		},
		{
			"overrideInstanceImageRepository",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.repository":    "ghcr.io/fiftyone-teams-cv-full",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.repository": "ghcr.io/fiftyone-teams-cv-slim",
			},
			[]string{
				fmt.Sprintf("ghcr.io/fiftyone-teams-cv-full:%s", chartAppVersion),
				fmt.Sprintf("ghcr.io/fiftyone-teams-cv-slim:%s", chartAppVersion),
			},
		},
		{
			"overrideInstanceImageTagAndRepository",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.tag":           "foo",
				"delegatedOperatorDeployments.deployments.teamsDo.image.repository":    "ghcr.io/fiftyone-teams-cv-full",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.tag":        "bar",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.repository": "ghcr.io/fiftyone-teams-cv-slim",
			},
			[]string{
				"ghcr.io/fiftyone-teams-cv-full:foo",
				"ghcr.io/fiftyone-teams-cv-slim:bar",
			},
		},
		// Conflict Tests
		{
			"overrideBaseTemplateAndInstanceImageTag",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.tag":    "foo",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.tag": "bar",
				"delegatedOperatorDeployments.template.image.tag":               "biz",
			},
			[]string{
				"voxel51/fiftyone-teams-cv-full:foo",
				"voxel51/fiftyone-teams-cv-full:bar",
			},
		},
		{
			"overrideBaseTemplateAndInstanceImageRepository",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.repository":    "ghcr.io/fiftyone-teams-cv-full",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.repository": "ghcr.io/fiftyone-teams-cv-slim",
				"delegatedOperatorDeployments.template.image.repository":               "ghcr.io/fiftyone-teams-cv-template",
			},
			[]string{
				fmt.Sprintf("ghcr.io/fiftyone-teams-cv-full:%s", chartAppVersion),
				fmt.Sprintf("ghcr.io/fiftyone-teams-cv-slim:%s", chartAppVersion),
			},
		},
		{
			"overrideBaseTemplateAndInstanceImageTagAndRepository",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.tag":           "foo",
				"delegatedOperatorDeployments.deployments.teamsDo.image.repository":    "ghcr.io/fiftyone-teams-cv-full",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.tag":        "bar",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.repository": "ghcr.io/fiftyone-teams-cv-slim",
				"delegatedOperatorDeployments.template.image.tag":                      "biz",
				"delegatedOperatorDeployments.template.image.repository":               "ghcr.io/fiftyone-teams-cv-template",
			},
			[]string{
				"ghcr.io/fiftyone-teams-cv-full:foo",
				"ghcr.io/fiftyone-teams-cv-slim:bar",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], deployment.Spec.Template.Spec.Containers[0].Image, "Image values should be equal.")
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerImagePullPolicy() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]string{"Always"},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]string{"Always", "Always"},
		},
		{
			"overrideBaseTemplatePullPolicy",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.image.pullPolicy":     "IfNotPresent",
			},
			[]string{"IfNotPresent", "IfNotPresent"},
		},
		{
			"overrideInstancePullPolicy",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.pullPolicy":    "Always",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.pullPolicy": "Never",
			},
			[]string{"Always", "Never"},
		},
		{
			"overrideBaseTemplateAndInstancePullPolicy",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.image.pullPolicy":    "Always",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.image.pullPolicy": "Never",
				"delegatedOperatorDeployments.template.image.pullPolicy":               "IfNotPresent",
			},
			[]string{"Always", "Never"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], string(deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy), "Image pull policy should be equal.")
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerName() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]string{"teams-do"},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]string{"teams-do", "teams-do-two"},
		},
		// Names not overridable, so omitting tests
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], deployment.Spec.Template.Spec.Containers[0].Name, "Container name should be equal.")
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerResourceRequirements() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(resourceRequirements corev1.ResourceRequirements)
	}{
		{
			"defaultValues",
			nil,
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					s.Empty(resourceRequirements, "Resource Requirements should be empty")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					s.Equal(resourceRequirements.Limits, corev1.ResourceList{}, "Limits should be equal")
					s.Equal(resourceRequirements.Requests, corev1.ResourceList{}, "Requests should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					s.Equal(resourceRequirements.Limits, corev1.ResourceList{}, "Limits should be equal")
					s.Equal(resourceRequirements.Requests, corev1.ResourceList{}, "Requests should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
				func(resourceRequirements corev1.ResourceRequirements) {
					s.Equal(resourceRequirements.Limits, corev1.ResourceList{}, "Limits should be equal")
					s.Equal(resourceRequirements.Requests, corev1.ResourceList{}, "Requests should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateResources",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":         "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":      "nil",
				"delegatedOperatorDeployments.template.resources.limits.cpu":      "1",
				"delegatedOperatorDeployments.template.resources.limits.memory":   "1Gi",
				"delegatedOperatorDeployments.template.resources.requests.cpu":    "500m",
				"delegatedOperatorDeployments.template.resources.requests.memory": "512Mi",
			},
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("1Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("500m"),
							"memory": resource.MustParse("512Mi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("1Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("500m"),
							"memory": resource.MustParse("512Mi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
			},
		},
		{
			"overrideInstanceResources",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.resources.limits.cpu":         "3",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.limits.memory":      "3Gi",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.requests.cpu":       "2",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.requests.memory":    "2Gi",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.limits.cpu":      "4",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.limits.memory":   "4Gi",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.requests.cpu":    "3",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.requests.memory": "3Gi",
			},
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("3"),
							"memory": resource.MustParse("3Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("2Gi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("4"),
							"memory": resource.MustParse("4Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("3"),
							"memory": resource.MustParse("3Gi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceResourcesLimits",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.resources.limits.cpu":       "3",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.limits.memory":    "3Gi",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.limits.cpu":    "4",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.limits.memory": "4Gi",
				"delegatedOperatorDeployments.template.resources.limits.cpu":                  "1",
				"delegatedOperatorDeployments.template.resources.limits.memory":               "1Gi",
				"delegatedOperatorDeployments.template.resources.requests.cpu":                "500m",
				"delegatedOperatorDeployments.template.resources.requests.memory":             "512Mi",
			},
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("3"),
							"memory": resource.MustParse("3Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("500m"),
							"memory": resource.MustParse("512Mi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("4"),
							"memory": resource.MustParse("4Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("500m"),
							"memory": resource.MustParse("512Mi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceResourcesRequests",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.resources.requests.cpu":       "2",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.requests.memory":    "2Gi",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.requests.cpu":    "3",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.requests.memory": "3Gi",
				"delegatedOperatorDeployments.template.resources.limits.cpu":                    "1",
				"delegatedOperatorDeployments.template.resources.limits.memory":                 "1Gi",
				"delegatedOperatorDeployments.template.resources.requests.cpu":                  "500m",
				"delegatedOperatorDeployments.template.resources.requests.memory":               "512Mi",
			},
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("1Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("2Gi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("1Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("3"),
							"memory": resource.MustParse("3Gi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceResources",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.resources.limits.cpu":         "3",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.limits.memory":      "3Gi",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.requests.cpu":       "2",
				"delegatedOperatorDeployments.deployments.teamsDo.resources.requests.memory":    "2Gi",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.limits.cpu":      "4",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.limits.memory":   "4Gi",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.requests.cpu":    "3",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.resources.requests.memory": "3Gi",
				"delegatedOperatorDeployments.template.resources.limits.cpu":                    "1",
				"delegatedOperatorDeployments.template.resources.limits.memory":                 "1Gi",
				"delegatedOperatorDeployments.template.resources.requests.cpu":                  "500m",
				"delegatedOperatorDeployments.template.resources.requests.memory":               "512Mi",
			},
			[]func(resourceRequirements corev1.ResourceRequirements){
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("3"),
							"memory": resource.MustParse("3Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("2Gi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
				func(resourceRequirements corev1.ResourceRequirements) {
					resourceExpected := corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("4"),
							"memory": resource.MustParse("4Gi"),
						},
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("3"),
							"memory": resource.MustParse("3Gi"),
						},
					}
					s.Equal(resourceExpected, resourceRequirements, "should be equal")
					s.Nil(resourceRequirements.Claims, "should be nil")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].Resources)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerSecurityContext() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(securityContext *corev1.SecurityContext)
	}{
		{
			"defaultValues",
			nil,
			[]func(securityContext *corev1.SecurityContext){
				func(securityContext *corev1.SecurityContext) {
					s.Empty(securityContext, "should be not be set")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(securityContext *corev1.SecurityContext){
				func(securityContext *corev1.SecurityContext) {
					s.Nil(securityContext.AllowPrivilegeEscalation, "should be nil")
					s.Nil(securityContext.Capabilities, "should be nil")
					s.Nil(securityContext.Privileged, "should be nil")
					s.Nil(securityContext.ProcMount, "should be nil")
					s.Nil(securityContext.ReadOnlyRootFilesystem, "should be nil")
					s.Nil(securityContext.RunAsGroup, "should be nil")
					s.Nil(securityContext.RunAsNonRoot, "should be nil")
					s.Nil(securityContext.RunAsUser, "should be nil")
					s.Nil(securityContext.SeccompProfile, "should be nil")
					s.Nil(securityContext.SELinuxOptions, "should be nil")
					s.Nil(securityContext.WindowsOptions, "should be nil")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(securityContext *corev1.SecurityContext){
				func(securityContext *corev1.SecurityContext) {
					s.Nil(securityContext.AllowPrivilegeEscalation, "should be nil")
					s.Nil(securityContext.Capabilities, "should be nil")
					s.Nil(securityContext.Privileged, "should be nil")
					s.Nil(securityContext.ProcMount, "should be nil")
					s.Nil(securityContext.ReadOnlyRootFilesystem, "should be nil")
					s.Nil(securityContext.RunAsGroup, "should be nil")
					s.Nil(securityContext.RunAsNonRoot, "should be nil")
					s.Nil(securityContext.RunAsUser, "should be nil")
					s.Nil(securityContext.SeccompProfile, "should be nil")
					s.Nil(securityContext.SELinuxOptions, "should be nil")
					s.Nil(securityContext.WindowsOptions, "should be nil")
				},
				func(securityContext *corev1.SecurityContext) {
					s.Nil(securityContext.AllowPrivilegeEscalation, "should be nil")
					s.Nil(securityContext.Capabilities, "should be nil")
					s.Nil(securityContext.Privileged, "should be nil")
					s.Nil(securityContext.ProcMount, "should be nil")
					s.Nil(securityContext.ReadOnlyRootFilesystem, "should be nil")
					s.Nil(securityContext.RunAsGroup, "should be nil")
					s.Nil(securityContext.RunAsNonRoot, "should be nil")
					s.Nil(securityContext.RunAsUser, "should be nil")
					s.Nil(securityContext.SeccompProfile, "should be nil")
					s.Nil(securityContext.SELinuxOptions, "should be nil")
					s.Nil(securityContext.WindowsOptions, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateSecurityContext",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":          "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":       "nil",
				"delegatedOperatorDeployments.template.securityContext.runAsGroup": "3000",
				"delegatedOperatorDeployments.template.securityContext.runAsUser":  "1000",
			},
			[]func(securityContext *corev1.SecurityContext){
				func(securityContext *corev1.SecurityContext) {
					s.Equal(int64(3000), *securityContext.RunAsGroup, "runAsGroup should be 3000")
					s.Equal(int64(1000), *securityContext.RunAsUser, "runAsUser should be 1000")
				},
				func(securityContext *corev1.SecurityContext) {
					s.Equal(int64(3000), *securityContext.RunAsGroup, "runAsGroup should be 3000")
					s.Equal(int64(1000), *securityContext.RunAsUser, "runAsUser should be 1000")
				},
			},
		},
		{
			"overrideInstanceSecurityContext",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.securityContext.runAsGroup":    "4000",
				"delegatedOperatorDeployments.deployments.teamsDo.securityContext.runAsUser":     "1001",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.securityContext.runAsGroup": "5000",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.securityContext.runAsUser":  "1002",
			},
			[]func(securityContext *corev1.SecurityContext){
				func(securityContext *corev1.SecurityContext) {
					s.Equal(int64(4000), *securityContext.RunAsGroup, "runAsGroup should be 4000")
					s.Equal(int64(1001), *securityContext.RunAsUser, "runAsUser should be 1001")
				},
				func(securityContext *corev1.SecurityContext) {
					s.Equal(int64(5000), *securityContext.RunAsGroup, "runAsGroup should be 5000")
					s.Equal(int64(1002), *securityContext.RunAsUser, "runAsUser should be 1002")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceSecurityContext",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.securityContext.runAsGroup":    "4000",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.securityContext.runAsGroup": "5000",
				"delegatedOperatorDeployments.template.securityContext.runAsGroup":               "3000",
				"delegatedOperatorDeployments.template.securityContext.runAsUser":                "1000",
			},
			[]func(securityContext *corev1.SecurityContext){
				func(securityContext *corev1.SecurityContext) {
					s.Equal(int64(4000), *securityContext.RunAsGroup, "runAsGroup should be 4000")
					s.Equal(int64(1000), *securityContext.RunAsUser, "runAsUser should be 1000")
				},
				func(securityContext *corev1.SecurityContext) {
					s.Equal(int64(5000), *securityContext.RunAsGroup, "runAsGroup should be 5000")
					s.Equal(int64(1000), *securityContext.RunAsUser, "runAsUser should be 1000")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].SecurityContext)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerVolumeMounts() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(volumeMounts []corev1.VolumeMount)
	}{
		{
			"defaultValues",
			nil,
			[]func(volumeMounts []corev1.VolumeMount){
				func(volumeMounts []corev1.VolumeMount) {
					s.Empty(volumeMounts, "VolumeMounts should not be set")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(volumeMounts []corev1.VolumeMount){
				func(volumeMounts []corev1.VolumeMount) {
					s.Nil(volumeMounts, "VolumeMounts should be nil")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(volumeMounts []corev1.VolumeMount){
				func(volumeMounts []corev1.VolumeMount) {
					s.Nil(volumeMounts, "VolumeMounts should be nil")
				},
				func(volumeMounts []corev1.VolumeMount) {
					s.Nil(volumeMounts, "VolumeMounts should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateVolumeMounts",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":         "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":      "nil",
				"delegatedOperatorDeployments.template.volumeMounts[0].mountPath": "/template-test-data-volume",
				"delegatedOperatorDeployments.template.volumeMounts[0].name":      "template-test-volume",
			},
			[]func(volumeMounts []corev1.VolumeMount){
				func(volumeMounts []corev1.VolumeMount) {
					expectedJSON := `[
          {
            "mountPath": "/template-test-data-volume",
            "name": "template-test-volume"
          }
        ]`
					var expectedVolumeMounts []corev1.VolumeMount
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
					s.NoError(err)
					s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
				},
				func(volumeMounts []corev1.VolumeMount) {
					expectedJSON := `[
          {
            "mountPath": "/template-test-data-volume",
            "name": "template-test-volume"
          }
        ]`
					var expectedVolumeMounts []corev1.VolumeMount
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
					s.NoError(err)
					s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
				},
			},
		},
		{
			"overrideInstanceVolumeMounts",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.volumeMounts[0].mountPath": "/teams-do-test-data-volume",
				"delegatedOperatorDeployments.deployments.teamsDo.volumeMounts[0].name":      "teams-do-test-volume",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                 "nil",
			},
			[]func(volumeMounts []corev1.VolumeMount){
				func(volumeMounts []corev1.VolumeMount) {
					expectedJSON := `[
          {
            "mountPath": "/teams-do-test-data-volume",
            "name": "teams-do-test-volume"
          }
        ]`
					var expectedVolumeMounts []corev1.VolumeMount
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
					s.NoError(err)
					s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
				},
				func(volumeMounts []corev1.VolumeMount) {
					s.Nil(volumeMounts, "VolumeMounts should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceVolumeMounts",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.volumeMounts[0].mountPath": "/teams-do-test-data-volume",
				"delegatedOperatorDeployments.deployments.teamsDo.volumeMounts[0].name":      "teams-do-test-volume",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                 "nil",
				"delegatedOperatorDeployments.template.volumeMounts[0].mountPath":            "/template-test-data-volume",
				"delegatedOperatorDeployments.template.volumeMounts[0].name":                 "template-test-volume",
			},
			[]func(volumeMounts []corev1.VolumeMount){
				func(volumeMounts []corev1.VolumeMount) {
					expectedJSON := `[
          {
            "mountPath": "/teams-do-test-data-volume",
            "name": "teams-do-test-volume"
          }
        ]`
					var expectedVolumeMounts []corev1.VolumeMount
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
					s.NoError(err)
					s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
				},
				func(volumeMounts []corev1.VolumeMount) {
					expectedJSON := `[
          {
            "mountPath": "/template-test-data-volume",
            "name": "template-test-volume"
          }
        ]`
					var expectedVolumeMounts []corev1.VolumeMount
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
					s.NoError(err)
					s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
				}
			}

		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestAffinity() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(affinity *corev1.Affinity)
	}{
		{
			"defaultValues",
			nil,
			[]func(affinity *corev1.Affinity){
				func(affinity *corev1.Affinity) {
					s.Nil(affinity, "should be nil")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(affinity *corev1.Affinity){
				func(affinity *corev1.Affinity) {
					s.Nil(affinity, "should be nil")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(affinity *corev1.Affinity){
				func(affinity *corev1.Affinity) {
					s.Nil(affinity, "should be nil")
				},
				func(affinity *corev1.Affinity) {
					s.Nil(affinity, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateAffinity",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "disktype",
				"delegatedOperatorDeployments.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
				"delegatedOperatorDeployments.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "ssd",
			},
			[]func(affinity *corev1.Affinity){
				func(affinity *corev1.Affinity) {
					affinityJSON := `{
          "nodeAffinity": {
            "requiredDuringSchedulingIgnoredDuringExecution": {
              "nodeSelectorTerms": [
                {
                  "matchExpressions": [
                    {
                      "key": "disktype",
                      "operator": "In",
                      "values": [
                        "ssd"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }`
					var expectedAffinity corev1.Affinity
					err := json.Unmarshal([]byte(affinityJSON), &expectedAffinity)
					s.NoError(err)

					s.Equal(expectedAffinity, *affinity, "Affinity should be equal")
				},
				func(affinity *corev1.Affinity) {
					affinityJSON := `{
          "nodeAffinity": {
            "requiredDuringSchedulingIgnoredDuringExecution": {
              "nodeSelectorTerms": [
                {
                  "matchExpressions": [
                    {
                      "key": "disktype",
                      "operator": "In",
                      "values": [
                        "ssd"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }`
					var expectedAffinity corev1.Affinity
					err := json.Unmarshal([]byte(affinityJSON), &expectedAffinity)
					s.NoError(err)

					s.Equal(expectedAffinity, *affinity, "Affinity should be equal")
				},
			},
		},
		{
			"overrideInstanceAffinity",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight":                                      "1",
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key":          "topology.kubernetes.io/zone",
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].operator":     "In",
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].values[0]":    "antarctica-west1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight":                                   "1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key":       "topology.kubernetes.io/zone",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].operator":  "In",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].values[0]": "antarctica-east1",
			},
			[]func(affinity *corev1.Affinity){
				func(affinity *corev1.Affinity) {
					affinityJSON := `{
          "nodeAffinity": {
            "preferredDuringSchedulingIgnoredDuringExecution": [
              {
                "weight": 1,
                "preference": {
                  "matchExpressions": [
                    {
                      "key": "topology.kubernetes.io/zone",
                      "operator": "In",
                      "values": [
                        "antarctica-west1"
                      ]
                    }
                  ]
                }
			  }
			]
          }
        }`
					var expectedAffinity corev1.Affinity
					err := json.Unmarshal([]byte(affinityJSON), &expectedAffinity)
					s.NoError(err)

					s.Equal(expectedAffinity, *affinity, "Affinity should be equal")
				},
				func(affinity *corev1.Affinity) {
					affinityJSON := `{
          "nodeAffinity": {
            "preferredDuringSchedulingIgnoredDuringExecution": [
              {
                "weight": 1,
                "preference": {
                  "matchExpressions": [
                    {
                      "key": "topology.kubernetes.io/zone",
                      "operator": "In",
                      "values": [
                        "antarctica-east1"
                      ]
                    }
                  ]
                }
			  }
			]
          }
        }`
					var expectedAffinity corev1.Affinity
					err := json.Unmarshal([]byte(affinityJSON), &expectedAffinity)
					s.NoError(err)

					s.Equal(expectedAffinity, *affinity, "Affinity should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceAffinity",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight":                                      "1",
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key":          "topology.kubernetes.io/zone",
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].operator":     "In",
				"delegatedOperatorDeployments.deployments.teamsDo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].values[0]":    "antarctica-west1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight":                                   "1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].key":       "topology.kubernetes.io/zone",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].operator":  "In",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].preference.matchExpressions[0].values[0]": "antarctica-east1",
				"delegatedOperatorDeployments.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":               "disktype",
				"delegatedOperatorDeployments.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":          "In",
				"delegatedOperatorDeployments.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]":         "ssd",
			},
			[]func(affinity *corev1.Affinity){
				func(affinity *corev1.Affinity) {
					affinityJSON := `{
          "nodeAffinity": {
            "preferredDuringSchedulingIgnoredDuringExecution": [
              {
                "weight": 1,
                "preference": {
                  "matchExpressions": [
                    {
                      "key": "topology.kubernetes.io/zone",
                      "operator": "In",
                      "values": [
                        "antarctica-west1"
                      ]
                    }
                  ]
                }
			  }
			],
            "requiredDuringSchedulingIgnoredDuringExecution": {
              "nodeSelectorTerms": [
                {
                  "matchExpressions": [
                    {
                      "key": "disktype",
                      "operator": "In",
                      "values": [
                        "ssd"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }`
					var expectedAffinity corev1.Affinity
					err := json.Unmarshal([]byte(affinityJSON), &expectedAffinity)
					s.NoError(err)

					s.Equal(expectedAffinity, *affinity, "Affinity should be equal")
				},
				func(affinity *corev1.Affinity) {
					affinityJSON := `{
          "nodeAffinity": {
            "preferredDuringSchedulingIgnoredDuringExecution": [
              {
                "weight": 1,
                "preference": {
                  "matchExpressions": [
                    {
                      "key": "topology.kubernetes.io/zone",
                      "operator": "In",
                      "values": [
                        "antarctica-east1"
                      ]
                    }
                  ]
                }
			  }
			],
            "requiredDuringSchedulingIgnoredDuringExecution": {
              "nodeSelectorTerms": [
                {
                  "matchExpressions": [
                    {
                      "key": "disktype",
                      "operator": "In",
                      "values": [
                        "ssd"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }`
					var expectedAffinity corev1.Affinity
					err := json.Unmarshal([]byte(affinityJSON), &expectedAffinity)
					s.NoError(err)

					s.Equal(expectedAffinity, *affinity, "Affinity should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Affinity)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestImagePullSecrets() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]string{""},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]string{"", ""},
		},
		{
			"overrideImagePullSecrets",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"imagePullSecrets[0].name":                                   "test-pull-secret",
			},
			[]string{"test-pull-secret", "test-pull-secret"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					if testCase.expected[i] == "" {
						s.Nil(deployment.Spec.Template.Spec.ImagePullSecrets, "ImagePullSecrets should be nil.")
					} else {
						s.Equal(testCase.expected[i], deployment.Spec.Template.Spec.ImagePullSecrets[0].Name, "Image pull secret should be equal.")
					}
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestNodeSelector() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]map[string]string{
				map[string]string{},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]map[string]string{
				map[string]string{},
				map[string]string{},
			},
		},
		{
			"overrideBaseTemplateNodeSelector",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":     "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":  "nil",
				"delegatedOperatorDeployments.template.nodeSelector.disktype": "ssd",
			},
			[]map[string]string{
				map[string]string{
					"disktype": "ssd",
				},
				map[string]string{
					"disktype": "ssd",
				},
			},
		},
		{
			"overrideInstanceNodeSelector",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.nodeSelector.region":      "us-east1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.nodeSelector.disktype": "pd-standard",
			},
			[]map[string]string{
				map[string]string{
					"region": "us-east1",
				},
				map[string]string{
					"disktype": "pd-standard",
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceNodeSelector",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.nodeSelector.region":      "us-east1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.nodeSelector.disktype": "pd-standard",
				"delegatedOperatorDeployments.template.nodeSelector.disktype":               "ssd",
			},
			[]map[string]string{
				map[string]string{
					"region":   "us-east1",
					"disktype": "ssd",
				},
				map[string]string{
					"disktype": "pd-standard",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				for i, rawOutput := range allRange[1:] {

					var deployment appsv1.Deployment

					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					for key, value := range testCase.expected[i] {
						foundValue := deployment.Spec.Template.Spec.NodeSelector[key]
						s.Equal(value, foundValue, "NodeSelector should contain all set labels.")
					}
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestDeploymentAnnotations() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]map[string]string{
				map[string]string{},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]map[string]string{
				nil,
				nil,
			},
		},
		{
			"overrideBaseTemplateDeploymentAnnotations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":                  "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":               "nil",
				"delegatedOperatorDeployments.template.deploymentAnnotations.annotation-1": "annotation-1-value",
			},
			[]map[string]string{
				map[string]string{
					"annotation-1": "annotation-1-value",
				},
				map[string]string{
					"annotation-1": "annotation-1-value",
				},
			},
		},
		{
			"overrideInstanceDeploymentAnnotations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.deploymentAnnotations.teams-do-annotation-1":        "teams-do-annotation-1-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.deploymentAnnotations.annotation-1":              "teams-do-two-annotation-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.deploymentAnnotations.teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
			},
			[]map[string]string{
				map[string]string{
					"teams-do-annotation-1": "teams-do-annotation-1-value",
				},
				map[string]string{
					"annotation-1":              "teams-do-two-annotation-value",
					"teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceDeploymentAnnotations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.deploymentAnnotations.teams-do-annotation":          "teams-do-annotation-1-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.deploymentAnnotations.annotation-1":              "teams-do-two-annotation-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.deploymentAnnotations.teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
				"delegatedOperatorDeployments.template.deploymentAnnotations.annotation-1":                            "annotation-1-value",
			},
			[]map[string]string{
				map[string]string{
					"annotation-1":        "annotation-1-value",
					"teams-do-annotation": "teams-do-annotation-1-value",
				},
				map[string]string{
					"annotation-1":              "teams-do-two-annotation-value",
					"teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)
				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {

					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					if testCase.expected[i] == nil {
						s.Nil(deployment.ObjectMeta.Annotations, "Annotations should be nil")
					} else {
						for key, value := range testCase.expected[i] {
							foundValue := deployment.ObjectMeta.Annotations[key]
							s.Equal(value, foundValue, "Annotations should contain all set annotations.")
						}
					}
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestPodAnnotations() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]map[string]string{
				map[string]string{},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]map[string]string{
				nil,
				nil,
			},
		},
		{
			"overrideBaseTemplatePodAnnotations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":           "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":        "nil",
				"delegatedOperatorDeployments.template.podAnnotations.annotation-1": "annotation-1-value",
			},
			[]map[string]string{
				map[string]string{
					"annotation-1": "annotation-1-value",
				},
				map[string]string{
					"annotation-1": "annotation-1-value",
				},
			},
		},
		{
			"overrideInstancePodAnnotations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.podAnnotations.teams-do-annotation-1":        "teams-do-annotation-1-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podAnnotations.annotation-1":              "teams-do-two-annotation-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podAnnotations.teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
			},
			[]map[string]string{
				map[string]string{
					"teams-do-annotation-1": "teams-do-annotation-1-value",
				},
				map[string]string{
					"annotation-1":              "teams-do-two-annotation-value",
					"teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
				},
			},
		},
		{
			"overrideBaseTemplateAndInstancePodAnnotations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.podAnnotations.teams-do-annotation":          "teams-do-annotation-1-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podAnnotations.annotation-1":              "teams-do-two-annotation-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podAnnotations.teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
				"delegatedOperatorDeployments.template.podAnnotations.annotation-1":                            "annotation-1-value",
			},
			[]map[string]string{
				map[string]string{
					"annotation-1":        "annotation-1-value",
					"teams-do-annotation": "teams-do-annotation-1-value",
				},
				map[string]string{
					"annotation-1":              "teams-do-two-annotation-value",
					"teams-do-two-annotation-1": "teams-do-two-annotation-1-value",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)
				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {

					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					if testCase.expected[i] == nil {
						s.Nil(deployment.Spec.Template.ObjectMeta.Annotations, "Annotations should be nil")
					} else {
						for key, value := range testCase.expected[i] {
							foundValue := deployment.Spec.Template.ObjectMeta.Annotations[key]
							s.Equal(value, foundValue, "Annotations should contain all set annotations.")
						}
					}
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestPodSecurityContext() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(podSecurityContext *corev1.PodSecurityContext)
	}{
		{
			"defaultValues",
			nil,
			[]func(podSecurityContext *corev1.PodSecurityContext){
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Empty(podSecurityContext.FSGroup, "should not be set")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(podSecurityContext *corev1.PodSecurityContext){
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Nil(podSecurityContext.FSGroup, "should be nil")
					s.Nil(podSecurityContext.FSGroupChangePolicy, "should be nil")
					s.Nil(podSecurityContext.RunAsGroup, "should be nil")
					s.Nil(podSecurityContext.RunAsNonRoot, "should be nil")
					s.Nil(podSecurityContext.RunAsUser, "should be nil")
					s.Nil(podSecurityContext.SeccompProfile, "should be nil")
					s.Nil(podSecurityContext.SELinuxOptions, "should be nil")
					s.Nil(podSecurityContext.SupplementalGroups, "should be nil")
					s.Nil(podSecurityContext.Sysctls, "should be nil")
					s.Nil(podSecurityContext.WindowsOptions, "should be nil")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(podSecurityContext *corev1.PodSecurityContext){
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Nil(podSecurityContext.FSGroup, "should be nil")
					s.Nil(podSecurityContext.FSGroupChangePolicy, "should be nil")
					s.Nil(podSecurityContext.RunAsGroup, "should be nil")
					s.Nil(podSecurityContext.RunAsNonRoot, "should be nil")
					s.Nil(podSecurityContext.RunAsUser, "should be nil")
					s.Nil(podSecurityContext.SeccompProfile, "should be nil")
					s.Nil(podSecurityContext.SELinuxOptions, "should be nil")
					s.Nil(podSecurityContext.SupplementalGroups, "should be nil")
					s.Nil(podSecurityContext.Sysctls, "should be nil")
					s.Nil(podSecurityContext.WindowsOptions, "should be nil")
				},
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Nil(podSecurityContext.FSGroup, "should be nil")
					s.Nil(podSecurityContext.FSGroupChangePolicy, "should be nil")
					s.Nil(podSecurityContext.RunAsGroup, "should be nil")
					s.Nil(podSecurityContext.RunAsNonRoot, "should be nil")
					s.Nil(podSecurityContext.RunAsUser, "should be nil")
					s.Nil(podSecurityContext.SeccompProfile, "should be nil")
					s.Nil(podSecurityContext.SELinuxOptions, "should be nil")
					s.Nil(podSecurityContext.SupplementalGroups, "should be nil")
					s.Nil(podSecurityContext.Sysctls, "should be nil")
					s.Nil(podSecurityContext.WindowsOptions, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateSecurityContext",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":             "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":          "nil",
				"delegatedOperatorDeployments.template.podSecurityContext.fsGroup":    "2000",
				"delegatedOperatorDeployments.template.podSecurityContext.runAsGroup": "3000",
				"delegatedOperatorDeployments.template.podSecurityContext.runAsUser":  "1000",
			},
			[]func(podSecurityContext *corev1.PodSecurityContext){
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Equal(int64(2000), *podSecurityContext.FSGroup, "fsGroup should be 2000")
					s.Equal(int64(3000), *podSecurityContext.RunAsGroup, "runAsGroup should be 3000")
					s.Equal(int64(1000), *podSecurityContext.RunAsUser, "runAsUser should be 1000")
				},
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Equal(int64(2000), *podSecurityContext.FSGroup, "fsGroup should be 2000")
					s.Equal(int64(3000), *podSecurityContext.RunAsGroup, "runAsGroup should be 3000")
					s.Equal(int64(1000), *podSecurityContext.RunAsUser, "runAsUser should be 1000")
				},
			},
		},
		{
			"overrideInstanceSecurityContext",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.podSecurityContext.fsGroup":       "2001",
				"delegatedOperatorDeployments.deployments.teamsDo.podSecurityContext.runAsGroup":    "3001",
				"delegatedOperatorDeployments.deployments.teamsDo.podSecurityContext.runAsUser":     "1001",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podSecurityContext.fsGroup":    "2002",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podSecurityContext.runAsGroup": "3002",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podSecurityContext.runAsUser":  "1002",
			},
			[]func(podSecurityContext *corev1.PodSecurityContext){
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Equal(int64(2001), *podSecurityContext.FSGroup, "fsGroup should be 2001")
					s.Equal(int64(3001), *podSecurityContext.RunAsGroup, "runAsGroup should be 3001")
					s.Equal(int64(1001), *podSecurityContext.RunAsUser, "runAsUser should be 1001")
				},
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Equal(int64(2002), *podSecurityContext.FSGroup, "fsGroup should be 2002")
					s.Equal(int64(3002), *podSecurityContext.RunAsGroup, "runAsGroup should be 3002")
					s.Equal(int64(1002), *podSecurityContext.RunAsUser, "runAsUser should be 1002")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceSecurityContext",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.podSecurityContext.fsGroup":    "2001",
				"delegatedOperatorDeployments.deployments.teamsDo.podSecurityContext.runAsGroup": "3001",
				"delegatedOperatorDeployments.deployments.teamsDo.podSecurityContext.runAsUser":  "1001",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.podSecurityContext.fsGroup": "2002",
				"delegatedOperatorDeployments.template.podSecurityContext.fsGroup":               "2000",
				"delegatedOperatorDeployments.template.podSecurityContext.runAsGroup":            "3000",
				"delegatedOperatorDeployments.template.podSecurityContext.runAsUser":             "1000",
			},
			[]func(podSecurityContext *corev1.PodSecurityContext){
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Equal(int64(2001), *podSecurityContext.FSGroup, "fsGroup should be 2001")
					s.Equal(int64(3001), *podSecurityContext.RunAsGroup, "runAsGroup should be 3001")
					s.Equal(int64(1001), *podSecurityContext.RunAsUser, "runAsUser should be 1001")
				},
				func(podSecurityContext *corev1.PodSecurityContext) {
					s.Equal(int64(2002), *podSecurityContext.FSGroup, "fsGroup should be 2002")
					s.Equal(int64(3000), *podSecurityContext.RunAsGroup, "runAsGroup should be 3000")
					s.Equal(int64(1000), *podSecurityContext.RunAsUser, "runAsUser should be 1000")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.SecurityContext)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestTemplateLabels() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []deploymentDelegatedOperatorInstanceTemplateLabelsExpected
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
				},
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
				},
			},
		},
		{
			"overrideBaseTemplateLabels",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.labels.myLabel":       "unruly",
			},
			[]deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
						"myLabel":                    "unruly",
					},
				},
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
						"myLabel":                    "unruly",
					},
				},
			},
		},
		{
			"overrideInstanceLabels",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.labels.teams-do-label":        "teams-do-label-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.labels.teams-do-two-label": "teams-do-two-label-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.labels.myLabel":            "very-ruly",
			},
			[]deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
						"teams-do-label":             "teams-do-label-value",
					},
				},
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
						"teams-do-two-label":         "teams-do-two-label-value",
						"myLabel":                    "very-ruly",
					},
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceLabels",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.labels.teams-do-label":        "teams-do-label-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.labels.teams-do-two-label": "teams-do-two-label-value",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.labels.myLabel":            "very-ruly",
				"delegatedOperatorDeployments.template.labels.myLabel":                          "unruly",
			},
			[]deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do",
						"app.kubernetes.io/instance": "fiftyone-test",
						"teams-do-label":             "teams-do-label-value",
						"myLabel":                    "unruly",
					},
				},
				deploymentDelegatedOperatorInstanceTemplateLabelsExpected{
					selectorMatch: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
					},
					templateMetadata: map[string]string{
						"app.kubernetes.io/name":     "teams-do-two",
						"app.kubernetes.io/instance": "fiftyone-test",
						"teams-do-two-label":         "teams-do-two-label-value",
						"myLabel":                    "very-ruly",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					for key, value := range testCase.expected[i].selectorMatch {

						foundValue := deployment.Spec.Selector.MatchLabels[key]
						s.Equal(value, foundValue, "Selector Labels should contain all set labels.")
					}

					for key, value := range testCase.expected[i].templateMetadata {

						foundValue := deployment.Spec.Template.ObjectMeta.Labels[key]
						s.Equal(value, foundValue, "Template Metadata Labels should contain all set labels.")
					}
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestServiceAccountName() {
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
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]string{"fiftyone-teams"},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]string{"fiftyone-teams", "fiftyone-teams"},
		},
		{
			"overrideServiceAccountName",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"serviceAccount.name": "test-service-account",
			},
			[]string{"test-service-account", "test-service-account"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					s.Equal(testCase.expected[i], deployment.Spec.Template.Spec.ServiceAccountName, "Service account name should be equal.")
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestTolerations() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(tolerations []corev1.Toleration)
	}{
		{
			"defaultValues",
			nil,
			[]func(tolerations []corev1.Toleration){
				func(tolerations []corev1.Toleration) {
					s.Empty(tolerations, "should not be set")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(tolerations []corev1.Toleration){
				func(tolerations []corev1.Toleration) {
					s.Nil(tolerations, "should be nil")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(tolerations []corev1.Toleration){
				func(tolerations []corev1.Toleration) {
					s.Nil(tolerations, "should be nil")
				},
				func(tolerations []corev1.Toleration) {
					s.Nil(tolerations, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateTolerations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":       "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":    "nil",
				"delegatedOperatorDeployments.template.tolerations[0].key":      "example-key",
				"delegatedOperatorDeployments.template.tolerations[0].operator": "Exists",
				"delegatedOperatorDeployments.template.tolerations[0].effect":   "NoSchedule",
			},
			[]func(tolerations []corev1.Toleration){
				func(tolerations []corev1.Toleration) {
					tolerationJSON := `[
                {
                  "key": "example-key",
                  "operator": "Exists",
                  "effect": "NoSchedule"
                }
              ]`
					var expectedTolerations []corev1.Toleration
					err := json.Unmarshal([]byte(tolerationJSON), &expectedTolerations)
					s.NoError(err)

					s.Len(tolerations, 1, "Should only have 1 toleration")
					s.Equal(expectedTolerations[0], tolerations[0], "Toleration should be equal")
				},
				func(tolerations []corev1.Toleration) {
					tolerationJSON := `[
                {
                  "key": "example-key",
                  "operator": "Exists",
                  "effect": "NoSchedule"
                }
              ]`
					var expectedTolerations []corev1.Toleration
					err := json.Unmarshal([]byte(tolerationJSON), &expectedTolerations)
					s.NoError(err)

					s.Len(tolerations, 1, "Should only have 1 toleration")
					s.Equal(expectedTolerations[0], tolerations[0], "Toleration should be equal")
				},
			},
		},
		{
			"overrideInstanceTolerations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.tolerations[0].key":      "example-key-teams-do",
				"delegatedOperatorDeployments.deployments.teamsDo.tolerations[0].operator": "Exists",
				"delegatedOperatorDeployments.deployments.teamsDo.tolerations[0].effect":   "PreferNoSchedule",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":               "nil",
			},
			[]func(tolerations []corev1.Toleration){
				func(tolerations []corev1.Toleration) {
					tolerationJSON := `[
                {
                  "key": "example-key-teams-do",
                  "operator": "Exists",
                  "effect": "PreferNoSchedule"
                }
              ]`
					var expectedTolerations []corev1.Toleration
					err := json.Unmarshal([]byte(tolerationJSON), &expectedTolerations)
					s.NoError(err)

					s.Len(tolerations, 1, "Should only have 1 toleration")
					s.Equal(expectedTolerations[0], tolerations[0], "Toleration should be equal")
				},
				func(tolerations []corev1.Toleration) {
					s.Nil(tolerations, "should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceTolerations",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.tolerations[0].key":      "example-key-teams-do",
				"delegatedOperatorDeployments.deployments.teamsDo.tolerations[0].operator": "Exists",
				"delegatedOperatorDeployments.deployments.teamsDo.tolerations[0].effect":   "PreferNoSchedule",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":               "nil",
				"delegatedOperatorDeployments.template.tolerations[0].key":                 "example-key",
				"delegatedOperatorDeployments.template.tolerations[0].operator":            "Exists",
				"delegatedOperatorDeployments.template.tolerations[0].effect":              "NoSchedule",
			},
			[]func(tolerations []corev1.Toleration){
				func(tolerations []corev1.Toleration) {
					tolerationJSON := `[
                {
                  "key": "example-key-teams-do",
                  "operator": "Exists",
                  "effect": "PreferNoSchedule"
                }
              ]`
					var expectedTolerations []corev1.Toleration
					err := json.Unmarshal([]byte(tolerationJSON), &expectedTolerations)
					s.NoError(err)

					s.Len(tolerations, 1, "Should only have 1 toleration")
					s.Equal(expectedTolerations[0], tolerations[0], "Toleration should be equal")
				},
				func(tolerations []corev1.Toleration) {
					tolerationJSON := `[
                {
                  "key": "example-key",
                  "operator": "Exists",
                  "effect": "NoSchedule"
                }
              ]`
					var expectedTolerations []corev1.Toleration
					err := json.Unmarshal([]byte(tolerationJSON), &expectedTolerations)
					s.NoError(err)

					s.Len(tolerations, 1, "Should only have 1 toleration")
					s.Equal(expectedTolerations[0], tolerations[0], "Toleration should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Tolerations)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestVolumes() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(volumes []corev1.Volume)
	}{
		{
			"defaultValues",
			nil,
			[]func(volumes []corev1.Volume){
				func(volumes []corev1.Volume) {
					s.Empty(volumes, "Volumes should be be set")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(volumes []corev1.Volume){
				func(volumes []corev1.Volume) {
					s.Nil(volumes, "Volumes should be nil")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(volumes []corev1.Volume){
				func(volumes []corev1.Volume) {
					s.Nil(volumes, "Volumes should be nil")
				},
				func(volumes []corev1.Volume) {
					s.Nil(volumes, "Volumes should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateVolumes",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":                          "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                       "nil",
				"delegatedOperatorDeployments.template.volumes[0].name":                            "template-test-volume1",
				"delegatedOperatorDeployments.template.volumes[0].hostPath.path":                   "/template-test-volume1",
				"delegatedOperatorDeployments.template.volumes[1].name":                            "template-pvc1",
				"delegatedOperatorDeployments.template.volumes[1].persistentVolumeClaim.claimName": "template-pvc1",
			},
			[]func(volumes []corev1.Volume){
				func(volumes []corev1.Volume) {
					expectedJSON := `[
          {
            "name": "template-test-volume1",
            "hostPath": {
              "path": "/template-test-volume1"
            }
          },
          {
            "name": "template-pvc1",
            "persistentVolumeClaim": {
              "claimName": "template-pvc1"
            }
          }
        ]`
					var expectedVolumes []corev1.Volume
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
					s.NoError(err)
					s.Equal(expectedVolumes, volumes, "Volumes should be equal")
				},
				func(volumes []corev1.Volume) {
					expectedJSON := `[
          {
            "name": "template-test-volume1",
            "hostPath": {
              "path": "/template-test-volume1"
            }
          },
          {
            "name": "template-pvc1",
            "persistentVolumeClaim": {
              "claimName": "template-pvc1"
            }
          }
        ]`
					var expectedVolumes []corev1.Volume
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
					s.NoError(err)
					s.Equal(expectedVolumes, volumes, "Volumes should be equal")
				},
			},
		},
		{
			"overrideInstanceVolumes",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.volumes[0].name":          "teams-do-test-volume1",
				"delegatedOperatorDeployments.deployments.teamsDo.volumes[0].hostPath.path": "/teams-do-test-volume1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                "nil",
			},
			[]func(volumes []corev1.Volume){
				func(volumes []corev1.Volume) {
					expectedJSON := `[
          {
            "name": "teams-do-test-volume1",
            "hostPath": {
              "path": "/teams-do-test-volume1"
            }
          }
        ]`
					var expectedVolumes []corev1.Volume
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
					s.NoError(err)
					s.Equal(expectedVolumes, volumes, "Volumes should be equal")
				},
				func(volumes []corev1.Volume) {
					s.Nil(volumes, "Volumes should be nil")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceVolumes",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.volumes[0].name":                 "teams-do-test-volume1",
				"delegatedOperatorDeployments.deployments.teamsDo.volumes[0].hostPath.path":        "/teams-do-test-volume1",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                       "nil",
				"delegatedOperatorDeployments.template.volumes[0].name":                            "template-test-volume1",
				"delegatedOperatorDeployments.template.volumes[0].hostPath.path":                   "/template-test-volume1",
				"delegatedOperatorDeployments.template.volumes[1].name":                            "template-pvc1",
				"delegatedOperatorDeployments.template.volumes[1].persistentVolumeClaim.claimName": "template-pvc1",
			},
			[]func(volumes []corev1.Volume){
				func(volumes []corev1.Volume) {
					expectedJSON := `[
          {
            "name": "teams-do-test-volume1",
            "hostPath": {
              "path": "/teams-do-test-volume1"
            }
          }
        ]`
					var expectedVolumes []corev1.Volume
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
					s.NoError(err)
					s.Equal(expectedVolumes, volumes, "Volumes should be equal")
				},
				func(volumes []corev1.Volume) {
					expectedJSON := `[
          {
            "name": "template-test-volume1",
            "hostPath": {
              "path": "/template-test-volume1"
            }
          },
          {
            "name": "template-pvc1",
            "persistentVolumeClaim": {
              "claimName": "template-pvc1"
            }
          }
        ]`
					var expectedVolumes []corev1.Volume
					err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
					s.NoError(err)
					s.Equal(expectedVolumes, volumes, "Volumes should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Volumes)
				}
			}

		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerLivenessProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					s.Empty(probe, "Liveness probe should not be set.")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateLivenessProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":         "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":      "nil",
				"delegatedOperatorDeployments.template.liveness.failureThreshold": "10",
				"delegatedOperatorDeployments.template.liveness.periodSeconds":    "10",
				"delegatedOperatorDeployments.template.liveness.timeoutSeconds":   "10",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 10,
          "periodSeconds": 10,
          "timeoutSeconds": 10
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 10,
          "periodSeconds": 10,
          "timeoutSeconds": 10
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
			},
		},
		{
			"overrideInstanceLivenessProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.liveness.failureThreshold":    "15",
				"delegatedOperatorDeployments.deployments.teamsDo.liveness.periodSeconds":       "20",
				"delegatedOperatorDeployments.deployments.teamsDo.liveness.timeoutSeconds":      "25",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.liveness.failureThreshold": "30",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.liveness.periodSeconds":    "35",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.liveness.timeoutSeconds":   "40",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 15,
          "periodSeconds": 20,
          "timeoutSeconds": 25
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 30,
          "periodSeconds": 35,
          "timeoutSeconds": 40
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceLivenessProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.liveness.failureThreshold":    "15",
				"delegatedOperatorDeployments.deployments.teamsDo.liveness.periodSeconds":       "20",
				"delegatedOperatorDeployments.deployments.teamsDo.liveness.timeoutSeconds":      "25",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.liveness.failureThreshold": "30",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.liveness.periodSeconds":    "35",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.liveness.timeoutSeconds":   "40",
				"delegatedOperatorDeployments.template.liveness.failureThreshold":               "10",
				"delegatedOperatorDeployments.template.liveness.periodSeconds":                  "10",
				"delegatedOperatorDeployments.template.liveness.timeoutSeconds":                 "10",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 15,
          "periodSeconds": 20,
          "timeoutSeconds": 25
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o liveness"
              ]
          },
          "failureThreshold": 30,
          "periodSeconds": 35,
          "timeoutSeconds": 40
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerReadinessProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					s.Empty(probe, "Readiness probe should not be set.")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateReadinessProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":          "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":       "nil",
				"delegatedOperatorDeployments.template.readiness.failureThreshold": "10",
				"delegatedOperatorDeployments.template.readiness.periodSeconds":    "10",
				"delegatedOperatorDeployments.template.readiness.timeoutSeconds":   "10",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 10,
          "periodSeconds": 10,
          "timeoutSeconds": 10
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 10,
          "periodSeconds": 10,
          "timeoutSeconds": 10
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
			},
		},
		{
			"overrideInstanceReadinessProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.readiness.failureThreshold":    "15",
				"delegatedOperatorDeployments.deployments.teamsDo.readiness.periodSeconds":       "20",
				"delegatedOperatorDeployments.deployments.teamsDo.readiness.timeoutSeconds":      "25",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.readiness.failureThreshold": "30",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.readiness.periodSeconds":    "35",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.readiness.timeoutSeconds":   "40",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 15,
          "periodSeconds": 20,
          "timeoutSeconds": 25
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 30,
          "periodSeconds": 35,
          "timeoutSeconds": 40
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceReadinessProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.readiness.failureThreshold":    "15",
				"delegatedOperatorDeployments.deployments.teamsDo.readiness.periodSeconds":       "20",
				"delegatedOperatorDeployments.deployments.teamsDo.readiness.timeoutSeconds":      "25",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.readiness.failureThreshold": "30",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.readiness.periodSeconds":    "35",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.readiness.timeoutSeconds":   "40",
				"delegatedOperatorDeployments.template.readiness.failureThreshold":               "10",
				"delegatedOperatorDeployments.template.readiness.periodSeconds":                  "10",
				"delegatedOperatorDeployments.template.readiness.timeoutSeconds":                 "10",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 15,
          "periodSeconds": 20,
          "timeoutSeconds": 25
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o readiness"
              ]
          },
          "failureThreshold": 30,
          "periodSeconds": 35,
          "timeoutSeconds": 40
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerStartupProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					s.Empty(probe, "Startup probe should not be set.")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "startup Probes should be equal")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "startup Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 5,
          "periodSeconds": 30,
          "timeoutSeconds": 30
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "startup Probes should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateStartupProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":        "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":     "nil",
				"delegatedOperatorDeployments.template.startup.failureThreshold": "10",
				"delegatedOperatorDeployments.template.startup.periodSeconds":    "10",
				"delegatedOperatorDeployments.template.startup.timeoutSeconds":   "10",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 10,
          "periodSeconds": 10,
          "timeoutSeconds": 10
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "startup Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 10,
          "periodSeconds": 10,
          "timeoutSeconds": 10
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "startup Probes should be equal")
				},
			},
		},
		{
			"overrideInstanceStartupProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.startup.failureThreshold":    "15",
				"delegatedOperatorDeployments.deployments.teamsDo.startup.periodSeconds":       "20",
				"delegatedOperatorDeployments.deployments.teamsDo.startup.timeoutSeconds":      "25",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.startup.failureThreshold": "30",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.startup.periodSeconds":    "35",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.startup.timeoutSeconds":   "40",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 15,
          "periodSeconds": 20,
          "timeoutSeconds": 25
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Startup Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 30,
          "periodSeconds": 35,
          "timeoutSeconds": 40
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Startup Probes should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateAndInstanceStartupProbe",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.startup.failureThreshold":    "15",
				"delegatedOperatorDeployments.deployments.teamsDo.startup.periodSeconds":       "20",
				"delegatedOperatorDeployments.deployments.teamsDo.startup.timeoutSeconds":      "25",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.startup.failureThreshold": "30",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.startup.periodSeconds":    "35",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.startup.timeoutSeconds":   "40",
				"delegatedOperatorDeployments.template.startup.failureThreshold":               "10",
				"delegatedOperatorDeployments.template.startup.periodSeconds":                  "10",
				"delegatedOperatorDeployments.template.startup.timeoutSeconds":                 "10",
			},
			[]func(probe *corev1.Probe){
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 15,
          "periodSeconds": 20,
          "timeoutSeconds": 25
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Startup Probes should be equal")
				},
				func(probe *corev1.Probe) {
					expectedProbeJSON := `{
          "exec": {
              "command": [
                "sh",
                "-c",
                "fiftyone delegated list --limit 1 -o startup"
              ]
          },
          "failureThreshold": 30,
          "periodSeconds": 35,
          "timeoutSeconds": 40
        }`
					var expectedProbe *corev1.Probe
					err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
					s.NoError(err)
					s.Equal(expectedProbe, probe, "Startup Probes should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].StartupProbe)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerCmdArgs() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(args []string)
	}{
		{
			"defaultValues",
			nil,
			[]func(args []string){
				func(args []string) {
					s.Empty(args, "Args should not be set.")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(args []string){
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do",
						"-d",
						"Long running operations delegated to teams-do",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(args []string){
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do",
						"-d",
						"Long running operations delegated to teams-do",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do-two",
						"-d",
						"Long running operations delegated to teams-do-two",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateDescription", // This should still show the defaults
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.description":          "Delegated Operator",
			},
			[]func(args []string){
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do",
						"-d",
						"Long running operations delegated to teams-do",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do-two",
						"-d",
						"Long running operations delegated to teams-do-two",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
			},
		},
		{
			"overrideInstanceDescription",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.description": "Used for non-gpu workloads",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":   "nil",
			},
			[]func(args []string){
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do",
						"-d",
						"Used for non-gpu workloads",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do-two",
						"-d",
						"Long running operations delegated to teams-do-two",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceDescription",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.description": "Used for non-gpu workloads",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":   "nil",
				"delegatedOperatorDeployments.template.description":            "Delegated Operator",
			},
			[]func(args []string){
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do",
						"-d",
						"Used for non-gpu workloads",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
				func(args []string) {
					expectedArgs := []string{
						"delegated",
						"launch",
						"-t",
						"remote",
						"-n",
						"teams-do-two",
						"-d",
						"Long running operations delegated to teams-do-two",
					}
					s.Equal(expectedArgs, args, "Args should be equal")
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Template.Spec.Containers[0].Args)
				}
			}
		})
	}
}

func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestDeploymentUpdateStrategy() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []func(deploymentStrategy appsv1.DeploymentStrategy)
	}{
		{
			"defaultValues",
			nil,
			[]func(deploymentStrategy appsv1.DeploymentStrategy){
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					s.Empty(deploymentStrategy.Type, "Type should be be empty")
				},
			},
		},
		{
			"defaultValuesDOEnabled",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused": "nil",
			},
			[]func(deploymentStrategy appsv1.DeploymentStrategy){
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "RollingUpdate"
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
			},
		},
		{
			"defaultValuesMultipleInstances",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
			},
			[]func(deploymentStrategy appsv1.DeploymentStrategy){
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "RollingUpdate"
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "RollingUpdate"
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateStrategyType",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":    "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused": "nil",
				"delegatedOperatorDeployments.template.updateStrategy.type":  "Recreate",
			},
			[]func(deploymentStrategy appsv1.DeploymentStrategy){
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "Recreate"
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "Recreate"
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateStrategyRollingUpdate",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":                           "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.unused":                        "nil",
				"delegatedOperatorDeployments.template.updateStrategy.type":                         "RollingUpdate",
				"delegatedOperatorDeployments.template.updateStrategy.rollingUpdate.maxUnavailable": "5",
			},
			[]func(deploymentStrategy appsv1.DeploymentStrategy){
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "RollingUpdate",
                "rollingUpdate": {
                    "maxUnavailable": 5
                }
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "RollingUpdate",
                "rollingUpdate": {
                    "maxUnavailable": 5
                }
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
			},
		},
		{
			"overrideBaseTemplateInstanceStrategy",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDo.unused":                           "nil",
				"delegatedOperatorDeployments.deployments.teamsDoTwo.updateStrategy.type":           "Recreate",
				"delegatedOperatorDeployments.template.updateStrategy.type":                         "RollingUpdate",
				"delegatedOperatorDeployments.template.updateStrategy.rollingUpdate.maxUnavailable": "5",
			},
			[]func(deploymentStrategy appsv1.DeploymentStrategy){
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					expectedJSON := `{
                "type": "RollingUpdate",
                "rollingUpdate": {
                    "maxUnavailable": 5
                }
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
				},
				func(deploymentStrategy appsv1.DeploymentStrategy) {
					// This provides the same behavior as our other "dict-wise" merges. While
					// the customer can mix styles here, this is consistent with how the other
					// parameters are handles. Therefore, this is intentional and is a design choice.
					expectedJSON := `{
                "type": "Recreate",
                "rollingUpdate": {
                    "maxUnavailable": 5
                }
            }`
					var expectedDeploymentStrategy appsv1.DeploymentStrategy
					err := json.Unmarshal([]byte(expectedJSON), &expectedDeploymentStrategy)
					s.NoError(err)
					s.Equal(expectedDeploymentStrategy, deploymentStrategy, "Deployment strategies should be equal")
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

			if testCase.values == nil {

				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected[0](deployment.Spec.Strategy)
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				// https://github.com/gruntwork-io/terratest/issues/586#issuecomment-848542351
				allRange := strings.Split(output, "---")
				s.Equal(len(allRange[1:]), len(testCase.expected), "Number of delegated operator instances should match expected number")

				for i, rawOutput := range allRange[1:] {
					var deployment appsv1.Deployment
					helm.UnmarshalK8SYaml(subT, rawOutput, &deployment)

					testCase.expected[i](deployment.Spec.Strategy)
				}
			}
		})
	}
}
