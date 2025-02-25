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
					"app.teams.operator/instance":  "teams-do",
				},
			},
		},
		{
			"multipleInstances",
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
					"app.teams.operator/instance":  "teams-do",
				},
				map[string]string{
					"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
					"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "teams-do-two",
					"app.kubernetes.io/instance":   "fiftyone-test",
					"app.teams.operator/instance":  "teams-do-two",
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
					"app.teams.operator/instance":  "teams-do-new-name",
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

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerEnv() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(envVars []corev1.EnvVar)
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			func(envVars []corev1.EnvVar) {
// 				expectedEnvVarJSON := `[]`
// 				var expectedEnvVars []corev1.EnvVar
// 				err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
// 				s.NoError(err)
// 				s.Equal(expectedEnvVars, envVars, "Envs should be equal")
// 			},
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(envVars []corev1.EnvVar) {
// 				expectedEnvVarJSON := `[
//           {
//             "name": "API_URL",
//             "value": "http://teams-api:80"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_ADMIN",
//             "value": "false"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_NAME",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "fiftyoneDatabaseName"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DATABASE_URI",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "mongodbConnectionString"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_ENCRYPTION_KEY",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "encryptionKey"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
//             "value": ""
//           },
//           {
//             "name": "FIFTYONE_INTERNAL_SERVICE",
//             "value": "true"
//           },
//           {
//             "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
//             "value": "-1"
//           }
//         ]`
// 				var expectedEnvVars []corev1.EnvVar
// 				err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
// 				s.NoError(err)
// 				s.Equal(expectedEnvVars, envVars, "Envs should be equal")
// 			},
// 		},
// 		{
// 			"overrideEnv",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                                       "true",
// 				"delegatedOperatorExecutorSettings.env.TEST_KEY":                                  "TEST_VALUE",
// 				"delegatedOperatorExecutorSettings.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretName": "an-existing-secret", // pragma: allowlist secret
// 				"delegatedOperatorExecutorSettings.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretKey":  "anExistingKey",      // pragma: allowlist secret
// 			},
// 			func(envVars []corev1.EnvVar) {
// 				expectedEnvVarJSON := `[
//           {
//             "name": "API_URL",
//             "value": "http://teams-api:80"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_ADMIN",
//             "value": "false"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_NAME",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "fiftyoneDatabaseName"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DATABASE_URI",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "mongodbConnectionString"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_ENCRYPTION_KEY",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "encryptionKey"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
//             "value": ""
//           },
//           {
//             "name": "FIFTYONE_INTERNAL_SERVICE",
//             "value": "true"
//           },
//           {
//             "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
//             "value": "-1"
//           },
//           {
//             "name": "TEST_KEY",
//             "value": "TEST_VALUE"
//           },
//           {
//             "name": "AN_ADDITIONAL_SECRET_ENV",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "an-existing-secret",
//                 "key": "anExistingKey"
//               }
//             }
//           }
//         ]`
// 				var expectedEnvVars []corev1.EnvVar
// 				err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
// 				s.NoError(err)
// 				s.Equal(expectedEnvVars, envVars, "Envs should be equal")
// 			},
// 		},
// 		{
// 			"overrideSecretName",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 				"secret.name": "override-secret-name",
// 			},
// 			func(envVars []corev1.EnvVar) {
// 				expectedEnvVarJSON := `[
//           {
//             "name": "API_URL",
//             "value": "http://teams-api:80"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_ADMIN",
//             "value": "false"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_NAME",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "override-secret-name",
//                 "key": "fiftyoneDatabaseName"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DATABASE_URI",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "override-secret-name",
//                 "key": "mongodbConnectionString"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_ENCRYPTION_KEY",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "override-secret-name",
//                 "key": "encryptionKey"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
//             "value": ""
//           },
//           {
//             "name": "FIFTYONE_INTERNAL_SERVICE",
//             "value": "true"
//           },
//           {
//             "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
//             "value": "-1"
//           }
//         ]`
// 				var expectedEnvVars []corev1.EnvVar
// 				err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
// 				s.NoError(err)
// 				s.Equal(expectedEnvVars, envVars, "Envs should be equal")
// 			},
// 		},
// 		{
// 			"overrideApiServiceNameAndPort",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 				"apiSettings.service.name":                  "teams-api-override",
// 				"apiSettings.service.port":                  "8000",
// 			},
// 			func(envVars []corev1.EnvVar) {
// 				expectedEnvVarJSON := `[
//           {
//             "name": "API_URL",
//             "value": "http://teams-api-override:8000"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_ADMIN",
//             "value": "false"
//           },
//           {
//             "name": "FIFTYONE_DATABASE_NAME",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "fiftyoneDatabaseName"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DATABASE_URI",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "mongodbConnectionString"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_ENCRYPTION_KEY",
//             "valueFrom": {
//               "secretKeyRef": {
//                 "name": "fiftyone-teams-secrets",
//                 "key": "encryptionKey"
//               }
//             }
//           },
//           {
//             "name": "FIFTYONE_DELEGATED_OPERATION_LOG_PATH",
//             "value": ""
//           },
//           {
//             "name": "FIFTYONE_INTERNAL_SERVICE",
//             "value": "true"
//           },
//           {
//             "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
//             "value": "-1"
//           }
//         ]`
// 				var expectedEnvVars []corev1.EnvVar
// 				err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
// 				s.NoError(err)
// 				s.Equal(expectedEnvVars, envVars, "Envs should be equal")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				testCase.expected(deployment.Spec.Template.Spec.Containers[0].Env)
// 			}
// 		})
// 	}
// }

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

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerVolumeMounts() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(volumeMounts []corev1.VolumeMount)
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			func(volumeMounts []corev1.VolumeMount) {
// 				s.Empty(volumeMounts, "VolumeMounts should not be set")
// 			},
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(volumeMounts []corev1.VolumeMount) {
// 				s.Nil(volumeMounts, "VolumeMounts should be nil")
// 			},
// 		},
// 		{
// 			"overrideVolumeMountsSingle",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                   "true",
// 				"delegatedOperatorExecutorSettings.volumeMounts[0].mountPath": "/test-data-volume",
// 				"delegatedOperatorExecutorSettings.volumeMounts[0].name":      "test-volume",
// 			},
// 			func(volumeMounts []corev1.VolumeMount) {
// 				expectedJSON := `[
//           {
//             "mountPath": "/test-data-volume",
//             "name": "test-volume"
//           }
//         ]`
// 				var expectedVolumeMounts []corev1.VolumeMount
// 				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
// 				s.NoError(err)
// 				s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
// 			},
// 		},
// 		{
// 			"overrideVolumeMountsMultiple",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                   "true",
// 				"delegatedOperatorExecutorSettings.volumeMounts[0].mountPath": "/test-data-volume1",
// 				"delegatedOperatorExecutorSettings.volumeMounts[0].name":      "test-volume1",
// 				"delegatedOperatorExecutorSettings.volumeMounts[1].mountPath": "/test-data-volume2",
// 				"delegatedOperatorExecutorSettings.volumeMounts[1].name":      "test-volume2",
// 			},
// 			func(volumeMounts []corev1.VolumeMount) {
// 				expectedJSON := `[
//           {
//             "mountPath": "/test-data-volume1",
//             "name": "test-volume1"
//           },
//           {
//             "mountPath": "/test-data-volume2",
//             "name": "test-volume2"
//           }
//         ]`
// 				var expectedVolumeMounts []corev1.VolumeMount
// 				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
// 				s.NoError(err)
// 				s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				testCase.expected(deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
// 			}

// 		})
// 	}
// }

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
					fmt.Printf("%v\n", testCase.expected)

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

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestNodeSelector() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected map[string]string
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			nil,
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			nil,
// 		},
// 		{
// 			"overrideNodeSelector",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":               "true",
// 				"delegatedOperatorExecutorSettings.nodeSelector.disktype": "ssd",
// 			},
// 			map[string]string{
// 				"disktype": "ssd",
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				for key, value := range testCase.expected {
// 					foundValue := deployment.Spec.Template.Spec.NodeSelector[key]
// 					s.Equal(value, foundValue, "NodeSelector should contain all set labels.")
// 				}
// 			}
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestPodAnnotations() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected map[string]string
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			nil,
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			nil,
// 		},
// 		{
// 			"overridePodAnnotations",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                     "true",
// 				"delegatedOperatorExecutorSettings.podAnnotations.annotation-1": "annotation-1-value",
// 			},
// 			map[string]string{
// 				"annotation-1": "annotation-1-value",
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				if testCase.expected == nil {
// 					s.Nil(deployment.Spec.Template.ObjectMeta.Annotations, "Annotations should be nil")
// 				} else {
// 					for key, value := range testCase.expected {
// 						foundValue := deployment.Spec.Template.ObjectMeta.Annotations[key]
// 						s.Equal(value, foundValue, "Annotations should contain all set annotations.")
// 					}
// 				}
// 			}
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestPodSecurityContext() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(podSecurityContext *corev1.PodSecurityContext)
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			func(podSecurityContext *corev1.PodSecurityContext) {
// 				s.Empty(podSecurityContext.FSGroup, "should not be set")
// 			},
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(podSecurityContext *corev1.PodSecurityContext) {
// 				s.Nil(podSecurityContext.FSGroup, "should be nil")
// 				s.Nil(podSecurityContext.FSGroupChangePolicy, "should be nil")
// 				s.Nil(podSecurityContext.RunAsGroup, "should be nil")
// 				s.Nil(podSecurityContext.RunAsNonRoot, "should be nil")
// 				s.Nil(podSecurityContext.RunAsUser, "should be nil")
// 				s.Nil(podSecurityContext.SeccompProfile, "should be nil")
// 				s.Nil(podSecurityContext.SELinuxOptions, "should be nil")
// 				s.Nil(podSecurityContext.SupplementalGroups, "should be nil")
// 				s.Nil(podSecurityContext.Sysctls, "should be nil")
// 				s.Nil(podSecurityContext.WindowsOptions, "should be nil")
// 			},
// 		},
// 		{
// 			"overridePodSecurityContext",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                       "true",
// 				"delegatedOperatorExecutorSettings.podSecurityContext.fsGroup":    "2000",
// 				"delegatedOperatorExecutorSettings.podSecurityContext.runAsGroup": "3000",
// 				"delegatedOperatorExecutorSettings.podSecurityContext.runAsUser":  "1000",
// 			},
// 			func(podSecurityContext *corev1.PodSecurityContext) {
// 				s.Equal(int64(2000), *podSecurityContext.FSGroup, "fsGroup should be 2000")
// 				s.Equal(int64(3000), *podSecurityContext.RunAsGroup, "runAsGroup should be 3000")
// 				s.Equal(int64(1000), *podSecurityContext.RunAsUser, "runAsUser should be 1000")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				testCase.expected(deployment.Spec.Template.Spec.SecurityContext)
// 			}
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestTemplateLabels() {
// 	testCases := []struct {
// 		name                     string
// 		values                   map[string]string
// 		selectorMatchExpected    map[string]string
// 		templateMetadataExpected map[string]string
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			nil,
// 			nil,
// 		},
// 		{
// 			"addTemplateMetadataLabels",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":        "true",
// 				"delegatedOperatorExecutorSettings.labels.myLabel": "unruly",
// 			},
// 			map[string]string{
// 				"app.kubernetes.io/name":     "teams-do",
// 				"app.kubernetes.io/instance": "fiftyone-test",
// 			},
// 			map[string]string{
// 				"app.kubernetes.io/name":     "teams-do",
// 				"app.kubernetes.io/instance": "fiftyone-test",
// 				"myLabel":                    "unruly",
// 			},
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			map[string]string{
// 				"app.kubernetes.io/name":     "teams-do",
// 				"app.kubernetes.io/instance": "fiftyone-test",
// 			},
// 			map[string]string{
// 				"app.kubernetes.io/name":     "teams-do",
// 				"app.kubernetes.io/instance": "fiftyone-test",
// 			},
// 		},
// 		{
// 			"overrideSelectorMatchLabels",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 				"delegatedOperatorExecutorSettings.name":    "test-service-name",
// 			},
// 			map[string]string{
// 				"app.kubernetes.io/name":     "test-service-name",
// 				"app.kubernetes.io/instance": "fiftyone-test",
// 			},
// 			map[string]string{
// 				"app.kubernetes.io/name":     "test-service-name",
// 				"app.kubernetes.io/instance": "fiftyone-test",
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				for key, value := range testCase.selectorMatchExpected {

// 					foundValue := deployment.Spec.Selector.MatchLabels[key]
// 					s.Equal(value, foundValue, "Selector Labels should contain all set labels.")
// 				}

// 				for key, value := range testCase.templateMetadataExpected {

// 					foundValue := deployment.Spec.Template.ObjectMeta.Labels[key]
// 					s.Equal(value, foundValue, "Template Metadata Labels should contain all set labels.")
// 				}
// 			}
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestServiceAccountName() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected string
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			"",
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			"fiftyone-teams",
// 		},
// 		{
// 			"overrideServiceAccountName",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 				"serviceAccount.name":                       "test-service-account",
// 			},
// 			"test-service-account",
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Equal(testCase.expected, deployment.Spec.Template.Spec.ServiceAccountName, "Service account name should be equal.")
// 			}
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestTolerations() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(tolerations []corev1.Toleration)
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			func(tolerations []corev1.Toleration) {
// 				s.Empty(tolerations, "should not be set")
// 			},
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(tolerations []corev1.Toleration) {
// 				s.Nil(tolerations, "should be nil")
// 			},
// 		},
// 		{
// 			"overrideTolerations",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                 "true",
// 				"delegatedOperatorExecutorSettings.tolerations[0].key":      "example-key",
// 				"delegatedOperatorExecutorSettings.tolerations[0].operator": "Exists",
// 				"delegatedOperatorExecutorSettings.tolerations[0].effect":   "NoSchedule",
// 			},
// 			func(tolerations []corev1.Toleration) {
// 				tolerationJSON := `[
//           {
//             "key": "example-key",
//             "operator": "Exists",
//             "effect": "NoSchedule"
//           }
//         ]`
// 				var expectedTolerations []corev1.Toleration
// 				err := json.Unmarshal([]byte(tolerationJSON), &expectedTolerations)
// 				s.NoError(err)

// 				s.Len(tolerations, 1, "Should only have 1 toleration")
// 				s.Equal(expectedTolerations[0], tolerations[0], "Toleration should be equal")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				testCase.expected(deployment.Spec.Template.Spec.Tolerations)
// 			}
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestVolumes() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(volumes []corev1.Volume)
// 	}{
// 		{
// 			"defaultValues",
// 			nil,
// 			func(volumes []corev1.Volume) {
// 				s.Empty(volumes, "Volumes should be be set")
// 			},
// 		},
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(volumes []corev1.Volume) {
// 				s.Nil(volumes, "Volumes should be nil")
// 			},
// 		},
// 		{
// 			"overrideVolumesSingle",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                  "true",
// 				"delegatedOperatorExecutorSettings.volumes[0].name":          "test-volume",
// 				"delegatedOperatorExecutorSettings.volumes[0].hostPath.path": "/test-volume",
// 			},
// 			func(volumes []corev1.Volume) {
// 				expectedJSON := `[
//           {
//             "name": "test-volume",
//             "hostPath": {
//               "path": "/test-volume"
//             }
//           }
//         ]`
// 				var expectedVolumes []corev1.Volume
// 				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
// 				s.NoError(err)
// 				s.Equal(expectedVolumes, volumes, "Volumes should be equal")
// 			},
// 		},
// 		{
// 			"overrideVolumesMultiple",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                                    "true",
// 				"delegatedOperatorExecutorSettings.volumes[0].name":                            "test-volume1",
// 				"delegatedOperatorExecutorSettings.volumes[0].hostPath.path":                   "/test-volume1",
// 				"delegatedOperatorExecutorSettings.volumes[1].name":                            "pvc1",
// 				"delegatedOperatorExecutorSettings.volumes[1].persistentVolumeClaim.claimName": "pvc1",
// 			},
// 			func(volumes []corev1.Volume) {
// 				expectedJSON := `[
//           {
//             "name": "test-volume1",
//             "hostPath": {
//               "path": "/test-volume1"
//             }
//           },
//           {
//             "name": "pvc1",
//             "persistentVolumeClaim": {
//               "claimName": "pvc1"
//             }
//           }
//         ]`
// 				var expectedVolumes []corev1.Volume
// 				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
// 				s.NoError(err)
// 				s.Equal(expectedVolumes, volumes, "Volumes should be equal")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			// when vars are set outside of the if statement, they aren't accessible from within the conditional
// 			if testCase.values == nil {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

// 				s.ErrorContains(err, "could not find template templates/delegated-operator-instance-deployment.yaml in chart")
// 				var deployment appsv1.Deployment

// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				s.Nil(deployment.Spec.Template.Spec.Containers)
// 			} else {
// 				options := &helm.Options{SetValues: testCase.values}
// 				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 				var deployment appsv1.Deployment
// 				helm.UnmarshalK8SYaml(subT, output, &deployment)

// 				testCase.expected(deployment.Spec.Template.Spec.Volumes)
// 			}

// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerLivenessProbe() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(probe *corev1.Probe)
// 	}{
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(probe *corev1.Probe) {
// 				expectedProbeJSON := `{
//           "exec": {
//               "command": [
//                 "sh",
//                 "-c",
//                 "fiftyone delegated list --limit 1 -o liveness"
//               ]
//           },
//           "failureThreshold": 5,
//           "periodSeconds": 30,
//           "timeoutSeconds": 30
//         }`
// 				var expectedProbe *corev1.Probe
// 				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
// 				s.NoError(err)
// 				s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
// 			},
// 		},
// 		{
// 			"overrideServiceStartupFailureThresholdAndPeriodSecondsAndShortName",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                   "true",
// 				"delegatedOperatorExecutorSettings.liveness.failureThreshold": "10",
// 				"delegatedOperatorExecutorSettings.liveness.periodSeconds":    "10",
// 				"delegatedOperatorExecutorSettings.liveness.timeoutSeconds":   "10",
// 			},
// 			func(probe *corev1.Probe) {
// 				expectedProbeJSON := `{
//           "exec": {
//               "command": [
//                 "sh",
//                 "-c",
//                 "fiftyone delegated list --limit 1 -o liveness"
//               ]
//           },
//           "failureThreshold": 10,
//           "periodSeconds": 10,
//           "timeoutSeconds": 10
//         }`
// 				var expectedProbe *corev1.Probe
// 				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
// 				s.NoError(err)
// 				s.Equal(expectedProbe, probe, "Startup Probes should be equal")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			options := &helm.Options{SetValues: testCase.values}
// 			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 			var deployment appsv1.Deployment
// 			helm.UnmarshalK8SYaml(subT, output, &deployment)

// 			testCase.expected(deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerReadinessProbe() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(probe *corev1.Probe)
// 	}{
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(probe *corev1.Probe) {
// 				expectedProbeJSON := `{
//           "exec": {
//               "command": [
//                 "sh",
//                 "-c",
//                 "fiftyone delegated list --limit 1 -o readiness"
//               ]
//           },
//           "failureThreshold": 5,
//           "periodSeconds": 30,
//           "timeoutSeconds": 30
//         }`
// 				var expectedProbe *corev1.Probe
// 				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
// 				s.NoError(err)
// 				s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
// 			},
// 		},
// 		{
// 			"overrideServiceStartupFailureThresholdAndPeriodSecondsAndShortName",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                    "true",
// 				"delegatedOperatorExecutorSettings.readiness.failureThreshold": "10",
// 				"delegatedOperatorExecutorSettings.readiness.periodSeconds":    "10",
// 				"delegatedOperatorExecutorSettings.readiness.timeoutSeconds":   "10",
// 			},
// 			func(probe *corev1.Probe) {
// 				expectedProbeJSON := `{
//           "exec": {
//               "command": [
//                 "sh",
//                 "-c",
//                 "fiftyone delegated list --limit 1 -o readiness"
//               ]
//           },
//           "failureThreshold": 10,
//           "periodSeconds": 10,
//           "timeoutSeconds": 10
//         }`
// 				var expectedProbe *corev1.Probe
// 				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
// 				s.NoError(err)
// 				s.Equal(expectedProbe, probe, "Startup Probes should be equal")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			options := &helm.Options{SetValues: testCase.values}
// 			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 			var deployment appsv1.Deployment
// 			helm.UnmarshalK8SYaml(subT, output, &deployment)

// 			testCase.expected(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
// 		})
// 	}
// }

// func (s *deploymentDelegatedOperatorInstanceTemplateTest) TestContainerStartupProbe() {
// 	testCases := []struct {
// 		name     string
// 		values   map[string]string
// 		expected func(probe *corev1.Probe)
// 	}{
// 		{
// 			"defaultValuesDOEnabled",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled": "true",
// 			},
// 			func(probe *corev1.Probe) {
// 				expectedProbeJSON := `{
//           "exec": {
//               "command": [
//                 "sh",
//                 "-c",
//                 "fiftyone delegated list --limit 1 -o startup"
//               ]
//           },
//           "failureThreshold": 5,
//           "periodSeconds": 30,
//           "timeoutSeconds": 30
//         }`
// 				var expectedProbe *corev1.Probe
// 				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
// 				s.NoError(err)
// 				s.Equal(expectedProbe, probe, "Startup Probes should be equal")
// 			},
// 		},
// 		{
// 			"overrideServiceStartupFailureThresholdAndPeriodSecondsAndShortName",
// 			map[string]string{
// 				"delegatedOperatorExecutorSettings.enabled":                  "true",
// 				"delegatedOperatorExecutorSettings.startup.failureThreshold": "10",
// 				"delegatedOperatorExecutorSettings.startup.periodSeconds":    "10",
// 				"delegatedOperatorExecutorSettings.startup.timeoutSeconds":   "10",
// 			},
// 			func(probe *corev1.Probe) {
// 				expectedProbeJSON := `{
//           "exec": {
//               "command": [
//                 "sh",
//                 "-c",
//                 "fiftyone delegated list --limit 1 -o startup"
//               ]
//           },
//           "failureThreshold": 10,
//           "periodSeconds": 10,
//           "timeoutSeconds": 10
//         }`
// 				var expectedProbe *corev1.Probe
// 				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
// 				s.NoError(err)
// 				s.Equal(expectedProbe, probe, "Startup Probes should be equal")
// 			},
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase

// 		s.Run(testCase.name, func() {
// 			subT := s.T()
// 			subT.Parallel()

// 			options := &helm.Options{SetValues: testCase.values}
// 			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

// 			var deployment appsv1.Deployment
// 			helm.UnmarshalK8SYaml(subT, output, &deployment)

// 			testCase.expected(deployment.Spec.Template.Spec.Containers[0].StartupProbe)
// 		})
// 	}
// }
