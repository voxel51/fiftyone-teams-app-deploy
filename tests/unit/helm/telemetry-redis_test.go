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
	chartPath      string
	releaseName    string
	namespace      string
	deploymentTpl  string
	pvcTpl         string
	serviceTpl     string
	allBundledTpls []string
}

func TestTelemetryRedisTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	deploymentTpl := "templates/telemetry-redis-deployment.yaml"
	pvcTpl := "templates/telemetry-redis-pvc.yaml"
	serviceTpl := "templates/telemetry-redis-service.yaml"

	suite.Run(t, &telemetryRedisTemplateTest{
		Suite:          suite.Suite{},
		chartPath:      helmChartPath,
		releaseName:    "fiftyone-test",
		namespace:      "fiftyone-" + strings.ToLower(random.UniqueId()),
		deploymentTpl:  deploymentTpl,
		pvcTpl:         pvcTpl,
		serviceTpl:     serviceTpl,
		allBundledTpls: []string{deploymentTpl, pvcTpl, serviceTpl},
	})
}

func (s *telemetryRedisTemplateTest) TestEnabledByDefault() {
	options := &helm.Options{SetValues: nil}

	// Deployment + Service render unconditionally when telemetry is enabled.
	deploymentOut, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, []string{s.deploymentTpl})
	s.Require().NoError(err)
	s.Contains(deploymentOut, "kind: Deployment", "Redis Deployment should be rendered by default")

	serviceOut, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, []string{s.serviceTpl})
	s.Require().NoError(err)
	s.Contains(serviceOut, "kind: Service", "Redis Service should be rendered by default")

	// persistence.enabled defaults to false → PVC template renders empty so the
	// chart installs cleanly on clusters without a default StorageClass.
	_, err = helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, []string{s.pvcTpl})
	s.ErrorContains(err, "could not find template",
		"Redis PVC should NOT be rendered by default")
}

func (s *telemetryRedisTemplateTest) TestExplicitlyDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "false",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.allBundledTpls)
	s.ErrorContains(err, "could not find template")
}

// TestExternalUrlSkipsBundled ensures that setting telemetry.redis.external.url
// causes the bundled Redis Deployment/Service/PVC to NOT be rendered. The
// chart should leave Redis provisioning to the operator in this case.
func (s *telemetryRedisTemplateTest) TestExternalUrlSkipsBundled() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.external.url": "redis://my-managed-redis:6379",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.allBundledTpls)
	s.ErrorContains(err, "could not find template")
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

	expectedURL := fmt.Sprintf("redis://%s-telemetry-redis.%s.svc.cluster.local:6379",
		s.releaseName, "fiftyone-teams")
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

// TestExternalUrlWiresDelegatedOperatorDeployment regression-tests that the
// delegated-operator deployment's workload-container env honors external.url.
// Earlier the DO env-vars helper built the URL inline via `printf "redis://..."`
// rather than going through the `telemetry.redis.url` helper, so it always
// pointed at the in-cluster Service even when external.url was set.
func (s *telemetryRedisTemplateTest) TestExternalUrlWiresDelegatedOperatorDeployment() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.external.url":                                       "redis://my-managed-redis:6379",
		"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName,
		[]string{"templates/delegated-operator-instance-deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	// Both the workload container and the auto-injected sidecar should see
	// the external URL.
	s.Require().GreaterOrEqual(len(deployment.Spec.Template.Spec.Containers), 1,
		"Expected at least one container in DO deployment")
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

// TestExternalUrlWiresDelegatedOperatorJobConfigMap regression-tests that the
// DO Job template (rendered into the do-templates ConfigMap) honors
// external.url. Same root cause as the DO deployment bug above — the
// templates env-vars helper built the URL inline rather than going through
// `telemetry.redis.url`.
func (s *telemetryRedisTemplateTest) TestExternalUrlWiresDelegatedOperatorJobConfigMap() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.external.url":                        "redis://my-managed-redis:6379",
		"delegatedOperatorJobTemplates.jobs.test-job.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName,
		[]string{"templates/delegated-operator-job-configmap.yaml"})

	// The ConfigMap embeds the Job spec as a multi-line YAML string in
	// .data.<job-name>. The render output is plain text, so we just need to
	// ensure no occurrence of the in-cluster URL leaks through.
	inClusterURL := fmt.Sprintf("redis://%s-telemetry-redis.%s.svc.cluster.local:6379",
		s.releaseName, "fiftyone-teams")
	s.NotContains(output, inClusterURL,
		"DO Job ConfigMap must not embed the in-cluster Redis URL when external.url is set")
	s.Contains(output, "redis://my-managed-redis:6379",
		"DO Job ConfigMap should embed the external Redis URL")
}

// renderDeployment renders just the Deployment template and unmarshals it.
func (s *telemetryRedisTemplateTest) renderDeployment(options *helm.Options) appsv1.Deployment {
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{s.deploymentTpl})
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)
	return deployment
}

func (s *telemetryRedisTemplateTest) TestDeploymentMetadata() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	deployment := s.renderDeployment(options)

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

	deployment := s.renderDeployment(options)
	s.Equal("my-ns", deployment.ObjectMeta.Namespace)
}

func (s *telemetryRedisTemplateTest) TestDeploymentDefaultImage() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	deployment := s.renderDeployment(options)

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

	deployment := s.renderDeployment(options)
	s.Equal("my-registry/redis:custom", deployment.Spec.Template.Spec.Containers[0].Image)
}

func (s *telemetryRedisTemplateTest) TestDeploymentRedisArgs() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled":         "true",
		"telemetry.redis.maxmemory": "1gb",
	}}

	deployment := s.renderDeployment(options)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	s.Contains(args, "redis-server")
	s.Contains(args, "1gb", "maxmemory arg should be present")
	s.Contains(args, "allkeys-lru", "maxmemory-policy should be hardcoded to allkeys-lru")
}

