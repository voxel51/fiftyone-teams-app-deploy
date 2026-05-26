//go:build kubeall || helm || unit || unitTelemetryRedisDeployment
// +build kubeall helm unit unitTelemetryRedisDeployment

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
)

type telemetryRedisDeploymentTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestTelemetryRedisDeploymentTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &telemetryRedisDeploymentTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/telemetry-redis-deployment.yaml"},
	})
}

func (s *telemetryRedisDeploymentTemplateTest) TestRenderConditions() {
	testCases := []struct {
		name    string
		values  map[string]string
		renders bool
	}{
		{
			"defaultValues",
			nil,
			true,
		},
		{
			"telemetryDisabled",
			map[string]string{"telemetry.enabled": "false"},
			false,
		},
		{
			// External Redis URL takes over; chart skips the bundled Deployment.
			"externalUrlSet",
			map[string]string{"telemetry.redis.external.url": "redis://my-managed-redis:6379"},
			false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.renders {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)
				s.Contains(output, "kind: Deployment", "Redis Deployment should be rendered")
			} else {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template")
			}
		})
	}
}

func (s *telemetryRedisDeploymentTemplateTest) TestMetadata() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(deployment appsv1.Deployment)
	}{
		{
			"defaultValues",
			nil,
			func(deployment appsv1.Deployment) {
				expectedName := fmt.Sprintf("%s-telemetry-redis", s.releaseName)
				s.Equal(expectedName, deployment.ObjectMeta.Name, "Deployment name should be release-prefixed")
				s.Equal("fiftyone-teams", deployment.ObjectMeta.Namespace, "Deployment namespace should default to fiftyone-teams")
				s.Equal("telemetry-redis", deployment.ObjectMeta.Labels["app.kubernetes.io/name"])
				s.Equal(s.releaseName, deployment.ObjectMeta.Labels["app.kubernetes.io/instance"])
				s.Equal("telemetry-redis", deployment.ObjectMeta.Labels["app.voxel51.com/component"])
			},
		},
		{
			"overrideNamespace",
			map[string]string{"namespace.name": "my-ns"},
			func(deployment appsv1.Deployment) {
				s.Equal("my-ns", deployment.ObjectMeta.Namespace)
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

			testCase.expected(deployment)
		})
	}
}

func (s *telemetryRedisDeploymentTemplateTest) TestContainerImage() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"redis:7-alpine",
		},
		{
			"overrideImage",
			map[string]string{"telemetry.redis.image": "my-registry/redis:custom"},
			"my-registry/redis:custom",
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

			containers := deployment.Spec.Template.Spec.Containers
			s.Require().Len(containers, 1, "Deployment should have exactly one container")
			s.Equal("redis", containers[0].Name)
			s.Equal(testCase.expected, containers[0].Image)
		})
	}
}

func (s *telemetryRedisDeploymentTemplateTest) TestContainerArgs() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(args []string)
	}{
		{
			"defaultMaxmemory",
			nil,
			func(args []string) {
				s.Contains(args, "redis-server", "redis-server arg should be present")
				s.Contains(args, "400mb", "default maxmemory should be 400mb")
				s.Contains(args, "allkeys-lru", "maxmemory-policy should be hardcoded to allkeys-lru")
			},
		},
		{
			"overrideMaxmemory",
			map[string]string{"telemetry.redis.maxmemory": "1gb"},
			func(args []string) {
				s.Contains(args, "redis-server")
				s.Contains(args, "1gb", "maxmemory arg should reflect override")
				s.Contains(args, "allkeys-lru")
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

			testCase.expected(deployment.Spec.Template.Spec.Containers[0].Args)
		})
	}
}

