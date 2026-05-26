//go:build kubeall || helm || unit || unitTelemetryRedisService
// +build kubeall helm unit unitTelemetryRedisService

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

type telemetryRedisServiceTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestTelemetryRedisServiceTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &telemetryRedisServiceTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/telemetry-redis-service.yaml"},
	})
}

func (s *telemetryRedisServiceTemplateTest) TestRenderConditions() {
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
			// External Redis URL takes over; chart skips the bundled Service.
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
				s.Contains(output, "kind: Service", "Redis Service should be rendered")
			} else {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template")
			}
		})
	}
}

func (s *telemetryRedisServiceTemplateTest) TestMetadata() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(svc corev1.Service)
	}{
		{
			"defaultValues",
			nil,
			func(svc corev1.Service) {
				expectedName := fmt.Sprintf("%s-telemetry-redis", s.releaseName)
				s.Equal(expectedName, svc.ObjectMeta.Name)
				s.Equal("fiftyone-teams", svc.ObjectMeta.Namespace)
				s.Equal("telemetry-redis", svc.ObjectMeta.Labels["app.kubernetes.io/name"])
				s.Equal(s.releaseName, svc.ObjectMeta.Labels["app.kubernetes.io/instance"])
				s.Equal("telemetry-redis", svc.ObjectMeta.Labels["app.voxel51.com/component"])
			},
		},
		{
			"overrideNamespace",
			map[string]string{"namespace.name": "my-ns"},
			func(svc corev1.Service) {
				s.Equal("my-ns", svc.ObjectMeta.Namespace)
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

			var svc corev1.Service
			helm.UnmarshalK8SYaml(subT, output, &svc)

			testCase.expected(svc)
		})
	}
}

func (s *telemetryRedisServiceTemplateTest) TestPort() {
	options := &helm.Options{SetValues: nil}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var svc corev1.Service
	helm.UnmarshalK8SYaml(s.T(), output, &svc)

	s.Require().Len(svc.Spec.Ports, 1)
	s.EqualValues(6379, svc.Spec.Ports[0].Port)
}
