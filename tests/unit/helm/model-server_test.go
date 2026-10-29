//go:build kubeall || helm || unit || unitModelServer
// +build kubeall helm unit unitModelServer

package unit

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const modelServerChartPath = "../../../helm/fiftyone-model-server/"

type modelServerTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestModelServerTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(modelServerChartPath)
	require.NoError(t, err)

	suite.Run(t, &modelServerTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "model-server",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

// The values every render needs: the image is not published, so it has no
// default, and a box with nothing to preload is never Ready.
func minimal(extra map[string]string) map[string]string {
	values := map[string]string{
		"image.repository": "registry.example/model-server",
		"image.tag":        "v1",
		"preload[0]":       "siglip-base-patch16-224-torch",
	}
	for key, value := range extra {
		values[key] = value
	}
	return values
}

func (s *modelServerTemplateTest) render(values map[string]string, template string) string {
	options := &helm.Options{SetValues: values}
	return helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{template})
}

func (s *modelServerTemplateTest) TestRefusesWhatCannotWork() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"noImageRepository",
			map[string]string{"image.tag": "v1", "preload[0]": "siglip"},
			"image.repository is required",
		},
		{
			"noImageTag",
			map[string]string{"image.repository": "r", "preload[0]": "siglip"},
			"image.tag is required",
		},
		{
			"nothingToPreload",
			map[string]string{"image.repository": "r", "image.tag": "v1"},
			"preload is required",
		},
		{
			"loadBalancerOpenToAll",
			minimal(map[string]string{"service.type": "LoadBalancer"}),
			"service.loadBalancerSourceRanges is required",
		},
		{
			"networkPolicyWithNoCaller",
			minimal(map[string]string{"networkPolicy.from": "null"}),
			"networkPolicy.from is required",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, []string{})

			s.ErrorContains(err, testCase.expected)
		})
	}

	// An empty list, unlike a missing one, would render a policy that
	// admits everyone
	empty := &helm.Options{SetValues: minimal(nil), SetJsonValues: map[string]string{"networkPolicy.from": "[]"}}
	_, err := helm.RenderTemplateE(s.T(), empty, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "networkPolicy.from is required")
}

func (s *modelServerTemplateTest) TestSettings() {
	var config corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"preload[1]":              "qwen3-vl-embedding-2b-torch",
		"allowedModels[0]":        "siglip-base-patch16-224-torch",
		"allowedMediaPrefixes[0]": "gs://media",
		"maxLoadedModels":         "3",
		"env.EXTRA":               "on",
	}), "templates/configmap.yaml"), &config)

	s.Equal(`["siglip-base-patch16-224-torch","qwen3-vl-embedding-2b-torch"]`, config.Data["MODEL_SERVER_PRELOAD"])
	s.Equal(`["siglip-base-patch16-224-torch"]`, config.Data["MODEL_SERVER_ALLOWED_MODELS"])
	s.Equal(`[]`, config.Data["MODEL_SERVER_ALLOWED_MODEL_KWARGS"])
	s.Equal(`["gs://media"]`, config.Data["MODEL_SERVER_ALLOWED_MEDIA_PREFIXES"])
	s.Equal("3", config.Data["MODEL_SERVER_MAX_LOADED_MODELS"])
	s.Equal("8080", config.Data["MODEL_SERVER_PORT"])
	s.Equal("/models/zoo", config.Data["FIFTYONE_MODEL_ZOO_DIR"])
	s.Equal("/models/hf", config.Data["HF_HOME"])
	s.Equal("/models/torch", config.Data["TORCH_HOME"])
	s.Equal("on", config.Data["EXTRA"])
}

func (s *modelServerTemplateTest) TestOneBoxOnCpu() {
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(nil), "templates/deployment.yaml"), &deployment)

	s.Equal(int32(1), *deployment.Spec.Replicas)
	s.Equal(appsv1.RecreateDeploymentStrategyType, deployment.Spec.Strategy.Type)
	container := deployment.Spec.Template.Spec.Containers[0]
	s.Equal("registry.example/model-server:v1", container.Image)
	_, hasGpu := container.Resources.Limits["nvidia.com/gpu"]
	s.False(hasGpu)
	s.Empty(container.Resources.Limits)
	s.Equal("/healthz", container.StartupProbe.HTTPGet.Path)
	s.Equal(int32(90), container.StartupProbe.FailureThreshold)
	s.Equal("/readyz", container.ReadinessProbe.HTTPGet.Path)
	s.Equal("/healthz", container.LivenessProbe.HTTPGet.Path)
	s.False(*deployment.Spec.Template.Spec.AutomountServiceAccountToken)
	s.NotNil(deployment.Spec.Template.Spec.Volumes[0].EmptyDir)
	// The weights paths in the settings are under this mount
	s.Equal("/models", container.VolumeMounts[0].MountPath)
	s.Equal(int32(8080), container.Ports[0].ContainerPort)

	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{"weights.sizeLimit": "20Gi"}), "templates/deployment.yaml"), &deployment)
	limit := resource.MustParse("20Gi")
	s.Equal(&limit, deployment.Spec.Template.Spec.Volumes[0].EmptyDir.SizeLimit)
}

