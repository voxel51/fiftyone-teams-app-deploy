//go:build kubeall || helm || unit || unitNamespace
// +build kubeall helm unit unitNamespace

package unit

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corev1 "k8s.io/api/core/v1"
)

type namespaceTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestNamespaceTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &namespaceTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/namespace.yaml"},
	})
}

func (s *namespaceTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"",
		},
		{
			"defaultValuesWithNamespaceCreateEnabled",
			map[string]string{
				"namespace.create": "true",
			},
			"fiftyone-teams",
		},
		{
			"overrideNamespaceName",
			map[string]string{
				"namespace.create": "true",
				"namespace.name":   "test-namespace-name",
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

			if testCase.expected == "" {
				output, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/namespace.yaml in chart")

				var namespace corev1.Namespace
				helm.UnmarshalK8SYaml(subT, output, &namespace)

				s.Empty(namespace.ObjectMeta.Name, "Name should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var namespace corev1.Namespace
				helm.UnmarshalK8SYaml(subT, output, &namespace)

				s.Equal(testCase.expected, namespace.ObjectMeta.Name, "Namespace name should be equal.")
			}
		})
	}
}
