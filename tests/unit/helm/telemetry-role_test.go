//go:build kubeall || helm || unit || unitTelemetryRole
// +build kubeall helm unit unitTelemetryRole

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

	rbacv1 "k8s.io/api/rbac/v1"
)

type telemetryRoleTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestTelemetryRoleTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &telemetryRoleTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/telemetry-role.yaml"},
	})
}

func (s *telemetryRoleTemplateTest) TestRenderConditions() {
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
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.renders {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)
				s.Contains(output, "kind: Role", "Role should be rendered")
			} else {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template")
			}
		})
	}
}

func (s *telemetryRoleTemplateTest) TestMetadata() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(role rbacv1.Role)
	}{
		{
			"defaultValues",
			nil,
			func(role rbacv1.Role) {
				expectedName := fmt.Sprintf("%s-telemetry-pod-logs", s.releaseName)
				s.Equal(expectedName, role.ObjectMeta.Name)
				s.Equal("fiftyone-teams", role.ObjectMeta.Namespace)
				s.Equal("telemetry", role.ObjectMeta.Labels["app.kubernetes.io/name"])
				s.Equal(s.releaseName, role.ObjectMeta.Labels["app.kubernetes.io/instance"])
				s.Equal("telemetry", role.ObjectMeta.Labels["app.voxel51.com/component"])
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

			var role rbacv1.Role
			helm.UnmarshalK8SYaml(subT, output, &role)

			testCase.expected(role)
		})
	}
}

func (s *telemetryRoleTemplateTest) TestRules() {
	options := &helm.Options{SetValues: nil}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var role rbacv1.Role
	helm.UnmarshalK8SYaml(s.T(), output, &role)

	s.Require().Len(role.Rules, 2, "Role should have exactly two rules")

	var hasPodsGet, hasPodsLogGet bool
	for _, rule := range role.Rules {
		for _, resource := range rule.Resources {
			if resource == "pods" {
				s.Contains(rule.Verbs, "get", "pods rule should grant get")
				hasPodsGet = true
			}
			if resource == "pods/log" {
				s.Contains(rule.Verbs, "get", "pods/log rule should grant get")
				hasPodsLogGet = true
			}
		}
	}
	s.True(hasPodsGet, "Role should grant get on pods")
	s.True(hasPodsLogGet, "Role should grant get on pods/log")
}
