//go:build kubeall || helm || unit || unitSecrets
// +build kubeall helm unit unitSecrets

package unit

import (
	// "encoding/json"
	// "fmt"
	"path/filepath"
	// "reflect"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corev1 "k8s.io/api/core/v1"
)

type secretsTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestSecretsTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &secretsTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/secrets.yaml"},
	})
}

func (s *secretsTemplateTest) TestDisabled() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-teams-secrets",
		},
		{
			"defaultValuesSecretsDisabled",
			map[string]string{
				"secret.create": "false",
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
				s.ErrorContains(err, "could not find template templates/secrets.yaml in chart")

				var secret corev1.Secret
				helm.UnmarshalK8SYaml(subT, output, &secret)

				s.Empty(secret.ObjectMeta.Name, "Name should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var secret corev1.Secret
				helm.UnmarshalK8SYaml(subT, output, &secret)

				s.Equal(testCase.expected, secret.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

func (s *secretsTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"fiftyone-teams-secrets",
		},
		{
			"overrideName",
			map[string]string{
				"secret.name": "test-secret-name",
			},
			"test-secret-name",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var secret corev1.Secret
			helm.UnmarshalK8SYaml(subT, output, &secret)

			s.Equal(testCase.expected, secret.ObjectMeta.Name, "Name name should be equal.")
		})
	}
}

func (s *secretsTemplateTest) TestMetadataNamespace() {
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
			"overrideName",
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

			var secret corev1.Secret
			helm.UnmarshalK8SYaml(subT, output, &secret)

			s.Equal(testCase.expected, secret.ObjectMeta.Namespace, "Namespace name should be equal.")
		})
	}
}

func (s *secretsTemplateTest) TestType() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			"Opaque",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var secret corev1.Secret
			helm.UnmarshalK8SYaml(subT, output, &secret)

			s.Equal(testCase.expected, string(secret.Type), "Type name should be Opaque.")
		})
	}
}

func (s *secretsTemplateTest) TestData() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected map[string]string
	}{
		{
			"defaultValues",
			nil,
			map[string]string{
				"cookieSecret":            "",
				"encryptionKey":           "",
				"fiftyoneDatabaseName":    "",
				"mongodbConnectionString": "",
			},
		},
		{
			"overrideSecretFiftyone",
			map[string]string{
				"secret.fiftyone.cookieSecret":            "test-cookie-secret",
				"secret.fiftyone.encryptionKey":           "test-encryption-key",
				"secret.fiftyone.fiftyoneDatabaseName":    "test-fiftyone-database",
				"secret.fiftyone.mongodbConnectionString": "test-mongodb-connection-string",
			},
			map[string]string{
				"cookieSecret":            "test-cookie-secret",
				"encryptionKey":           "test-encryption-key",
				"fiftyoneDatabaseName":    "test-fiftyone-database",
				"mongodbConnectionString": "test-mongodb-connection-string",
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

			var secret corev1.Secret
			helm.UnmarshalK8SYaml(subT, output, &secret)

			for key, value := range testCase.expected {
				foundValue := secret.Data[key]
				s.Equal(value, string(foundValue), "Data should contain all set values.")
			}
		})
	}
}
