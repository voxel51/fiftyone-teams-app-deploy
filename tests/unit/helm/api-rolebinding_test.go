//go:build kubeall || helm || unit || unitRoleBinding
// +build kubeall helm unit unitRoleBinding

package unit

import (
	"encoding/json"
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

type apiRoleBindingTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestAPIRoleBindingTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &apiRoleBindingTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/api-rolebinding.yaml"},
	})
}

func (s *apiRoleBindingTemplateTest) TestDisabled() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			fmt.Sprintf("%s-fiftyone-teams-app-do-management", s.releaseName),
		},
		{
			"overrideRbacDisabled",
			map[string]string{
				"apiSettings.rbac.create": "false",
			},
			"",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			if testCase.expected == "" {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/api-rolebinding.yaml in chart")

				var roleBinding rbacv1.RoleBinding
				helm.UnmarshalK8SYaml(subT, output, &roleBinding)

				s.Empty(roleBinding.ObjectMeta.Name, "Name should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var roleBinding rbacv1.RoleBinding
				helm.UnmarshalK8SYaml(subT, output, &roleBinding)

				s.Equal(testCase.expected, roleBinding.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

func (s *apiRoleBindingTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			fmt.Sprintf("%s-fiftyone-teams-app-do-management", s.releaseName),
		},
		{
			"overrideName",
			map[string]string{
				"apiSettings.rbac.roleBinding.name": "test-role-binding-name",
			},
			"test-role-binding-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var roleBinding rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(subT, output, &roleBinding)

			s.Equal(testCase.expected, roleBinding.ObjectMeta.Name, "Name name should be equal.")
		})
	}
}

func (s *apiRoleBindingTemplateTest) TestMetadataNamespace() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-teams",
		},
		{
			"overrideNamespace",
			map[string]string{
				"namespace.name": "test-namespace-name",
			},
			"test-namespace-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var roleBinding rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(subT, output, &roleBinding)

			s.Equal(testCase.expected, roleBinding.ObjectMeta.Namespace, "Namespace name should be equal.")
		})
	}
}

func (s *apiRoleBindingTemplateTest) TestMetadataAnnotations() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			nil,
			nil,
		},
		{
			"overrideAnnotations",
			map[string]string{
				"apiSettings.rbac.roleBinding.annotations.annotation-1": "annotation-1-value",
			},
			map[string]string{
				"annotation-1": "annotation-1-value",
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

			var roleBinding rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(subT, output, &roleBinding)

			if testCase.expected == nil {
				s.Nil(roleBinding.ObjectMeta.Annotations, "Annotations should be nil")
			} else {
				for key, value := range testCase.expected {
					foundValue := roleBinding.ObjectMeta.Annotations[key]
					s.Equal(value, foundValue, "Annotations should contain all set annotations.")
				}
			}
		})
	}
}

func (s *apiRoleBindingTemplateTest) TestMetadataLabels() {
	// Get chart info (to later obtain the chart's appVersion)
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)

	// Get appVersion from chart info
	chartAppVersion, exists := cInfo["appVersion"]
	s.True(exists, "failed to get app version from chart info")

	// Get version from chart info
	chartVersion, exists := cInfo["version"]
	s.True(exists, "failed to get version from chart info")

	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			nil,
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       fmt.Sprintf("%s-fiftyone-teams-app-do-management", s.releaseName),
				"app.kubernetes.io/instance":   "fiftyone-test",
				"app.voxel51.com/component":    "on-demand-delegated-operators",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
				"apiSettings.rbac.roleBinding.name":         "test-service-account-name",
				"apiSettings.rbac.roleBinding.labels.color": "blue",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "test-service-account-name",
				"app.kubernetes.io/instance":   "fiftyone-test",
				"app.voxel51.com/component":    "on-demand-delegated-operators",
				"color":                        "blue",
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

			var roleBinding rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(subT, output, &roleBinding)

			for key, value := range testCase.expected {
				foundValue := roleBinding.ObjectMeta.Labels[key]
				s.Equal(value, foundValue, "Labels should contain all set labels.")
			}
		})
	}
}