// TestVolumes asserts that the redis-data volume tracks persistence settings:
// emptyDir when persistence is disabled, the user-supplied PVC when
// existingClaim is set, and persistence.enabled=false beats existingClaim.
func (s *telemetryRedisDeploymentTemplateTest) TestVolumes() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(deployment appsv1.Deployment)
	}{
		{
			"persistenceDisabledUsesEmptyDir",
			map[string]string{"telemetry.redis.persistence.enabled": "false"},
			func(deployment appsv1.Deployment) {
				volumes := deployment.Spec.Template.Spec.Volumes
				s.Require().Len(volumes, 1, "Deployment should have exactly one volume")
				s.Equal("redis-data", volumes[0].Name)
				s.NotNil(volumes[0].EmptyDir, "redis-data volume should be emptyDir when persistence is disabled")
				s.Nil(volumes[0].PersistentVolumeClaim, "redis-data volume should NOT reference a PVC when persistence is disabled")
			},
		},
		{
			"existingClaimMountsNamedClaim",
			map[string]string{
				"telemetry.redis.persistence.enabled":       "true",
				"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
			},
			func(deployment appsv1.Deployment) {
				volumes := deployment.Spec.Template.Spec.Volumes
				s.Require().Len(volumes, 1, "Deployment should have exactly one volume")
				s.Equal("redis-data", volumes[0].Name)
				s.Require().NotNil(volumes[0].PersistentVolumeClaim, "redis-data volume should reference a PVC when existingClaim is set")
				s.Equal("my-prebuilt-redis-pvc", volumes[0].PersistentVolumeClaim.ClaimName,
					"claimName should match the user-supplied existingClaim, not the chart-generated name")
				s.Nil(volumes[0].EmptyDir, "redis-data volume should NOT be emptyDir when existingClaim is set")
			},
		},
		{
			// size/storageClass have no effect once the user has taken over
			// claim provisioning via existingClaim — the chart has no PVC to
			// apply them to and they must not leak into the Deployment.
			"existingClaimIgnoresSizeAndStorageClass",
			map[string]string{
				"telemetry.redis.persistence.enabled":       "true",
				"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
				"telemetry.redis.persistence.size":          "100Gi",
				"telemetry.redis.persistence.storageClass":  "gp3",
			},
			func(deployment appsv1.Deployment) {
				volumes := deployment.Spec.Template.Spec.Volumes
				s.Require().Len(volumes, 1)
				s.Require().NotNil(volumes[0].PersistentVolumeClaim)
				s.Equal("my-prebuilt-redis-pvc", volumes[0].PersistentVolumeClaim.ClaimName)
			},
		},
		{
			// persistence.enabled=false is the kill-switch — even with
			// existingClaim set, opting out of persistence yields emptyDir.
			"persistenceDisabledBeatsExistingClaim",
			map[string]string{
				"telemetry.redis.persistence.enabled":       "false",
				"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
			},
			func(deployment appsv1.Deployment) {
				volumes := deployment.Spec.Template.Spec.Volumes
				s.Require().Len(volumes, 1)
				s.NotNil(volumes[0].EmptyDir,
					"persistence.enabled=false should take precedence and yield an emptyDir volume")
				s.Nil(volumes[0].PersistentVolumeClaim,
					"redis-data volume should NOT reference the existingClaim when persistence is disabled")
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

			testCase.expected(deployment)
		})
	}
}

// TestStorageClassDoesNotLeakWhenExistingClaim asserts that size/storageClass
// values do not appear in the rendered Deployment when existingClaim is set —
// they belong on the PVC, which is suppressed in this mode.
func (s *telemetryRedisDeploymentTemplateTest) TestStorageClassDoesNotLeakWhenExistingClaim() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.redis.persistence.enabled":       "true",
		"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
		"telemetry.redis.persistence.size":          "100Gi",
		"telemetry.redis.persistence.storageClass":  "gp3",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.NotContains(output, "100Gi",
		"size should not leak into the rendered Deployment when existingClaim is set")
	s.NotContains(output, "storageClassName",
		"storageClassName should not appear in the Deployment when existingClaim is set")
}
