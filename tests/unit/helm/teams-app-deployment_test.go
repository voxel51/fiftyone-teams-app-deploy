//go:build kubeall || helm || unit || unitTeamsAppDeployment
// +build kubeall helm unit unitTeamsAppDeployment

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

type deploymentTeamsAppTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestDeploymentTeamsAppTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &deploymentTeamsAppTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/teams-app-deployment.yaml"},
	})
}

func (s *deploymentTeamsAppTemplateTest) TestMetadataLabels() {
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
				"app.kubernetes.io/name":       "fiftyone-teams-app",
				"app.kubernetes.io/instance":   "fiftyone-test",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			for key, value := range testCase.expected {
				foundValue := deployment.ObjectMeta.Labels[key]
				s.Equal(value, foundValue, "Labels should contain all set labels.")
			}
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"teams-app",
		},
		{
			"overrideMetadataName",
			map[string]string{
				"teamsAppSettings.service.name": "test-service-name",
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

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			s.Equal(testCase.expected, deployment.ObjectMeta.Name, "Deployment name should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestMetadataNamespace() {
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

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			s.Equal(testCase.expected, deployment.ObjectMeta.Namespace, "Namespace name should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestReplicas() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int32
	}{
		{
			"defaultValues",
			nil,
			2,
		},
		{
			"overrideReplicaCount",
			map[string]string{
				"teamsAppSettings.replicaCount": "3",
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
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			s.Equal(testCase.expected, *deployment.Spec.Replicas, "Replica count should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerCount() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected int
	}{
		{
			"defaultValues",
			nil,
			1,
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

			s.Equal(testCase.expected, len(deployment.Spec.Template.Spec.Containers), "Container count should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerEnv() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(envVars []corev1.EnvVar)
	}{
		{
			"defaultValues", // legacy auth mode
			nil,
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_API_URI",
            "value": "https://"
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
            "name": "FIFTYONE_SERVER_ADDRESS",
            "value": ""
          },
          {
            "name": "FIFTYONE_SERVER_PATH_PREFIX",
            "value": "/api/proxy/fiftyone-teams"
          },
          {
            "name": "FIFTYONE_TEAMS_PROXY_URL",
            "value": "http://fiftyone-app:80"
          },
          {
            "name": "FIFTYONE_TEAMS_PLUGIN_URL",
            "value": "http://fiftyone-app:80"
          },
          {
            "name": "APP_USE_HTTPS",
            "value": "true"
          },
          {
            "name": "FIFTYONE_APP_ALLOW_MEDIA_EXPORT",
            "value": "true"
          },
          {
            "name": "FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED",
            "value": "true"
          },
          {
            "name": "FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION",
            "value": "2.2.0"
          },
          {
            "name": "FIFTYONE_APP_THEME",
            "value": "dark"
          },
          {
            "name": "RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED",
            "value": "false"
          }
        ]`
				var expectedEnvVars []corev1.EnvVar
				err := json.Unmarshal([]byte(expectedEnvVarJSON), &expectedEnvVars)
				s.NoError(err)
				s.Equal(expectedEnvVars, envVars, "Envs should be equal")
			},
		},
		{
			"overrideEnv",
			map[string]string{
				"teamsAppSettings.env.TEST_KEY": "TEST_VALUE",
			},
			func(envVars []corev1.EnvVar) {
				expectedEnvVarJSON := `[
          {
            "name": "API_URL",
            "value": "http://teams-api:80"
          },
          {
            "name": "FIFTYONE_API_URI",
            "value": "https://"
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
            "name": "FIFTYONE_SERVER_ADDRESS",
            "value": ""
          },
          {
            "name": "FIFTYONE_SERVER_PATH_PREFIX",
            "value": "/api/proxy/fiftyone-teams"
          },
          {
            "name": "FIFTYONE_TEAMS_PROXY_URL",
            "value": "http://fiftyone-app:80"
          },
          {
            "name": "FIFTYONE_TEAMS_PLUGIN_URL",
            "value": "http://fiftyone-app:80"
          },
          {
            "name": "APP_USE_HTTPS",
            "value": "true"
          },
          {
            "name": "FIFTYONE_APP_ALLOW_MEDIA_EXPORT",
            "value": "true"
          },
          {
            "name": "FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED",
            "value": "true"
          },
          {
            "name": "FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION",
            "value": "2.2.0"
          },
          {
            "name": "FIFTYONE_APP_THEME",
            "value": "dark"
          },
          {
            "name": "RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED",
            "value": "false"
          },
          {
            "name": "TEST_KEY",
            "value": "TEST_VALUE"
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].Env)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerImage() {

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
			fmt.Sprintf("voxel51/fiftyone-teams-app:%s", chartAppVersion),
		},
		{
			"overrideImageTag",
			map[string]string{
				"teamsAppSettings.image.tag": "testTag",
			},
			"voxel51/fiftyone-teams-app:testTag",
		},
		{
			"overrideImageRepository",
			map[string]string{
				"teamsAppSettings.image.repository": "ghcr.io/fiftyone-teams-app",
			},
			fmt.Sprintf("ghcr.io/fiftyone-teams-app:%s", chartAppVersion),
		},
		{
			"overrideImageVersionAndRepository",
			map[string]string{
				"teamsAppSettings.image.tag":        "testTag",
				"teamsAppSettings.image.repository": "ghcr.io/fiftyone-teams-app",
			},
			"ghcr.io/fiftyone-teams-app:testTag",
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

			s.Equal(testCase.expected, deployment.Spec.Template.Spec.Containers[0].Image, "Image values should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerImagePullPolicy() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"Always",
		},
		{
			"overrideImagePullPolicy",
			map[string]string{
				"teamsAppSettings.image.pullPolicy": "IfNotPresent",
			},
			"IfNotPresent",
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

			// Convert type returned by `.ImagePullPolicy` to a string
			s.Equal(testCase.expected, string(deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy), "Image pull policy should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"teams-app",
		},
		{
			"overrideServiceAccountName",
			map[string]string{
				"teamsAppSettings.service.name": "test-service-account",
			},
			"test-service-account",
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

			s.Equal(testCase.expected, deployment.Spec.Template.Spec.Containers[0].Name, "Container name should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerLivenessProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "httpGet": {
            "path": "/api/hello",
            "port": "teams-app"
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
			"overrideServiceLivenessShortName",
			map[string]string{
				"teamsAppSettings.service.shortname": "test-service-shortname",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "httpGet": {
            "path": "/api/hello",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerPorts() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(port []corev1.ContainerPort)
	}{
		{
			"defaultValues",
			nil,
			func(ports []corev1.ContainerPort) {
				expectedPortsJSON := `[
          {
            "name": "teams-app",
            "containerPort": 3000,
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
				"teamsAppSettings.service.containerPort": "3001",
				"teamsAppSettings.service.shortname":     "test-service-shortname",
			},
			func(ports []corev1.ContainerPort) {
				expectedPortsJSON := `[
          {
            "name": "test-service-shortname",
            "containerPort": 3001,
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].Ports)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerReadinessProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "httpGet": {
            "path": "/api/hello",
            "port": "teams-app"
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
			"overrideServiceReadinessShortName",
			map[string]string{
				"teamsAppSettings.service.shortname": "test-service-shortname",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "httpGet": {
            "path": "/api/hello",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerStartupProbe() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(probe *corev1.Probe)
	}{
		{
			"defaultValues",
			nil,
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "httpGet": {
            "path": "/api/hello",
            "port": "teams-app"
          },
          "failureThreshold": 5,
          "periodSeconds": 5,
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
				"teamsAppSettings.service.shortname":                "test-service-shortname",
				"teamsAppSettings.service.startup.failureThreshold": "10",
				"teamsAppSettings.service.startup.periodSeconds":    "10",
			},
			func(probe *corev1.Probe) {
				expectedProbeJSON := `{
          "httpGet": {
            "path": "/api/hello",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].StartupProbe)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerResourceRequirements() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(resourceRequirements corev1.ResourceRequirements)
	}{
		{
			"defaultValues",
			nil,
			func(resourceRequirements corev1.ResourceRequirements) {
				s.Equal(resourceRequirements.Limits, corev1.ResourceList{}, "Limits should be equal")
				s.Equal(resourceRequirements.Requests, corev1.ResourceList{}, "Requests should be equal")
				s.Nil(resourceRequirements.Claims, "should be nil")
			},
		},
		{
			"overrideResources",
			map[string]string{
				"teamsAppSettings.resources.limits.cpu":      "1",
				"teamsAppSettings.resources.limits.memory":   "1Gi",
				"teamsAppSettings.resources.requests.cpu":    "500m",
				"teamsAppSettings.resources.requests.memory": "512Mi",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].Resources)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerSecurityContext() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(securityContext *corev1.SecurityContext)
	}{
		{
			"defaultValues",
			nil,
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
				"teamsAppSettings.securityContext.runAsGroup": "3000",
				"teamsAppSettings.securityContext.runAsUser":  "1000",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].SecurityContext)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestContainerVolumeMounts() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(volumeMounts []corev1.VolumeMount)
	}{
		{
			"defaultValues",
			nil,
			func(volumeMounts []corev1.VolumeMount) {
				s.Nil(volumeMounts, "VolumeMounts should be nil")
			},
		},
		{
			"overrideVolumeMountsSingle",
			map[string]string{
				"teamsAppSettings.volumeMounts[0].mountPath": "/test-data-volume",
				"teamsAppSettings.volumeMounts[0].name":      "test-volume",
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
				"teamsAppSettings.volumeMounts[0].mountPath": "/test-data-volume1",
				"teamsAppSettings.volumeMounts[0].name":      "test-volume1",
				"teamsAppSettings.volumeMounts[1].mountPath": "/test-data-volume2",
				"teamsAppSettings.volumeMounts[1].name":      "test-volume2",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestAffinity() {
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
			"overrideAffinity",
			map[string]string{
				"teamsAppSettings.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "disktype",
				"teamsAppSettings.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
				"teamsAppSettings.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "ssd",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Affinity)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestImagePullSecrets() {
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
			"overrideImagePullSecrets",
			map[string]string{
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			if testCase.expected == "" {
				s.Nil(deployment.Spec.Template.Spec.ImagePullSecrets, "Image pull secret should be nil")
			} else {
				s.Equal(testCase.expected, deployment.Spec.Template.Spec.ImagePullSecrets[0].Name, "Image pull secret should be equal.")
			}
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestNodeSelector() {
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
			"overrideNodeSelector",
			map[string]string{
				"teamsAppSettings.nodeSelector.disktype": "ssd",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			for key, value := range testCase.expected {
				foundValue := deployment.Spec.Template.Spec.NodeSelector[key]
				s.Equal(value, foundValue, "NodeSelector should contain all set labels.")
			}
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestPodAnnotations() {
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
			"overridePodAnnotations",
			map[string]string{
				"teamsAppSettings.podAnnotations.annotation-1": "annotation-1-value",
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
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestPodSecurityContext() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(podSecurityContext *corev1.PodSecurityContext)
	}{
		{
			"defaultValues",
			nil,
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
				"teamsAppSettings.podSecurityContext.fsGroup":    "2000",
				"teamsAppSettings.podSecurityContext.runAsGroup": "3000",
				"teamsAppSettings.podSecurityContext.runAsUser":  "1000",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.SecurityContext)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestTemplateLabels() {
	testCases := []struct {
		name                     string
		values                   map[string]string
		selectorMatchExpected    map[string]string
		templateMetadataExpected map[string]string
	}{
		{
			"addTemplateMetadataLabels",
			map[string]string{
				"teamsAppSettings.labels.someLabel": "orange",
			},
			map[string]string{
				"app.kubernetes.io/name":     "fiftyone-teams-app",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
			map[string]string{
				"app.kubernetes.io/name":     "fiftyone-teams-app",
				"app.kubernetes.io/instance": "fiftyone-test",
				"someLabel":                  "orange",
			},
		},
		{
			"defaultValues",
			nil,
			map[string]string{
				"app.kubernetes.io/name":     "fiftyone-teams-app",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
			map[string]string{
				"app.kubernetes.io/name":     "fiftyone-teams-app",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
		},
		{
			"overrideSelectorMatchLabels",
			map[string]string{
				// Unlike teams-api, fiftyone-app, and teams-plugins, setting `teamsAppSettings.service.name`
				// does not affect the label `app.kubernetes.io/name` for teams-app.
				// See note in _helpers.tpl.
				"teamsAppSettings.service.name": "test-service-name",
			},
			map[string]string{
				"app.kubernetes.io/name":     "fiftyone-teams-app",
				"app.kubernetes.io/instance": "fiftyone-test",
			},
			map[string]string{
				"app.kubernetes.io/name":     "fiftyone-teams-app",
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
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestServiceAccountName() {
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
			"overrideServiceAccountName",
			map[string]string{
				"serviceAccount.name": "test-service-account",
			},
			"test-service-account",
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

			s.Equal(testCase.expected, deployment.Spec.Template.Spec.ServiceAccountName, "Service account name should be equal.")
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestTolerations() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(tolerations []corev1.Toleration)
	}{
		{
			"defaultValues",
			nil,
			func(tolerations []corev1.Toleration) {
				s.Nil(tolerations, "should be nil")
			},
		},
		{
			"overrideTolerations",
			map[string]string{
				"teamsAppSettings.tolerations[0].key":      "example-key",
				"teamsAppSettings.tolerations[0].operator": "Exists",
				"teamsAppSettings.tolerations[0].effect":   "NoSchedule",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Tolerations)
		})
	}
}

func (s *deploymentTeamsAppTemplateTest) TestVolumes() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(volumes []corev1.Volume)
	}{
		{
			"defaultValues",
			nil,
			func(volumes []corev1.Volume) {
				s.Nil(volumes, "Volumes should be nil")
			},
		},
		{
			"overrideVolumesSingle",
			map[string]string{
				"teamsAppSettings.volumes[0].name":          "test-volume",
				"teamsAppSettings.volumes[0].hostPath.path": "/test-volume",
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
				"teamsAppSettings.volumes[0].name":                            "test-volume1",
				"teamsAppSettings.volumes[0].hostPath.path":                   "/test-volume1",
				"teamsAppSettings.volumes[1].name":                            "pvc1",
				"teamsAppSettings.volumes[1].persistentVolumeClaim.claimName": "pvc1",
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

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(subT, output, &deployment)

			testCase.expected(deployment.Spec.Template.Spec.Volumes)
		})
	}
}
