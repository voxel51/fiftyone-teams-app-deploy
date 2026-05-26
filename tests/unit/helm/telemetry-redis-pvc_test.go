//go:build kubeall || helm || unit || unitTelemetryRedisPVC
// +build kubeall helm unit unitTelemetryRedisPVC

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

	corev1 "k8s.io/api/core/v1"
)

type telemetryRedisPVCTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestTelemetryRedisPVCTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &telemetryRedisPVCTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/telemetry-redis-pvc.yaml"},
	})
}

// TestRenderConditions covers every input combination where the PVC template
// should and should not produce output. persistence.enabled defaults to false
// so the chart installs cleanly on clusters without a default StorageClass.
func (s *telemetryRedisPVCTemplateTest) TestRenderConditions() {
	testCases := []struct {
		name    string
		values  map[string]string
		renders bool
	}{
		{
			// Defaults: persistence.enabled=false → PVC suppressed.
			"defaultValues",
			nil,
			false,
		},
		{
			"telemetryDisabled",
			map[string]string{"telemetry.enabled": "false"},
			false,
		},
		{
			// External Redis URL → bundled PVC suppressed.
			"externalUrlSet",
			map[string]string{"telemetry.redis.external.url": "redis://my-managed-redis:6379"},
			false,
		},
		{
			"persistenceExplicitlyDisabled",
			map[string]string{"telemetry.redis.persistence.enabled": "false"},
			false,
		},
		{
			// existingClaim → user manages the PVC, chart suppresses its own.
			"existingClaimSet",
			map[string]string{
				"telemetry.redis.persistence.enabled":       "true",
				"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
			},
			false,
		},
		{
			// size/storageClass do not change suppression behavior when
			// existingClaim is set — user still owns the PVC.
			"existingClaimWithSizeAndStorageClass",
			map[string]string{
				"telemetry.redis.persistence.enabled":       "true",
				"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
				"telemetry.redis.persistence.size":          "100Gi",
				"telemetry.redis.persistence.storageClass":  "gp3",
			},
			false,
		},
		{
			// persistence.enabled=false is the kill switch — even with
			// existingClaim set, no PVC is rendered.
			"persistenceDisabledBeatsExistingClaim",
			map[string]string{
				"telemetry.redis.persistence.enabled":       "false",
				"telemetry.redis.persistence.existingClaim": "my-prebuilt-redis-pvc",
			},
			false,
		},
		{
			"persistenceEnabled",
			map[string]string{"telemetry.redis.persistence.enabled": "true"},
			true,
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
				s.Contains(output, "kind: PersistentVolumeClaim", "PVC should be rendered")
			} else {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template")
			}
		})
	}
}

func (s *telemetryRedisPVCTemplateTest) TestMetadata() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(pvc corev1.PersistentVolumeClaim)
	}{
		{
			"persistenceEnabledWithSizeAndStorageClass",
			map[string]string{
				"telemetry.redis.persistence.enabled":      "true",
				"telemetry.redis.persistence.size":         "5Gi",
				"telemetry.redis.persistence.storageClass": "gp3",
			},
			func(pvc corev1.PersistentVolumeClaim) {
				expectedName := fmt.Sprintf("%s-telemetry-redis-data", s.releaseName)
				s.Equal(expectedName, pvc.ObjectMeta.Name)
				req := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				s.Equal("5Gi", req.String())
				s.Require().NotNil(pvc.Spec.StorageClassName, "StorageClassName should be set")
				s.Equal("gp3", *pvc.Spec.StorageClassName)
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

			var pvc corev1.PersistentVolumeClaim
			helm.UnmarshalK8SYaml(subT, output, &pvc)

			testCase.expected(pvc)
		})
	}
}
