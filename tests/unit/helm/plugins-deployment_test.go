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

type deploymentPluginsTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestDeploymentPluginsTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &deploymentPluginsTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/plugins-deployment.yaml"},
	})
}

func (s *deploymentPluginsTemplateTest) TestMetadataLabels() {
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

			if testCase.expected == nil {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.ObjectMeta.Labels, "Labels should be nil")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				for key, value := range testCase.expected {
					foundValue := deployment.ObjectMeta.Labels[key]
					s.Equal(value, foundValue, "Labels should contain all set labels.")
				}
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestMetadataName() {
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
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == "" {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Empty(deployment.ObjectMeta.Name, "Metadata name should be nil")

			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, deployment.ObjectMeta.Name, "Deployment name should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestMetadataNamespace() {
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
				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Empty(deployment.ObjectMeta.Namespace, "Metadata namespace should be nil")

			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, deployment.ObjectMeta.Namespace, "Namespace name should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestReplicas() {
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
			2,
		},
		{
			"overrideReplicaCount",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.replicaCount": "3",
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
				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Empty(&deployment.Spec.Replicas, "Replica count should be nil.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, *deployment.Spec.Replicas, "Replica count should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestTopologySpreadConstraints() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(constraint []corev1.TopologySpreadConstraint)
	}{
		{
			"defaultValues",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(constraint []corev1.TopologySpreadConstraint) {
				var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
				s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
			},
		},
		{
			"overrideTopologySpreadConstraintsRequiredValues",
			map[string]string{
				"pluginsSettings.enabled":                                        "true",
				"pluginsSettings.topologySpreadConstraints[0].maxSkew":           "1",
				"pluginsSettings.topologySpreadConstraints[0].topologyKey":       "kubernetes.io/hostname",
				"pluginsSettings.topologySpreadConstraints[0].whenUnsatisfiable": "DoNotSchedule",
			},
			func(constraint []corev1.TopologySpreadConstraint) {
				var expectedTopologySpreadConstraint []corev1.TopologySpreadConstraint
				expectedTopologySpreadConstraintJSON := `[
                    {
                      "maxSkew": 1,
                      "topologyKey": "kubernetes.io/hostname",
                      "whenUnsatisfiable": "DoNotSchedule",
                      "labelSelector": {
                          "matchLabels": {
                            "app.kubernetes.io/name": "teams-plugins",
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
		{
			"overrideTopologySpreadConstraintsOptionalValues",
			map[string]string{
				"pluginsSettings.enabled": "true",
				"pluginsSettings.topologySpreadConstraints[0].matchLabelKeys[0]":  "pod-template-hash",
				"pluginsSettings.topologySpreadConstraints[0].maxSkew":            "1",
				"pluginsSettings.topologySpreadConstraints[0].minDomains":         "1",
				"pluginsSettings.topologySpreadConstraints[0].nodeAffinityPolicy": "Honor",
				"pluginsSettings.topologySpreadConstraints[0].nodeTaintsPolicy":   "Honor",
				"pluginsSettings.topologySpreadConstraints[0].topologyKey":        "kubernetes.io/hostname",
				"pluginsSettings.topologySpreadConstraints[0].whenUnsatisfiable":  "DoNotSchedule",
				"pluginsSettings.topologySpreadConstraints[1].matchLabelKeys[0]":  "pod-template-hash",
				"pluginsSettings.topologySpreadConstraints[1].maxSkew":            "2",
				"pluginsSettings.topologySpreadConstraints[1].minDomains":         "2",
				"pluginsSettings.topologySpreadConstraints[1].nodeAffinityPolicy": "Ignore",
				"pluginsSettings.topologySpreadConstraints[1].nodeTaintsPolicy":   "Ignore",
				"pluginsSettings.topologySpreadConstraints[1].topologyKey":        "kubernetes.io/region",
				"pluginsSettings.topologySpreadConstraints[1].whenUnsatisfiable":  "ScheduleAnyway",
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
                            "app.kubernetes.io/name": "teams-plugins",
                            "app.kubernetes.io/instance": "fiftyone-test"
                        }
                      }
                    },
                    {
                      "matchLabelKeys": [
                          "pod-template-hash"
                      ],
                      "maxSkew": 2,
                      "minDomains": 2,
                      "nodeAffinityPolicy": "Ignore",
                      "nodeTaintsPolicy": "Ignore",
                      "topologyKey": "kubernetes.io/region",
                      "whenUnsatisfiable": "ScheduleAnyway",
                      "labelSelector": {
                          "matchLabels": {
                            "app.kubernetes.io/name": "teams-plugins",
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
		{
			"overrideTopologySpreadConstraintsSelectorLabels",
			map[string]string{
				"pluginsSettings.enabled": "true",
				"pluginsSettings.topologySpreadConstraints[0].matchLabelKeys[0]":             "pod-template-hash",
				"pluginsSettings.topologySpreadConstraints[0].maxSkew":                       "1",
				"pluginsSettings.topologySpreadConstraints[0].minDomains":                    "1",
				"pluginsSettings.topologySpreadConstraints[0].nodeAffinityPolicy":            "Honor",
				"pluginsSettings.topologySpreadConstraints[0].nodeTaintsPolicy":              "Honor",
				"pluginsSettings.topologySpreadConstraints[0].labelSelector.matchLabels.app": "foo",
				"pluginsSettings.topologySpreadConstraints[0].topologyKey":                   "kubernetes.io/hostname",
				"pluginsSettings.topologySpreadConstraints[0].whenUnsatisfiable":             "DoNotSchedule",
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
							"app": "foo"
						}
					  }
					}
				  ]`
				err := json.Unmarshal([]byte(expectedTopologySpreadConstraintJSON), &expectedTopologySpreadConstraint)
				s.NoError(err)
				s.Equal(expectedTopologySpreadConstraint, constraint, "Constraints should be equal")
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

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.TopologySpreadConstraints)
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerCount() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int
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
			1,
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
				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, len(deployment.Spec.Template.Spec.Containers), "Container count should be equal.")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, len(deployment.Spec.Template.Spec.Containers), "Container count should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerEnv() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(envVars []corev1.EnvVar)
	}{
		{
			"defaultValues",
			nil,
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[]`
				var expectedEnvVars []corev1.EnvVar
				err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
				s.NoError(err)
				s.Equal(expectedEnvVars, envVars, "Envs should be equal")
			},
		},
		{
			"defaultValuesPluginsEnabled", // legacy auth mode
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_AUTH_SECRET",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneAuthSecret"
              }
            }
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
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_APP_IMAGES",
            "value": "false"
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
		{
			"overrideEnv", // legacy auth mode
			map[string]string{
				"pluginsSettings.enabled":                                       "true",
				"pluginsSettings.env.TEST_KEY":                                  "TEST_VALUE",
				"pluginsSettings.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretName": "an-existing-secret", // pragma: allowlist secret
				"pluginsSettings.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretKey":  "anExistingKey",      // pragma: allowlist secret
			},
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_AUTH_SECRET",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneAuthSecret"
              }
            }
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
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_APP_IMAGES",
            "value": "false"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
            "value": "-1"
          },
          {
            "name": "TEST_KEY",
            "value": "TEST_VALUE"
          },
          {
            "name": "AN_ADDITIONAL_SECRET_ENV",
            "valueFrom": {
              "secretKeyRef": {
                "name": "an-existing-secret",
                "key": "anExistingKey"
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
		{
			"internalAuthMode",
			map[string]string{
				"casSettings.env.FIFTYONE_AUTH_MODE":                            "internal",
				"pluginsSettings.enabled":                                       "true",
				"pluginsSettings.env.TEST_KEY":                                  "TEST_VALUE",
				"pluginsSettings.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretName": "an-existing-secret", // pragma: allowlist secret
				"pluginsSettings.secretEnv.AN_ADDITIONAL_SECRET_ENV.secretKey":  "anExistingKey",      // pragma: allowlist secret
			},
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_AUTH_SECRET",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneAuthSecret"
              }
            }
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
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_APP_IMAGES",
            "value": "false"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_SIZE_BYTES",
            "value": "-1"
          },
          {
            "name": "TEST_KEY",
            "value": "TEST_VALUE"
          },
          {
            "name": "AN_ADDITIONAL_SECRET_ENV",
            "valueFrom": {
              "secretKeyRef": {
                "name": "an-existing-secret",
                "key": "anExistingKey"
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
		{
			"overrideSecretName",
			map[string]string{
				"pluginsSettings.enabled": "true",
				"secret.name":             "override-secret-name",
			},
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_AUTH_SECRET",
            "valueFrom": {
              "secretKeyRef": {
                "name": "override-secret-name",
                "key": "fiftyoneAuthSecret"
              }
            }
          },
          {
            "name": "FIFTYONE_DATABASE_ADMIN",
            "value": "false"
          },
          {
            "name": "FIFTYONE_DATABASE_NAME",
            "valueFrom": {
              "secretKeyRef": {
                "name": "override-secret-name",
                "key": "fiftyoneDatabaseName"
              }
            }
          },
          {
            "name": "FIFTYONE_DATABASE_URI",
            "valueFrom": {
              "secretKeyRef": {
                "name": "override-secret-name",
                "key": "mongodbConnectionString"
              }
            }
          },
          {
            "name": "FIFTYONE_ENCRYPTION_KEY",
            "valueFrom": {
              "secretKeyRef": {
                "name": "override-secret-name",
                "key": "encryptionKey"
              }
            }
          },
          {
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_APP_IMAGES",
            "value": "false"
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
		{
			"overrideApiServiceNameAndPort",
			map[string]string{
				"pluginsSettings.enabled":  "true",
				"apiSettings.service.name": "teams-api-override",
				"apiSettings.service.port": "8000",
			},
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api-override:8000"
          },
          {
            "name": "FIFTYONE_AUTH_SECRET",
            "valueFrom": {
              "secretKeyRef": {
                "name": "fiftyone-teams-secrets",
                "key": "fiftyoneAuthSecret"
              }
            }
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
            "name": "FIFTYONE_INTERNAL_SERVICE",
            "value": "true"
          },
          {
            "name": "FIFTYONE_MEDIA_CACHE_APP_IMAGES",
            "value": "false"
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].Env)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerImage() {

	// Get chart info (to later obtain the chart's appVersion)
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)

	// Get appVersion from chart info
	chartAppVersion, exists := cInfo["appVersion"]
	s.True(exists, "failed to get app version from chart info")

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
			fmt.Sprintf("voxel51/fiftyone-app:%s", chartAppVersion),
		},
		{
			"overrideImageTag",
			map[string]string{
				"pluginsSettings.enabled":   "true",
				"pluginsSettings.image.tag": "testTag",
			},
			"voxel51/fiftyone-app:testTag",
		},
		{
			"overrideImageRepository",
			map[string]string{
				"pluginsSettings.enabled":          "true",
				"pluginsSettings.image.repository": "ghcr.io/fiftyone-app",
			},
			fmt.Sprintf("ghcr.io/fiftyone-app:%s", chartAppVersion),
		},
		{
			"overrideImageVersionAndRepository",
			map[string]string{
				"pluginsSettings.enabled":          "true",
				"pluginsSettings.image.tag":        "testTag",
				"pluginsSettings.image.repository": "ghcr.io/fiftyone-app",
			},
			"ghcr.io/fiftyone-app:testTag",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, deployment.Spec.Template.Spec.Containers[0].Image, "Image values should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerImagePullPolicy() {
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
			"Always",
		},
		{
			"overrideImagePullPolicy",
			map[string]string{
				"pluginsSettings.enabled":          "true",
				"pluginsSettings.image.pullPolicy": "IfNotPresent",
			},
			"IfNotPresent",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, string(deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy), "Image pull policy should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerName() {
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
			"overrideServiceAccountName",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.name": "test-service-account",
			},
			"test-service-account",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, deployment.Spec.Template.Spec.Containers[0].Name, "Container name should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerLivenessProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			func(probe *corev1.Probe) {
				s.Empty(probe, "Liveness Probes should not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "tcpSocket": {
            "port": "teams-plugins"
          },
          "timeoutSeconds": 5
        }`
				var expectedProbe *corev1.Probe
				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
				s.NoError(err)
				s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
			},
		},
		{
			"overrideServiceLivenessAndShortName",
			map[string]string{
				"pluginsSettings.enabled":           "true",
				"pluginsSettings.service.shortname": "test-service-shortname",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "tcpSocket": {
            "port": "test-service-shortname"
          },
          "timeoutSeconds": 5
        }`
				var expectedProbe *corev1.Probe
				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
				s.NoError(err)
				s.Equal(expectedProbe, probe, "Liveness Probes should be equal")
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerPorts() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(port []corev1.ContainerPort)
	}{
		{
			"defaultValues",
			nil,
			func(ports []corev1.ContainerPort) {
				s.Empty(ports, "Ports should not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(ports []corev1.ContainerPort) {
				expectedPortsJSON := `[
          {
            "name": "teams-plugins",
            "containerPort": 5151,
            "protocol": "TCP"
          }
        ]`
				var expectedPorts []corev1.ContainerPort
				err := json.Unmarshal([]byte(expectedPortsJSON), &expectedPorts)
				s.NoError(err)
				s.Equal(expectedPorts, ports, "Ports should be equal")
			},
		},
		{
			"overrideServiceContainerPortAndShortName",
			map[string]string{
				"pluginsSettings.enabled":               "true",
				"pluginsSettings.service.containerPort": "5155",
				"pluginsSettings.service.shortname":     "test-service-shortname",
			},
			func(ports []corev1.ContainerPort) {
				expectedPortsJSON := `[
          {
            "name": "test-service-shortname",
            "containerPort": 5155,
            "protocol": "TCP"
          }
        ]`
				var expectedPorts []corev1.ContainerPort
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

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].Ports)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerReadinessProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			func(probe *corev1.Probe) {
				s.Empty(probe, "Readiness Probes should not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "tcpSocket": {
            "port": "teams-plugins"
          },
          "timeoutSeconds": 5
        }`
				var expectedProbe *corev1.Probe
				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
				s.NoError(err)
				s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
			},
		},
		{
			"overrideServiceReadinessAndShortName",
			map[string]string{
				"pluginsSettings.enabled":           "true",
				"pluginsSettings.service.shortname": "test-service-shortname",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "tcpSocket": {
            "port": "test-service-shortname"
          },
          "timeoutSeconds": 5
        }`
				var expectedProbe *corev1.Probe
				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
				s.NoError(err)
				s.Equal(expectedProbe, probe, "Readiness Probes should be equal")
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerStartupProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			func(probe *corev1.Probe) {
				s.Empty(probe, "Startup Probes should not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "tcpSocket": {
            "port": "teams-plugins"
          },
          "failureThreshold": 5,
          "periodSeconds": 15,
          "timeoutSeconds": 5
        }`
				var expectedProbe *corev1.Probe
				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
				s.NoError(err)
				s.Equal(expectedProbe, probe, "Startup Probes should be equal")
			},
		},
		{
			"overrideServiceStartupFailureThresholdAndPeriodSecondsAndShortName",
			map[string]string{
				"pluginsSettings.enabled":                          "true",
				"pluginsSettings.service.shortname":                "test-service-shortname",
				"pluginsSettings.service.startup.failureThreshold": "10",
				"pluginsSettings.service.startup.periodSeconds":    "10",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "tcpSocket": {
            "port": "test-service-shortname"
          },
          "failureThreshold": 10,
          "periodSeconds": 10,
          "timeoutSeconds": 5
        }`
				var expectedProbe *corev1.Probe
				err := json.Unmarshal([]byte(expectedProbeJSON), &expectedProbe)
				s.NoError(err)
				s.Equal(expectedProbe, probe, "Startup Probes should be equal")
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].StartupProbe)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerResourceRequirements() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(resourceRequirements corev1.ResourceRequirements)
	}{
		{
			"defaultValues",
			nil,
			func(resourceRequirements corev1.ResourceRequirements) {
				s.Empty(resourceRequirements, "Resource Requirements should be empty")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(resourceRequirements corev1.ResourceRequirements) {
				s.Equal(resourceRequirements.Limits, corev1.ResourceList{}, "Limits should be equal")
				s.Equal(resourceRequirements.Requests, corev1.ResourceList{}, "Requests should be equal")
				s.Nil(resourceRequirements.Claims, "should be nil")
			},
		},
		{
			"overrideResources",
			map[string]string{
				"pluginsSettings.enabled":                   "true",
				"pluginsSettings.resources.limits.cpu":      "1",
				"pluginsSettings.resources.limits.memory":   "1Gi",
				"pluginsSettings.resources.requests.cpu":    "500m",
				"pluginsSettings.resources.requests.memory": "512Mi",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].Resources)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerSecurityContext() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(securityContext *corev1.SecurityContext)
	}{
		{
			"defaultValues",
			nil,
			func(securityContext *corev1.SecurityContext) {
				s.Empty(securityContext, "should be not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
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
		{
			"overrideSecurityContext",
			map[string]string{
				"pluginsSettings.enabled":                    "true",
				"pluginsSettings.securityContext.runAsGroup": "3000",
				"pluginsSettings.securityContext.runAsUser":  "1000",
			},
			func(securityContext *corev1.SecurityContext) {
				s.Equal(int64(3000), *securityContext.RunAsGroup, "runAsGroup should be 3000")
				s.Equal(int64(1000), *securityContext.RunAsUser, "runAsUser should be 1000")
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].SecurityContext)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestContainerVolumeMounts() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(volumeMounts []corev1.VolumeMount)
	}{
		{
			"defaultValues",
			nil,
			func(volumeMounts []corev1.VolumeMount) {
				s.Empty(volumeMounts, "VolumeMounts should not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(volumeMounts []corev1.VolumeMount) {
				s.Nil(volumeMounts, "VolumeMounts should be nil")
			},
		},
		{
			"overrideVolumeMountsSingle",
			map[string]string{
				"pluginsSettings.enabled":                   "true",
				"pluginsSettings.volumeMounts[0].mountPath": "/test-data-volume",
				"pluginsSettings.volumeMounts[0].name":      "test-volume",
			},
			func(volumeMounts []corev1.VolumeMount) {
				expectedJSON := `[
          {
            "mountPath": "/test-data-volume",
            "name": "test-volume"
          }
        ]`
				var expectedVolumeMounts []corev1.VolumeMount
				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
				s.NoError(err)
				s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
			},
		},
		{
			"overrideVolumeMountsMultiple",
			map[string]string{
				"pluginsSettings.enabled":                   "true",
				"pluginsSettings.volumeMounts[0].mountPath": "/test-data-volume1",
				"pluginsSettings.volumeMounts[0].name":      "test-volume1",
				"pluginsSettings.volumeMounts[1].mountPath": "/test-data-volume2",
				"pluginsSettings.volumeMounts[1].name":      "test-volume2",
			},
			func(volumeMounts []corev1.VolumeMount) {
				expectedJSON := `[
          {
            "mountPath": "/test-data-volume1",
            "name": "test-volume1"
          },
          {
            "mountPath": "/test-data-volume2",
            "name": "test-volume2"
          }
        ]`
				var expectedVolumeMounts []corev1.VolumeMount
				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumeMounts)
				s.NoError(err)
				s.Equal(expectedVolumeMounts, volumeMounts, "Volume Mounts should be equal")
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
			}

		})
	}
}

func (s *deploymentPluginsTemplateTest) TestInitContainerCount() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int
	}{
		{
			"defaultValues",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			1,
		},
		{
			"overrideInitContainersEnabled",
			map[string]string{
				"pluginsSettings.enabled":                "true",
				"pluginsSettings.initContainers.enabled": "false",
			},
			0,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			s.Equal(testCase.expected, len(deployment.Spec.Template.Spec.InitContainers), "Init container count should be equal.")
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestInitContainerImage() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			"docker.io/busybox:stable-glibc",
		},
		{
			"overrideImageRepositoryAndTag",
			map[string]string{
				"pluginsSettings.enabled":                         "true",
				"pluginsSettings.initContainers.image.repository": "docker.io/bash",
				"pluginsSettings.initContainers.image.tag":        "devel-alpine3.20",
			},
			"docker.io/bash:devel-alpine3.20",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			s.Equal(testCase.expected, deployment.Spec.Template.Spec.InitContainers[0].Image, "Image values should be equal.")
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestInitContainerCommand() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(cmd []string)
	}{
		{
			"defaultValues",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(cmd []string) {
				expectedCmd := []string{
					"sh",
					"-c",
					"until wget -qO /dev/null teams-cas.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local/cas/api; do echo waiting for cas; sleep 2; done",
				}
				s.Equal(expectedCmd, cmd, "InitContainer commands should be equal")
			},
		},
		{
			"overrideCasHostname",
			map[string]string{
				"casSettings.service.name": "test-service-name",
				"pluginsSettings.enabled":  "true",
			},
			func(cmd []string) {
				expectedCmd := []string{
					"sh",
					"-c",
					"until wget -qO /dev/null test-service-name.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local/cas/api; do echo waiting for cas; sleep 2; done",
				}
				s.Equal(expectedCmd, cmd, "InitContainer commands should be equal")
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

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.InitContainers[0].Command)
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestAffinity() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(affinity *corev1.Affinity)
	}{
		{
			"defaultValues",
			nil,
			func(affinity *corev1.Affinity) {
				s.Nil(affinity, "should be nil")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(affinity *corev1.Affinity) {
				s.Nil(affinity, "should be nil")
			},
		},
		{
			"overrideAffinity",
			map[string]string{
				"pluginsSettings.enabled": "true",
				"pluginsSettings.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "disktype",
				"pluginsSettings.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
				"pluginsSettings.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "ssd",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Affinity)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestImagePullSecrets() {
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
			"overrideImagePullSecrets",
			map[string]string{
				"pluginsSettings.enabled":  "true",
				"imagePullSecrets[0].name": "test-pull-secret",
			},
			"test-pull-secret",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				if testCase.expected == "" {
					s.Nil(deployment.Spec.Template.Spec.ImagePullSecrets, "Image pull secret should be nil")
				} else {
					s.Equal(testCase.expected, deployment.Spec.Template.Spec.ImagePullSecrets[0].Name, "Image pull secret should be equal.")
				}
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestNodeSelector() {
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
			"overrideNodeSelector",
			map[string]string{
				"pluginsSettings.enabled":               "true",
				"pluginsSettings.nodeSelector.disktype": "ssd",
			},
			map[string]string{
				"disktype": "ssd",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				for key, value := range testCase.expected {
					foundValue := deployment.Spec.Template.Spec.NodeSelector[key]
					s.Equal(value, foundValue, "NodeSelector should contain all set labels.")
				}
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestPodAnnotations() {
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
			"overridePodAnnotations",
			map[string]string{
				"pluginsSettings.enabled":                     "true",
				"pluginsSettings.podAnnotations.annotation-1": "annotation-1-value",
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

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				if testCase.expected == nil {
					s.Nil(deployment.Spec.Template.ObjectMeta.Annotations, "Annotations should be nil")
				} else {
					for key, value := range testCase.expected {
						foundValue := deployment.Spec.Template.ObjectMeta.Annotations[key]
						s.Equal(value, foundValue, "Annotations should contain all set annotations.")
					}
				}
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestPodSecurityContext() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(podSecurityContext *corev1.PodSecurityContext)
	}{
		{
			"defaultValues",
			nil,
			func(podSecurityContext *corev1.PodSecurityContext) {
				s.Empty(podSecurityContext.FSGroup, "should not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
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
		{
			"overridePodSecurityContext",
			map[string]string{
				"pluginsSettings.enabled":                       "true",
				"pluginsSettings.podSecurityContext.fsGroup":    "2000",
				"pluginsSettings.podSecurityContext.runAsGroup": "3000",
				"pluginsSettings.podSecurityContext.runAsUser":  "1000",
			},
			func(podSecurityContext *corev1.PodSecurityContext) {
				s.Equal(int64(2000), *podSecurityContext.FSGroup, "fsGroup should be 2000")
				s.Equal(int64(3000), *podSecurityContext.RunAsGroup, "runAsGroup should be 3000")
				s.Equal(int64(1000), *podSecurityContext.RunAsUser, "runAsUser should be 1000")
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.SecurityContext)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestTemplateLabels() {
	testCases := []struct {
		name                     string
		values                   map[string]string
		selectorMatchExpected    map[string]string
		templateMetadataExpected map[string]string
	}{
		{
			"defaultValues",
			nil,
			nil,
			nil,
		},
		{
			"addTemplateMetadataLabels",
			map[string]string{
				"pluginsSettings.enabled":        "true",
				"pluginsSettings.labels.myLabel": "unruly",
			},
			map[string]string{
				"app.kubernetes.io/name":     "teams-plugins",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
			map[string]string{
				"app.kubernetes.io/name":     "teams-plugins",
				"app.kubernetes.io/instance": "fiftyone-test",
				"myLabel":                    "unruly",
			},
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
			map[string]string{
				"app.kubernetes.io/name":     "teams-plugins",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
		},
		{
			"overrideSelectorMatchLabels",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.name": "test-service-name",
			},
			map[string]string{
				"app.kubernetes.io/name":     "test-service-name",
				"app.kubernetes.io/instance": "fiftyone-test",
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

			// when vars are set outside of the if statement, they aren't accessible from within the conditional
			if testCase.values == nil {
				options := &helm.Options{SetValues: testCase.values}
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				for key, value := range testCase.selectorMatchExpected {

					foundValue := deployment.Spec.Selector.MatchLabels[key]
					s.Equal(value, foundValue, "Selector Labels should contain all set labels.")
				}

				for key, value := range testCase.templateMetadataExpected {

					foundValue := deployment.Spec.Template.ObjectMeta.Labels[key]
					s.Equal(value, foundValue, "Template Metadata Labels should contain all set labels.")
				}
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestServiceAccountName() {
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
			"overrideServiceAccountName",
			map[string]string{
				"pluginsSettings.enabled": "true",
				"serviceAccount.name":     "test-service-account",
			},
			"test-service-account",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Equal(testCase.expected, deployment.Spec.Template.Spec.ServiceAccountName, "Service account name should be equal.")
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestTolerations() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(tolerations []corev1.Toleration)
	}{
		{
			"defaultValues",
			nil,
			func(tolerations []corev1.Toleration) {
				s.Empty(tolerations, "should not be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(tolerations []corev1.Toleration) {
				s.Nil(tolerations, "should be nil")
			},
		},
		{
			"overrideTolerations",
			map[string]string{
				"pluginsSettings.enabled":                 "true",
				"pluginsSettings.tolerations[0].key":      "example-key",
				"pluginsSettings.tolerations[0].operator": "Exists",
				"pluginsSettings.tolerations[0].effect":   "NoSchedule",
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Tolerations)
			}
		})
	}
}

func (s *deploymentPluginsTemplateTest) TestVolumes() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(volumes []corev1.Volume)
	}{
		{
			"defaultValues",
			nil,
			func(volumes []corev1.Volume) {
				s.Empty(volumes, "Volumes should be be set")
			},
		},
		{
			"defaultValuesPluginsEnabled",
			map[string]string{
				"pluginsSettings.enabled": "true",
			},
			func(volumes []corev1.Volume) {
				s.Nil(volumes, "Volumes should be nil")
			},
		},
		{
			"overrideVolumesSingle",
			map[string]string{
				"pluginsSettings.enabled":                  "true",
				"pluginsSettings.volumes[0].name":          "test-volume",
				"pluginsSettings.volumes[0].hostPath.path": "/test-volume",
			},
			func(volumes []corev1.Volume) {
				expectedJSON := `[
          {
            "name": "test-volume",
            "hostPath": {
              "path": "/test-volume"
            }
          }
        ]`
				var expectedVolumes []corev1.Volume
				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
				s.NoError(err)
				s.Equal(expectedVolumes, volumes, "Volumes should be equal")
			},
		},
		{
			"overrideVolumesMultiple",
			map[string]string{
				"pluginsSettings.enabled":                                    "true",
				"pluginsSettings.volumes[0].name":                            "test-volume1",
				"pluginsSettings.volumes[0].hostPath.path":                   "/test-volume1",
				"pluginsSettings.volumes[1].name":                            "pvc1",
				"pluginsSettings.volumes[1].persistentVolumeClaim.claimName": "pvc1",
			},
			func(volumes []corev1.Volume) {
				expectedJSON := `[
          {
            "name": "test-volume1",
            "hostPath": {
              "path": "/test-volume1"
            }
          },
          {
            "name": "pvc1",
            "persistentVolumeClaim": {
              "claimName": "pvc1"
            }
          }
        ]`
				var expectedVolumes []corev1.Volume
				err := json.Unmarshal([]byte(expectedJSON), &expectedVolumes)
				s.NoError(err)
				s.Equal(expectedVolumes, volumes, "Volumes should be equal")
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

				s.ErrorContains(err, "could not find template templates/plugins-deployment.yaml in chart")
				var deployment appsv1.Deployment

				helm.UnmarshalK8SYaml(subT, output, &deployment)

				s.Nil(deployment.Spec.Template.Spec.Containers)
			} else {
				options := &helm.Options{SetValues: testCase.values}
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(subT, output, &deployment)

				testCase.expected(deployment.Spec.Template.Spec.Volumes)
			}

		})
	}
}