func (s *modelServerTemplateTest) TestGpuKeepsTheOtherResources() {
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"gpu.enabled":               "true",
		"gpu.count":                 "2",
		"nodeSelector.gpu":          "yes",
		"tolerations[0].key":        "nvidia.com/gpu",
		"tolerations[0].operator":   "Exists",
		"resources.limits.memory":   "32Gi",
		"resources.requests.cpu":    "8",
		"resources.requests.memory": "32Gi",
	}), "templates/deployment.yaml"), &deployment)

	container := deployment.Spec.Template.Spec.Containers[0]
	s.Equal(resource.MustParse("2"), container.Resources.Limits["nvidia.com/gpu"])
	s.Equal(resource.MustParse("2"), container.Resources.Requests["nvidia.com/gpu"])
	s.Equal(resource.MustParse("32Gi"), container.Resources.Limits[corev1.ResourceMemory])
	s.Equal(resource.MustParse("8"), container.Resources.Requests[corev1.ResourceCPU])
	s.Equal("yes", deployment.Spec.Template.Spec.NodeSelector["gpu"])
	s.Equal("nvidia.com/gpu", deployment.Spec.Template.Spec.Tolerations[0].Key)
}

func (s *modelServerTemplateTest) TestPodSettings() {
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"serviceAccountName":                     "model-server",
		"imagePullSecrets[0].name":               "registry",
		"podAnnotations.team":                    "inference",
		"podSecurityContext.fsGroup":             "1000",
		"securityContext.readOnlyRootFilesystem": "false",
		"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":      "gpu",
		"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator": "Exists",
		"gpu.enabled":                       "true",
		"resources.limits.nvidia\\.com/gpu": "4",
	}), "templates/deployment.yaml"), &deployment)

	spec := deployment.Spec.Template.Spec
	s.Equal("model-server", spec.ServiceAccountName)
	s.Equal("registry", spec.ImagePullSecrets[0].Name)
	s.Equal("inference", deployment.Spec.Template.Annotations["team"])
	s.Equal(int64(1000), *spec.SecurityContext.FSGroup)
	s.False(*spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem)
	s.Equal("gpu", spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key)
	// gpu.count wins over a count set in resources
	s.Equal(resource.MustParse("1"), spec.Containers[0].Resources.Limits["nvidia.com/gpu"])
}

func (s *modelServerTemplateTest) TestInTheReleaseNamespace() {
	options := &helm.Options{SetValues: minimal(map[string]string{"weights.persistence.enabled": "true"})}
	for _, template := range []string{"templates/configmap.yaml", "templates/deployment.yaml", "templates/pvc.yaml", "templates/service.yaml", "templates/networkpolicy.yaml"} {
		output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{template}, "--namespace", s.namespace)
		var object corev1.ConfigMap
		helm.UnmarshalK8SYaml(s.T(), output, &object)
		s.Equal(s.namespace, object.Namespace, template)
	}
}

func (s *modelServerTemplateTest) TestWeightsOnAClaim() {
	values := minimal(map[string]string{"weights.persistence.enabled": "true"})
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), s.render(values, "templates/deployment.yaml"), &deployment)
	var claim corev1.PersistentVolumeClaim
	helm.UnmarshalK8SYaml(s.T(), s.render(values, "templates/pvc.yaml"), &claim)

	s.Equal("model-server-weights", deployment.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
	s.Equal("model-server-weights", claim.Name)
	s.Equal(resource.MustParse("30Gi"), claim.Spec.Resources.Requests[corev1.ResourceStorage])
	s.Nil(claim.Spec.StorageClassName)
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"weights.persistence.enabled":      "true",
		"weights.persistence.storageClass": "premium-rwo",
	}), "templates/pvc.yaml"), &claim)
	s.Equal("premium-rwo", *claim.Spec.StorageClassName)

	existing := minimal(map[string]string{
		"weights.persistence.enabled":       "true",
		"weights.persistence.existingClaim": "shared-weights",
	})
	helm.UnmarshalK8SYaml(s.T(), s.render(existing, "templates/deployment.yaml"), &deployment)
	s.Equal("shared-weights", deployment.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
	_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: existing}, s.chartPath, s.releaseName, []string{"templates/pvc.yaml"})
	s.ErrorContains(err, "could not find template templates/pvc.yaml in chart")
}

