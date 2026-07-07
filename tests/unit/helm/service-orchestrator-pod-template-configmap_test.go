//go:build kubeall || helm || unit || unitServiceOrchestrator
// +build kubeall helm unit unitServiceOrchestrator

package unit

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/noirbizarre/gonja"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corev1 "k8s.io/api/core/v1"
)

type servicePodTemplateConfigMapTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

// The per-task variables the kubernetes-service broker renders into the
// jinja2 pod template. Mirrors the escaped placeholders in the template.
func brokerTemplateVars() map[string]interface{} {
	return map[string]interface{}{
		"_id":                "task1",
		"_name":              "svc-task1",
		"_namespace":         "svc-ns",
		"_image":             "registry:5000/svc:tag",
		"_command":           "python",
		"_args":              []string{"-m", "some.module"},
		"_env":               []map[string]interface{}{{"name": "A", "value": "b"}},
		"_port":              8000,
		"_health_path":       "/healthz",
		"_health_port":       8000,
		"_cpu":               "2",
		"_memory":            "8Gi",
		"_ephemeral_storage": "9Gi",
		"_gpu_count":         1,
		"_gpu_type":          "nvidia-h100-80gb",
	}
}

func renderPodTemplate(s *suite.Suite, payload string, vars map[string]interface{}) corev1.Pod {
	template, err := gonja.FromString(payload)
	s.Require().NoError(err, "pod.yaml.j2 should parse as jinja")

	rendered, err := template.Execute(vars)
	s.Require().NoError(err, "pod.yaml.j2 should render with broker vars")

	var pod corev1.Pod
	helm.UnmarshalK8SYaml(s.T(), rendered, &pod)
	return pod
}

func TestServicePodTemplateConfigMapTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &servicePodTemplateConfigMapTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates: []string{
			"templates/service-orchestrator-pod-template-configmap.yaml",
		},
	})
}

func (s *servicePodTemplateConfigMapTemplateTest) TestGatingAndName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string // empty => the template must not render
	}{
		{
			"defaultValues",
			nil,
			"",
		},
		{
			"enabledButCreateDisabled",
			map[string]string{
				"serviceOrchestrator.enabled":                      "true",
				"serviceOrchestrator.podTemplate.configMap.create": "false",
			},
			"",
		},
		{
			"enabled",
			map[string]string{
				"serviceOrchestrator.enabled": "true",
			},
			fmt.Sprintf("%s-fiftyone-teams-app-service-pod-template", s.releaseName),
		},
		{
			"enabledWithNameOverride",
			map[string]string{
				"serviceOrchestrator.enabled":                    "true",
				"serviceOrchestrator.podTemplate.configMap.name": "custom-pod-template",
			},
			"custom-pod-template",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

			if testCase.expected == "" {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/service-orchestrator-pod-template-configmap.yaml in chart")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var configMap corev1.ConfigMap
				helm.UnmarshalK8SYaml(subT, output, &configMap)

				s.Equal(testCase.expected, configMap.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

func (s *servicePodTemplateConfigMapTemplateTest) TestPayloadRendersAServicePod() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"serviceOrchestrator.enabled": "true",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	payload, ok := configMap.Data["pod.yaml.j2"]
	s.True(ok, "ConfigMap should carry pod.yaml.j2")

	pod := renderPodTemplate(&s.Suite, payload, brokerTemplateVars())

	s.Equal("svc-task1", pod.ObjectMeta.Name)
	s.Equal("svc-ns", pod.ObjectMeta.Namespace)

	s.Require().Len(pod.Spec.Containers, 1)
	container := pod.Spec.Containers[0]
	s.Equal("service", container.Name)
	s.Equal("registry:5000/svc:tag", container.Image)
	s.Equal([]string{"python"}, container.Command)
	s.Equal([]string{"-m", "some.module"}, container.Args)

	// GPU request + node selector from the broker vars
	gpu := container.Resources.Limits[corev1.ResourceName("nvidia.com/gpu")]
	s.Equal("1", gpu.String())
	s.Equal(
		"nvidia-h100-80gb",
		pod.Spec.NodeSelector["cloud.google.com/gke-accelerator"],
	)

	// Health-checked service gets a readiness probe on the health port
	s.Require().NotNil(container.ReadinessProbe)
	s.Equal("/healthz", container.ReadinessProbe.HTTPGet.Path)

	// Telemetry disabled: no sidecar, no socket plumbing
	s.Empty(pod.Spec.InitContainers, "No sidecar with telemetry disabled")
	for _, env := range container.Env {
		s.NotEqual("TELEMETRY_SOCKET", env.Name)
	}
}

func (s *servicePodTemplateConfigMapTemplateTest) TestPayloadWithoutHealthcheck() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"serviceOrchestrator.enabled": "true",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	vars := brokerTemplateVars()
	vars["_health_path"] = ""
	pod := renderPodTemplate(&s.Suite, configMap.Data["pod.yaml.j2"], vars)

	s.Nil(pod.Spec.Containers[0].ReadinessProbe, "No probe without a health path")
}

func (s *servicePodTemplateConfigMapTemplateTest) TestPayloadWithTelemetry() {
	// Telemetry left at its default (enabled): the pod gets the native
	// sidecar and the shared telemetry socket.
	options := &helm.Options{
		SetValues: map[string]string{
			"serviceOrchestrator.enabled": "true",
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	pod := renderPodTemplate(&s.Suite, configMap.Data["pod.yaml.j2"], brokerTemplateVars())

	s.Require().Len(pod.Spec.InitContainers, 1)
	s.Equal("telemetry-sidecar", pod.Spec.InitContainers[0].Name)

	volumeNames := []string{}
	for _, volume := range pod.Spec.Volumes {
		volumeNames = append(volumeNames, volume.Name)
	}
	s.Contains(volumeNames, "telemetry-socket")
}