func (s *telemetryRedisTemplateTest) TestServiceMetadata() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{s.serviceTpl})

	var svc corev1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &svc)
	expectedName := fmt.Sprintf("%s-telemetry-redis", s.releaseName)
	s.Equal(expectedName, svc.ObjectMeta.Name)
	s.Require().Len(svc.Spec.Ports, 1)
	s.EqualValues(6379, svc.Spec.Ports[0].Port)
}

func (s *telemetryRedisTemplateTest) TestPVCMetadata() {
	// persistence.enabled defaults to false; re-enable to exercise the PVC-rendering path.
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled":                        "true",
		"telemetry.redis.persistence.enabled":      "true",
		"telemetry.redis.persistence.size":         "5Gi",
		"telemetry.redis.persistence.storageClass": "gp3",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{s.pvcTpl})

	var pvc corev1.PersistentVolumeClaim
	helm.UnmarshalK8SYaml(s.T(), output, &pvc)
	expectedName := fmt.Sprintf("%s-telemetry-redis-data", s.releaseName)
	s.Equal(expectedName, pvc.ObjectMeta.Name)
	req := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	s.Equal("5Gi", req.String())
	s.Require().NotNil(pvc.Spec.StorageClassName, "StorageClassName should be set")
	s.Equal("gp3", *pvc.Spec.StorageClassName)
}

// assertPVCNotRendered confirms the PVC template renders empty for the given
// values — i.e. the chart does NOT create a PVC. Helm errors with "could not
// find template" when --show-only targets a template that produces no output.
func (s *telemetryRedisTemplateTest) assertPVCNotRendered(options *helm.Options, msg string) {
	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, []string{s.pvcTpl})
	s.ErrorContains(err, "could not find template", msg)
}

// TestPersistenceDisabledSkipsPVCAndUsesEmptyDir ensures that with
// persistence.enabled=false, the PVC is not rendered and the Deployment
// switches the redis-data volume to emptyDir.
func (s *telemetryRedisTemplateTest) TestPersistenceDisabledSkipsPVCAndUsesEmptyDir() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.persistence.enabled": "false",
	}}

	s.assertPVCNotRendered(options, "PVC should not render when persistence.enabled=false")

	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1, "Deployment should have exactly one volume")
	s.Equal("redis-data", volumes[0].Name)
	s.NotNil(volumes[0].EmptyDir,
		"redis-data volume should be emptyDir when persistence is disabled")
	s.Nil(volumes[0].PersistentVolumeClaim,
		"redis-data volume should NOT reference a PVC when persistence is disabled")
}

// TestExistingClaimSkipsPVCAndMountsNamedClaim ensures that
// persistence.existingClaim points the Deployment at the named PVC and
// suppresses chart-managed PVC creation — the path for clusters without
// a dynamic PV provisioner where the operator pre-creates the claim.
func (s *telemetryRedisTemplateTest) TestExistingClaimSkipsPVCAndMountsNamedClaim() {
	// persistence.enabled defaults to false; re-enable to exercise the existingClaim mount path.
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.persistence.enabled":       "true",
		"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
	}}

	s.assertPVCNotRendered(options, "PVC should not render when persistence.existingClaim is set")

	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1, "Deployment should have exactly one volume")
	s.Equal("redis-data", volumes[0].Name)
	s.Require().NotNil(volumes[0].PersistentVolumeClaim,
		"redis-data volume should reference a PVC when existingClaim is set")
	s.Equal("my-prebuilt-redis-pvc", volumes[0].PersistentVolumeClaim.ClaimName,
		"claimName should match the user-supplied existingClaim, not the chart-generated name")
	s.Nil(volumes[0].EmptyDir,
		"redis-data volume should NOT be emptyDir when existingClaim is set")
}

// TestExistingClaimIgnoresStorageClassAndSize ensures that size/storageClass
// have no effect once the user has taken over claim provisioning via
// existingClaim — the chart has no PVC to apply them to.
func (s *telemetryRedisTemplateTest) TestExistingClaimIgnoresStorageClassAndSize() {
	// persistence.enabled defaults to false; re-enable so the test exercises the existingClaim path with size/storageClass set.
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.persistence.enabled":       "true",
		"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
		"telemetry.redis.persistence.size":          "100Gi",
		"telemetry.redis.persistence.storageClass":  "gp3",
	}}

	s.assertPVCNotRendered(options, "PVC should not render when existingClaim is set, even with size/storageClass")

	// Confirm size/storageClass don't leak into the deployment either.
	deploymentOut := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{s.deploymentTpl})
	s.NotContains(deploymentOut, "100Gi",
		"size should not leak into the rendered Deployment when existingClaim is set")
	s.NotContains(deploymentOut, "storageClassName",
		"storageClassName should not appear in the Deployment when existingClaim is set")
}

// TestPersistenceDisabledBeatsExistingClaim ensures persistence.enabled=false
// remains the kill-switch — even with existingClaim set, the user opting out
// of persistence should yield an emptyDir volume rather than a stale claim
// mount.
func (s *telemetryRedisTemplateTest) TestPersistenceDisabledBeatsExistingClaim() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.persistence.enabled":       "false",
		"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
	}}

	s.assertPVCNotRendered(options, "PVC should not render when persistence.enabled=false")

	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1)
	s.NotNil(volumes[0].EmptyDir,
		"persistence.enabled=false should take precedence and yield an emptyDir volume")
	s.Nil(volumes[0].PersistentVolumeClaim,
		"redis-data volume should NOT reference the existingClaim when persistence is disabled")
}