func (s *modelServerTemplateTest) TestReachableFromAnotherCluster() {
	var service corev1.Service
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"service.type":                        "LoadBalancer",
		"service.loadBalancerSourceRanges[0]": "10.68.0.0/17",
		"service.annotations.networking\\.gke\\.io/load-balancer-type": "Internal",
	}), "templates/service.yaml"), &service)

	s.Equal(corev1.ServiceTypeLoadBalancer, service.Spec.Type)
	s.Equal([]string{"10.68.0.0/17"}, service.Spec.LoadBalancerSourceRanges)
	s.Equal("Internal", service.Annotations["networking.gke.io/load-balancer-type"])
	s.Equal("model-server", service.Spec.Selector["app.kubernetes.io/instance"])
	s.Equal(corev1.ServiceExternalTrafficPolicyLocal, service.Spec.ExternalTrafficPolicy)

	// The NetworkPolicy admits the same ranges, or the callers it is for
	// would be dropped
	var policy networkingv1.NetworkPolicy
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"service.type":                        "LoadBalancer",
		"service.loadBalancerSourceRanges[0]": "10.68.0.0/17",
	}), "templates/networkpolicy.yaml"), &policy)
	peers := policy.Spec.Ingress[0].From
	s.Len(peers, 2)
	s.Equal("inference", peers[0].PodSelector.MatchLabels["app"])
	s.Equal("10.68.0.0/17", peers[1].IPBlock.CIDR)

	// Load balancer callers only
	only := &helm.Options{
		SetValues: minimal(map[string]string{
			"service.type":                        "LoadBalancer",
			"service.loadBalancerSourceRanges[0]": "10.68.0.0/17",
		}),
		SetJsonValues: map[string]string{"networkPolicy.from": "[]"},
	}
	helm.UnmarshalK8SYaml(s.T(), helm.RenderTemplate(s.T(), only, s.chartPath, s.releaseName, []string{"templates/networkpolicy.yaml"}), &policy)
	s.Len(policy.Spec.Ingress[0].From, 1)
	s.Equal("10.68.0.0/17", policy.Spec.Ingress[0].From[0].IPBlock.CIDR)
}

func (s *modelServerTemplateTest) TestOnlyInClusterByDefault() {
	var service corev1.Service
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"service.loadBalancerSourceRanges[0]": "10.68.0.0/17",
	}), "templates/service.yaml"), &service)

	s.Equal(corev1.ServiceTypeClusterIP, service.Spec.Type)
	// The port the inference service's URL names
	s.Equal(int32(8080), service.Spec.Ports[0].Port)
	s.Equal("http", service.Spec.Ports[0].TargetPort.StrVal)
	s.Empty(service.Spec.LoadBalancerSourceRanges)
	s.Empty(service.Spec.ExternalTrafficPolicy)

	var policy networkingv1.NetworkPolicy
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(map[string]string{
		"service.loadBalancerSourceRanges[0]": "10.68.0.0/17",
	}), "templates/networkpolicy.yaml"), &policy)
	s.Len(policy.Spec.Ingress[0].From, 1)
}

func (s *modelServerTemplateTest) TestOnlyInferenceMayCall() {
	var policy networkingv1.NetworkPolicy
	helm.UnmarshalK8SYaml(s.T(), s.render(minimal(nil), "templates/networkpolicy.yaml"), &policy)

	s.Equal("model-server", policy.Spec.PodSelector.MatchLabels["app.kubernetes.io/instance"])
	s.Equal("inference", policy.Spec.Ingress[0].From[0].PodSelector.MatchLabels["app"])
	s.Equal(int32(8080), policy.Spec.Ingress[0].Ports[0].Port.IntVal)

	_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: minimal(map[string]string{"networkPolicy.enabled": "false"})}, s.chartPath, s.releaseName, []string{"templates/networkpolicy.yaml"})
	s.ErrorContains(err, "could not find template templates/networkpolicy.yaml in chart")
}

func (s *modelServerTemplateTest) TestRollsWhenTheSettingsChange() {
	checksum := func(values map[string]string) string {
		var deployment appsv1.Deployment
		helm.UnmarshalK8SYaml(s.T(), s.render(values, "templates/deployment.yaml"), &deployment)
		return deployment.Spec.Template.Annotations["checksum/config"]
	}

	s.NotEqual(checksum(minimal(nil)), checksum(minimal(map[string]string{"preload[1]": "qwen3-vl-embedding-2b-torch"})))
	s.Equal(checksum(minimal(nil)), checksum(minimal(map[string]string{"gpu.enabled": "true"})))
}
