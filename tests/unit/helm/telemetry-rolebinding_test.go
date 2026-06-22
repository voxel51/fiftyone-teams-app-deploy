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
		templates: []string{
			"templates/telemetry-role.yaml",
			"templates/telemetry-rolebinding.yaml",
		},
	})
}

func (s *telemetryRoleBindingTemplateTest) TestEnabledByDefault() {
	options := &helm.Options{SetValues: nil}

	output, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.Require().NoError(err)
	s.Contains(output, "kind: Role", "Role should be rendered by default")
	s.Contains(output, "kind: RoleBinding", "RoleBinding should be rendered by default")
}

func (s *telemetryRoleBindingTemplateTest) TestExplicitlyDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "false",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

// With telemetry enabled but rbac.create false, the Role/RoleBinding are
// skipped so an install identity without namespaced RBAC permissions can still
// deploy telemetry. Metrics keep working (shared process namespace), but the
// sidecar loses pods/log API access, so the log viewer stays empty unless
// serviceAccounts points at an SA that already carries pods/log.
func (s *telemetryRoleBindingTemplateTest) TestRbacCreateDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled":     "true",
		"telemetry.rbac.create": "false",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

// extractRole finds the Role document (not RoleBinding) in multi-doc render output.
func (s *telemetryRoleBindingTemplateTest) extractRole(output string) (rbacv1.Role, bool) {
	for _, doc := range splitYAMLDocs(output) {
		if !strings.Contains(doc, "\nkind: Role\n") {
			continue
		}
		var role rbacv1.Role
		helm.UnmarshalK8SYaml(s.T(), doc, &role)
		return role, true
	}
	return rbacv1.Role{}, false
}

// extractRoleBinding finds the RoleBinding document in multi-doc render output.
func (s *telemetryRoleBindingTemplateTest) extractRoleBinding(output string) (rbacv1.RoleBinding, bool) {
	for _, doc := range splitYAMLDocs(output) {
		if !strings.Contains(doc, "kind: RoleBinding") {
			continue
		}
		var rb rbacv1.RoleBinding
		helm.UnmarshalK8SYaml(s.T(), doc, &rb)
		return rb, true
	}
	return rbacv1.RoleBinding{}, false
}

func (s *telemetryRoleBindingTemplateTest) TestRoleMetadata() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	role, ok := s.extractRole(output)
	s.Require().True(ok, "Role document not found in rendered output")

	expectedName := fmt.Sprintf("%s-telemetry-pod-logs", s.releaseName)
	s.Equal(expectedName, role.ObjectMeta.Name)
	s.Equal("fiftyone-teams", role.ObjectMeta.Namespace)
	s.Equal("telemetry", role.ObjectMeta.Labels["app.kubernetes.io/name"])
	s.Equal(s.releaseName, role.ObjectMeta.Labels["app.kubernetes.io/instance"])
	s.Equal("telemetry", role.ObjectMeta.Labels["app.voxel51.com/component"])
}

func (s *telemetryRoleBindingTemplateTest) TestRoleRules() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	role, ok := s.extractRole(output)
	s.Require().True(ok, "Role document not found in rendered output")

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

func (s *telemetryRoleBindingTemplateTest) TestRoleBindingDefaultSubject() {
	// With an unset serviceAccounts list, the binding falls back to the
	// chart's main app service account — the SA used by sidecars on
	// fiftyone-app, teams-plugins, and delegated-operator workloads. The
	// teams-api SA is intentionally NOT bound here; api-role.yaml already
	// grants it pods/log GET.
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	rb, ok := s.extractRoleBinding(output)
	s.Require().True(ok, "RoleBinding document not found in rendered output")

	s.Require().Len(rb.Subjects, 1, "Default RoleBinding should bind only to the main app SA")
	s.Equal("ServiceAccount", rb.Subjects[0].Kind)
	s.Equal("fiftyone-teams", rb.Subjects[0].Namespace)
	// Main app SA: chart default serviceAccount.name = "fiftyone-teams"
	s.Equal("fiftyone-teams", rb.Subjects[0].Name)
}

func (s *telemetryRoleBindingTemplateTest) TestRoleBindingMultipleSubjects() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled":            "true",
		"telemetry.serviceAccounts[0]": "sa-one",
		"telemetry.serviceAccounts[1]": "sa-two",
		"namespace.name":               "my-ns",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	rb, ok := s.extractRoleBinding(output)
	s.Require().True(ok, "RoleBinding document not found in rendered output")

	s.Require().Len(rb.Subjects, 2)
	s.Equal("sa-one", rb.Subjects[0].Name)
	s.Equal("my-ns", rb.Subjects[0].Namespace)
	s.Equal("sa-two", rb.Subjects[1].Name)
	s.Equal("my-ns", rb.Subjects[1].Namespace)
}

func (s *telemetryRoleBindingTemplateTest) TestRoleBindingRoleRef() {
	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
	}}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	rb, ok := s.extractRoleBinding(output)
	s.Require().True(ok, "RoleBinding document not found in rendered output")

	expectedName := fmt.Sprintf("%s-telemetry-pod-logs", s.releaseName)
	s.Equal("Role", rb.RoleRef.Kind)
	s.Equal(expectedName, rb.RoleRef.Name)
	s.Equal("rbac.authorization.k8s.io", rb.RoleRef.APIGroup)
}