func (s *apiRoleBindingTemplateTest) TestSubjects() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(subjects []rbacv1.Subject)
	}{
		{
			"defaultValues",
			nil,
			func(subjects []rbacv1.Subject) {
				expectedSubjectsJson := fmt.Sprintf(`[
          {
            "kind": "ServiceAccount",
            "name": "%s-fiftyone-teams-app-teams-api",
            "namespace": "fiftyone-teams"
          }
        ]`, s.releaseName)
				var expectedSubjects []rbacv1.Subject
				err := json.Unmarshal([]byte(expectedSubjectsJson), &expectedSubjects)
				s.NoError(err)
				s.Equal(expectedSubjects, subjects, "Subjects should be equal")
			},
		},
		{
			"overrideNamespace",
			map[string]string{
				"namespace.name": "test-namespace-name",
			},
			func(subjects []rbacv1.Subject) {
				expectedSubjectsJson := fmt.Sprintf(`[
          {
            "kind": "ServiceAccount",
            "name": "%s-fiftyone-teams-app-teams-api",
            "namespace": "test-namespace-name"
          }
        ]`, s.releaseName)
				var expectedSubjects []rbacv1.Subject
				err := json.Unmarshal([]byte(expectedSubjectsJson), &expectedSubjects)
				s.NoError(err)
				s.Equal(expectedSubjects, subjects, "Subjects should be equal")
			},
		},
		{
			"overrideServiceAccountName",
			map[string]string{
				"apiSettings.rbac.serviceAccount.name": "test-service-account-name",
			},
			func(subjects []rbacv1.Subject) {
				expectedSubjectsJson := `[
          {
            "kind": "ServiceAccount",
            "name": "test-service-account-name",
            "namespace": "fiftyone-teams"
          }
        ]`
				var expectedSubjects []rbacv1.Subject
				err := json.Unmarshal([]byte(expectedSubjectsJson), &expectedSubjects)
				s.NoError(err)
				s.Equal(expectedSubjects, subjects, "Subjects should be equal")
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

			var roleBinding rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(subT, output, &roleBinding)

			testCase.expected(roleBinding.Subjects)
		})
	}
}

func (s *apiRoleBindingTemplateTest) TestRoleRef() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected func(subjects rbacv1.RoleRef)
	}{
		{
			"defaultValues",
			nil,
			func(roleRef rbacv1.RoleRef) {
				expectedRoleRefJson := fmt.Sprintf(`{
            "kind": "Role",
            "name": "%s-fiftyone-teams-app-do-management",
            "apiGroup": "rbac.authorization.k8s.io"
          }`, s.releaseName)
				var expectedRoleRef rbacv1.RoleRef
				err := json.Unmarshal([]byte(expectedRoleRefJson), &expectedRoleRef)
				s.NoError(err)
				s.Equal(expectedRoleRef, roleRef, "RoleRef should be equal")
			},
		},
		{
			"overrideRoleName",
			map[string]string{
				"apiSettings.rbac.role.name": "test-role-name",
			},
			func(roleRef rbacv1.RoleRef) {
				expectedRoleRefJson := `{
            "kind": "Role",
            "name": "test-role-name",
            "apiGroup": "rbac.authorization.k8s.io"
          }`
				var expectedRoleRef rbacv1.RoleRef
				err := json.Unmarshal([]byte(expectedRoleRefJson), &expectedRoleRef)
				s.NoError(err)
				s.Equal(expectedRoleRef, roleRef, "RoleRef should be equal")
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

			var roleBinding rbacv1.RoleBinding
			helm.UnmarshalK8SYaml(subT, output, &roleBinding)

			testCase.expected(roleBinding.RoleRef)
		})
	}
}
