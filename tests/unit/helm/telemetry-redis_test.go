//go:build kubeall || helm || unit || unitTelemetryRedis
// +build kubeall helm unit unitTelemetryRedis

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type telemetryRedisTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestTelemetryRedisTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &telemetryRedisTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/telemetry-redis.yaml"},
	})
}

func (s *telemetryRedisTemplateTest) TestEnabledByDefault() {
	options := &helm.Options{SetValues: nil}

	output, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.Require().NoError(err)
	s.Contains(output, "kind: Deployment", "Redis Deployment should be rendered by default")
	s.Contains(output, "kind: Service", "Redis Service should be rendered by default")
	s.Contains(output, "kind: PersistentVolumeClaim", "Redis PVC should be rendered by default")
}

func (s *telemetryRedisTemplateTest) TestExplicitlyDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "false",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template templates/telemetry-redis.yaml in chart")
}

// TestExternalUrlSkipsBundled ensures that setting telemetry.redis.external.url
// causes the bundled Redis Deployment/Service/PVC to NOT be rendered. The
// chart should leave Redis provisioning to the operator in this case.
func (s *telemetryRedisTemplateTest) TestExternalUrlSkipsBundled() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.external.url": "redis://my-managed-redis:6379",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template templates/telemetry-redis.yaml in chart")
}

// TestExternalUrlWiresApiDeployment ensures that setting external.url causes
// FIFTYONE_TELEMETRY_REDIS_URL on the api deployment to point at the external
// URL rather than the in-cluster Service.
func (s *telemetryRedisTemplateTest) TestExternalUrlWiresApiDeployment() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.external.url": "redis://my-managed-redis:6379",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName,
		[]string{"templates/api-deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)
	s.Require().Len(deployment.Spec.Template.Spec.Containers, 2,
		"Expected api + telemetry-sidecar containers")

	for _, container := range deployment.Spec.Template.Spec.Containers {
		var found *corev1.EnvVar
		for i, ev := range container.Env {
			if ev.Name == "FIFTYONE_TELEMETRY_REDIS_URL" {
				found = &container.Env[i]
				break
			}
		}
		s.Require().NotNil(found,
			"FIFTYONE_TELEMETRY_REDIS_URL should be set on %s container", container.Name)
		s.Equal("redis://my-managed-redis:6379", found.Value,
			"FIFTYONE_TELEMETRY_REDIS_URL on %s should point at the external URL",
			container.Name)
	}
}

// TestBundledUrlWiresApiDeployment ensures the default in-cluster URL is
// release-scoped on the api deployment's containers.
func (s *telemetryRedisTemplateTest) TestBundledUrlWiresApiDeployment() {
	options := &helm.Options{SetValues: nil}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName,
		[]string{"templates/api-deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	expectedURL := fmt.Sprintf("redis://%s-telemetry-redis:6379", s.releaseName)
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, ev := range container.Env {
			if ev.Name == "FIFTYONE_TELEMETRY_REDIS_URL" {
				s.Equal(expectedURL, ev.Value,
					"FIFTYONE_TELEMETRY_REDIS_URL on %s should be release-scoped in-cluster URL",
					container.Name)
			}
		}
	}
}

// extractDeployment finds the Deployment document in the multi-doc render output.
func (s *telemetryRedisTemplateTest) extractDeployment(output string) appsv1.Deployment {
	for _, doc := range splitYAMLDocs(output) {
		if !strings.Contains(doc, "kind: Deployment") {
			continue
		}
		var deployment appsv1.Deployment
		helm.UnmarshalK8SYaml(s.T(), doc, &deployment)
		return deployment
	}
	s.Fail("Deployment document not found in rendered output")
	return appsv1.Deployment{}
}

func (s *telemetryRedisTemplateTest) TestDeploymentMetadata() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	deployment := s.extractDeployment(output)

	expectedName := fmt.Sprintf("%s-telemetry-redis", s.releaseName)
	s.Equal(expectedName, deployment.ObjectMeta.Name, "Deployment name should be release-prefixed")
	s.Equal("fiftyone-teams", deployment.ObjectMeta.Namespace, "Deployment namespace should default to fiftyone-teams")
	s.Equal("telemetry-redis", deployment.ObjectMeta.Labels["app.kubernetes.io/name"])
	s.Equal(s.releaseName, deployment.ObjectMeta.Labels["app.kubernetes.io/instance"])
	s.Equal("telemetry-redis", deployment.ObjectMeta.Labels["app.voxel51.com/component"])
}

func (s *telemetryRedisTemplateTest) TestDeploymentNamespaceOverride() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
		"namespace.name":    "my-ns",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	deployment := s.extractDeployment(output)
	s.Equal("my-ns", deployment.ObjectMeta.Namespace)
}

func (s *telemetryRedisTemplateTest) TestDeploymentDefaultImage() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	deployment := s.extractDeployment(output)

	containers := deployment.Spec.Template.Spec.Containers
	s.Require().Len(containers, 1, "Deployment should have exactly one container")
	s.Equal("redis", containers[0].Name)
	s.Equal("redis:7-alpine", containers[0].Image)
}

func (s *telemetryRedisTemplateTest) TestDeploymentImageOverride() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled":     "true",
		"telemetry.redis.image": "my-registry/redis:custom",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	deployment := s.extractDeployment(output)
	s.Equal("my-registry/redis:custom", deployment.Spec.Template.Spec.Containers[0].Image)
}

func (s *telemetryRedisTemplateTest) TestDeploymentRedisArgs() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled":               "true",
		"telemetry.redis.maxmemory":       "1gb",
		"telemetry.redis.maxmemoryPolicy": "noeviction",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	deployment := s.extractDeployment(output)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	s.Contains(args, "redis-server")
	s.Contains(args, "1gb", "maxmemory arg should be present")
	s.Contains(args, "noeviction", "maxmemory-policy arg should be present")
}

func (s *telemetryRedisTemplateTest) TestServiceMetadata() {
	// The render output contains PVC + Deployment + Service. Use a Service-only
	// unmarshal by isolating the Service doc.
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	for _, doc := range splitYAMLDocs(output) {
		if !strings.Contains(doc, "kind: Service") {
			continue
		}
		var svc corev1.Service
		helm.UnmarshalK8SYaml(s.T(), doc, &svc)
		expectedName := fmt.Sprintf("%s-telemetry-redis", s.releaseName)
		s.Equal(expectedName, svc.ObjectMeta.Name)
		s.Require().Len(svc.Spec.Ports, 1)
		s.EqualValues(6379, svc.Spec.Ports[0].Port)
		return
	}
	s.Fail("Service document not found in rendered output")
}

func (s *telemetryRedisTemplateTest) TestPVCMetadata() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled":       "true",
		"telemetry.redis.storage": "5Gi",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	for _, doc := range splitYAMLDocs(output) {
		if !strings.Contains(doc, "kind: PersistentVolumeClaim") {
			continue
		}
		var pvc corev1.PersistentVolumeClaim
		helm.UnmarshalK8SYaml(s.T(), doc, &pvc)
		expectedName := fmt.Sprintf("%s-telemetry-redis-data", s.releaseName)
		s.Equal(expectedName, pvc.ObjectMeta.Name)
		req := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		s.Equal("5Gi", req.String())
		return
	}
	s.Fail("PVC document not found in rendered output")
}
