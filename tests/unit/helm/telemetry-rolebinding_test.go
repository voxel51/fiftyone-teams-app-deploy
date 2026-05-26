//go:build kubeall || helm || unit || unitTelemetryRoleBinding
// +build kubeall helm unit unitTelemetryRoleBinding

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

type telemetryRoleBindingTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestTelemetryRoleBindingTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &telemetryRoleBindingTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/telemetry-rolebinding.yaml"},
	})
}

func (s *telemetryRoleBindingTemplateTest) TestRenderConditions() {
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
				s.Contains(output, "kind: RoleBinding", "RoleBinding should be rendered")
			} else {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template")
			}
		})
	}
}

func (s *telemetryRoleBindingTemplateTest) TestSubjects() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(subjects []rbacv1.Subject)
	}{
		{
			// With an unset serviceAccounts list, the binding falls back to
			// the chart's main app service account — the SA used by sidecars
			// on fiftyone-app, teams-plugins, and delegated-operator
			// workloads. The teams-api SA is intentionally NOT bound here;
			// api-role.yaml already grants it pods/log GET.
			"defaultValues",
			nil,
			func(subjects []rbacv1.Subject) {
				s.Require().Len(subjects, 1, "Default RoleBinding should bind only to the main app SA")
				s.Equal("ServiceAccount", subjects[0].Kind)
				s.Equal("fiftyone-teams", subjects[0].Namespace)
				s.Equal("fiftyone-teams", subjects[0].Name)
			},
		},
		{
			"overrideServiceAccountsAndNamespace",
			map[string]string{
				"telemetry.serviceAccounts[0]": "sa-one",
				"telemetry.serviceAccounts[1]": "sa-two",
				"namespace.name":               "my-ns",
			},
			func(subjects []rbacv1.Subject) {
				s.Require().Len(subjects, 2)
				s.Equal("sa-one", subjects[0].Name)
				s.Equal("my-ns", subjects[0].Namespace)
				s.Equal("sa-two", subjects[1].Name)
				s.Equal("my-ns", subjects[1].Namespace)
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

			var rb rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(subT, output, &rb)

			testCase.expected(rb.Subjects)
		})
	}
}

func (s *telemetryRoleBindingTemplateTest) TestRoleRef() {
	options := &helm.Options{SetValues: nil}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var rb rbacv1.RoleBinding
	helm.UnmarshalK8SYaml(s.T(), output, &rb)

	expectedName := fmt.Sprintf("%s-telemetry-pod-logs", s.releaseName)
	s.Equal("Role", rb.RoleRef.Kind)
	s.Equal(expectedName, rb.RoleRef.Name)
	s.Equal("rbac.authorization.k8s.io", rb.RoleRef.APIGroup)
}
