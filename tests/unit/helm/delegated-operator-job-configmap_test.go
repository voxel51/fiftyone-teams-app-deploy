//go:build kubeall || helm || unit || unitConfigMap
// +build kubeall helm unit unitConfigMap

package unit

import (
	"fmt"
	"os"
	"path/filepath"

	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/noirbizarre/gonja"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type doK8sConfigMapTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func convertJinjaToYAML(jinjaTemplate string, data map[string]interface{}) (string, error) {
	// Parse the Jinja template
	template, err := gonja.FromString(jinjaTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse Jinja template: %w", err)
	}

	// Render the template with provided data
	rendered, err := template.Execute(data)
	if err != nil {
		return "", fmt.Errorf("failed to render Jinja template: %w", err)
	}

	return rendered, nil
}

func TestDoK8sConfigMapTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &doK8sConfigMapTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/delegated-operator-job-configmap.yaml"},
	})
}

func (s *doK8sConfigMapTemplateTest) TestDisabled() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			fmt.Sprintf("%s-fiftyone-teams-app-do-templates", s.releaseName),
		},
		{
			"defaultValuesConfigMapDisabled",
			map[string]string{
				"delegatedOperatorJobTemplates.configMap.create": "false",
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
				s.ErrorContains(err, "could not find template templates/delegated-operator-job-configmap.yaml in chart")

				var configMap corev1.ConfigMap
				helm.UnmarshalK8SYaml(subT, output, &configMap)

				s.Empty(configMap.ObjectMeta.Name, "Name should be empty")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var configMap corev1.ConfigMap
				helm.UnmarshalK8SYaml(subT, output, &configMap)

				s.Equal(testCase.expected, configMap.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

func (s *doK8sConfigMapTemplateTest) TestMetadataName() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			nil,
			fmt.Sprintf("%s-fiftyone-teams-app-do-templates", s.releaseName),
		},
		{
			"overrideName",
			map[string]string{
				"delegatedOperatorJobTemplates.configMap.name": "test-config-map",
			},
			"test-config-map",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(subT, output, &configMap)

			s.Equal(testCase.expected, configMap.ObjectMeta.Name, "Name name should be equal.")
		})
	}
}

func (s *doK8sConfigMapTemplateTest) TestMetadataNamespace() {
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

			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(subT, output, &configMap)

			s.Equal(testCase.expected, configMap.ObjectMeta.Namespace, "Namespace name should be equal.")
		})
	}
}

func (s *doK8sConfigMapTemplateTest) TestMetadataAnnotations() {
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
				"delegatedOperatorJobTemplates.configMap.annotations.annotation-1": "annotation-1-value",
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

			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(subT, output, &configMap)

			if testCase.expected == nil {
				s.Nil(configMap.ObjectMeta.Annotations, "Annotations should be nil")
			} else {
				for key, value := range testCase.expected {
					foundValue := configMap.ObjectMeta.Annotations[key]
					s.Equal(value, foundValue, "Annotations should contain all set annotations.")
				}
			}
		})
	}
}

func (s *doK8sConfigMapTemplateTest) TestMetadataLabels() {
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
				"app.kubernetes.io/name":       "fiftyone-test-fiftyone-teams-app-do-templates",
				"app.kubernetes.io/instance":   "fiftyone-test",
				"app.voxel51.com/component":    "do-templates",
			},
		},
		{
			"overrideMetadataLabels",
			map[string]string{
				"delegatedOperatorJobTemplates.configMap.name":         "test-config-map",
				"delegatedOperatorJobTemplates.configMap.labels.color": "blue",
			},
			map[string]string{
				"helm.sh/chart":                fmt.Sprintf("fiftyone-teams-app-%s", chartVersion),
				"app.kubernetes.io/version":    fmt.Sprintf("%s", chartAppVersion),
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/name":       "test-config-map",
				"app.kubernetes.io/instance":   "fiftyone-test",
				"app.voxel51.com/component":    "do-templates",
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

			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(subT, output, &configMap)

			for key, value := range testCase.expected {
				foundValue := configMap.ObjectMeta.Labels[key]
				s.Equal(value, foundValue, "Labels should contain all set labels.")
			}
		})
	}
}

func (s *doK8sConfigMapTemplateTest) TestData() {
	// Get chart info (to later obtain the chart's appVersion)
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)

	// Get version from chart info
	chartVersion, exists := cInfo["version"]
	s.True(exists, "failed to get version from chart info")

	testCases := []struct {
		name     string
		values   map[string]string
		expected func(data map[string]string)
	}{
		{
			"defaultValues",
			nil,
			func(data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"id":      strings.ToLower(random.UniqueId()),
					"command": "fiftyone",
					"args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default.yaml": "cpu-default.yaml",
				}

				s.Equal(len(tests), len(data), "Number of test entries should be equal to number of data entries.")

				for expectedFile, actualYamlKey := range tests {
					expectedJobJinja := s.loadTestFile(expectedFile, chartVersion.(string))
					expectedJobYaml, err := convertJinjaToYAML(expectedJobJinja, jinjaArgs)
					s.NoError(err)

					actualJobYaml, err := convertJinjaToYAML(data[actualYamlKey], jinjaArgs)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(expectedJobYaml), &expectedJobConfig)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(actualJobYaml), &actualJobConfig)
					s.NoError(err)

					s.Equal(expectedJobConfig, actualJobConfig, "Jobs should be equal")
				}
			},
		},
		{
			"overrideTemplateValues",
			map[string]string{
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "disktype",
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "ssd",
				"delegatedOperatorJobTemplates.template.backoffLimit":                              "1",
				"delegatedOperatorJobTemplates.template.ttlSecondsAfterFinished":                   "2",
				"delegatedOperatorJobTemplates.template.activeDeadlineSeconds":                     "3",
				"delegatedOperatorJobTemplates.template.completions":                               "4",
				"delegatedOperatorJobTemplates.template.containerSecurityContext.runAsGroup":       "2000",
				"delegatedOperatorJobTemplates.template.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH": "/tmp/foo",
				"delegatedOperatorJobTemplates.template.env.ADDITIONAL_ENV_VAR":                    "an-env-var",
				"delegatedOperatorJobTemplates.template.image.pullPolicy":                          "Never",
				"delegatedOperatorJobTemplates.template.image.repository":                          "us.gcr.io",
				"delegatedOperatorJobTemplates.template.tag":                                       "0.0.0",
				"delegatedOperatorJobTemplates.template.jobAnnotations.annotation-1":               "annotation-1-value",
				"delegatedOperatorJobTemplates.template.nodeSelector.node-selector-1":              "node-selector-1-value",
				"delegatedOperatorJobTemplates.template.labels.labels-1":                           "label-1-value",
				"delegatedOperatorJobTemplates.template.podAnnotations.pod-annotation-1":           "pod-annotation-1-value",
				"delegatedOperatorJobTemplates.template.podSecurityContext.runAsUser":              "1000",
				"delegatedOperatorJobTemplates.template.resources.cpu":                             "2",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretName":           "secret-name",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretKey":            "secret-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].key":                        "example-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].operator":                   "Exists",
				"delegatedOperatorJobTemplates.template.tolerations[0].effect":                     "NoSchedule",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].mountPath":                 "/test-data-volume",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].name":                      "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].name":                           "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].hostPath.path":                  "/test-volume",
			},
			func(data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"id":      strings.ToLower(random.UniqueId()),
					"command": "fiftyone",
					"args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default-override-template-values.yaml": "cpu-default.yaml",
				}

				s.Equal(len(tests), len(data), "Number of test entries should be equal to number of data entries.")

				for expectedFile, actualYamlKey := range tests {
					expectedJobJinja := s.loadTestFile(expectedFile, chartVersion.(string))
					expectedJobYaml, err := convertJinjaToYAML(expectedJobJinja, jinjaArgs)
					s.NoError(err)

					actualJobYaml, err := convertJinjaToYAML(data[actualYamlKey], jinjaArgs)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(expectedJobYaml), &expectedJobConfig)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(actualJobYaml), &actualJobConfig)
					s.NoError(err)

					s.Equal(expectedJobConfig, actualJobConfig, "Jobs should be equal")
				}
			},
		},
		{
			"overrideTemplateAndInstanceValues",
			map[string]string{
				// Template Values
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "disktype",
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "ssd",
				"delegatedOperatorJobTemplates.template.backoffLimit":                              "1",
				"delegatedOperatorJobTemplates.template.ttlSecondsAfterFinished":                   "2",
				"delegatedOperatorJobTemplates.template.activeDeadlineSeconds":                     "3",
				"delegatedOperatorJobTemplates.template.completions":                               "4",
				"delegatedOperatorJobTemplates.template.containerSecurityContext.runAsGroup":       "2000",
				"delegatedOperatorJobTemplates.template.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH": "/tmp/foo",
				"delegatedOperatorJobTemplates.template.env.ADDITIONAL_ENV_VAR":                    "an-env-var",
				"delegatedOperatorJobTemplates.template.image.pullPolicy":                          "Never",
				"delegatedOperatorJobTemplates.template.image.repository":                          "us.gcr.io",
				"delegatedOperatorJobTemplates.template.tag":                                       "0.0.0",
				"delegatedOperatorJobTemplates.template.jobAnnotations.annotation-1":               "annotation-1-value",
				"delegatedOperatorJobTemplates.template.nodeSelector.node-selector-1":              "node-selector-1-value",
				"delegatedOperatorJobTemplates.template.labels.labels-1":                           "label-1-value",
				"delegatedOperatorJobTemplates.template.podAnnotations.pod-annotation-1":           "pod-annotation-1-value",
				"delegatedOperatorJobTemplates.template.podSecurityContext.runAsUser":              "1000",
				"delegatedOperatorJobTemplates.template.resources.cpu":                             "2",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretName":           "secret-name",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretKey":            "secret-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].key":                        "example-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].operator":                   "Exists",
				"delegatedOperatorJobTemplates.template.tolerations[0].effect":                     "NoSchedule",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].mountPath":                 "/test-data-volume",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].name":                      "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].name":                           "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].hostPath.path":                  "/test-volume",

				// Override All Values From Template
				"delegatedOperatorJobTemplates.jobs.override-example.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "hostname",
				"delegatedOperatorJobTemplates.jobs.override-example.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "NotIn",
				"delegatedOperatorJobTemplates.jobs.override-example.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "override",
				"delegatedOperatorJobTemplates.jobs.override-example.backoffLimit":                              "10",
				"delegatedOperatorJobTemplates.jobs.override-example.ttlSecondsAfterFinished":                   "20",
				"delegatedOperatorJobTemplates.jobs.override-example.activeDeadlineSeconds":                     "30",
				"delegatedOperatorJobTemplates.jobs.override-example.completions":                               "40",
				"delegatedOperatorJobTemplates.jobs.override-example.containerSecurityContext.runAsGroup":       "4000",
				"delegatedOperatorJobTemplates.jobs.override-example.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH": "/tmp/foo-override",
				"delegatedOperatorJobTemplates.jobs.override-example.env.ADDITIONAL_ENV_VAR":                    "an-env-var-override",
				"delegatedOperatorJobTemplates.jobs.override-example.image.pullPolicy":                          "Always",
				"delegatedOperatorJobTemplates.jobs.override-example.image.repository":                          "eu.gcr.io",
				"delegatedOperatorJobTemplates.jobs.override-example.tag":                                       "1.1.1",
				"delegatedOperatorJobTemplates.jobs.override-example.jobAnnotations.annotation-1":               "annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.nodeSelector.node-selector-1":              "node-selector-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.labels.labels-1":                           "label-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podAnnotations.pod-annotation-1":           "pod-annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podSecurityContext.runAsUser":              "3000",
				"delegatedOperatorJobTemplates.jobs.override-example.resources.cpu":                             "20",
				"delegatedOperatorJobTemplates.jobs.override-example.secretEnv.SECRET_ENV.secretName":           "secret-name-override",
				"delegatedOperatorJobTemplates.jobs.override-example.secretEnv.SECRET_ENV.secretKey":            "secret-key-override",
				"delegatedOperatorJobTemplates.jobs.override-example.tolerations[0].key":                        "example-key-override",
				"delegatedOperatorJobTemplates.jobs.override-example.tolerations[0].operator":                   "NotExists",
				"delegatedOperatorJobTemplates.jobs.override-example.tolerations[0].effect":                     "Schedule",
				"delegatedOperatorJobTemplates.jobs.override-example.volumeMounts[0].mountPath":                 "/test-data-volume-override",
				"delegatedOperatorJobTemplates.jobs.override-example.volumeMounts[0].name":                      "test-volume-override",
				"delegatedOperatorJobTemplates.jobs.override-example.volumes[0].name":                           "test-volume-override",
				"delegatedOperatorJobTemplates.jobs.override-example.volumes[0].hostPath.path":                  "/test-volume-override",
			},
			func(data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"id":      strings.ToLower(random.UniqueId()),
					"command": "fiftyone",
					"args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default-override-template-values.yaml": "cpu-default.yaml",
					"test_data/delegated-operator-job-configmap_test/expected-override-example-template.yaml":            "override-example.yaml",
				}

				s.Equal(len(tests), len(data), "Number of test entries should be equal to number of data entries.")

				for expectedFile, actualYamlKey := range tests {
					expectedJobJinja := s.loadTestFile(expectedFile, chartVersion.(string))
					expectedJobYaml, err := convertJinjaToYAML(expectedJobJinja, jinjaArgs)
					s.NoError(err)

					actualJobYaml, err := convertJinjaToYAML(data[actualYamlKey], jinjaArgs)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(expectedJobYaml), &expectedJobConfig)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(actualJobYaml), &actualJobConfig)
					s.NoError(err)

					s.Equal(expectedJobConfig, actualJobConfig, "Jobs should be equal")
				}
			},
		},
		{
			"overrideTemplateAndInstanceValuesCascading",
			map[string]string{
				// Template Values
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "disktype",
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
				"delegatedOperatorJobTemplates.template.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "ssd",
				"delegatedOperatorJobTemplates.template.backoffLimit":                              "1",
				"delegatedOperatorJobTemplates.template.ttlSecondsAfterFinished":                   "2",
				"delegatedOperatorJobTemplates.template.activeDeadlineSeconds":                     "3",
				"delegatedOperatorJobTemplates.template.completions":                               "4",
				"delegatedOperatorJobTemplates.template.containerSecurityContext.runAsGroup":       "2000",
				"delegatedOperatorJobTemplates.template.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH": "/tmp/foo",
				"delegatedOperatorJobTemplates.template.env.ADDITIONAL_ENV_VAR":                    "an-env-var",
				"delegatedOperatorJobTemplates.template.image.pullPolicy":                          "Never",
				"delegatedOperatorJobTemplates.template.image.repository":                          "us.gcr.io",
				"delegatedOperatorJobTemplates.template.tag":                                       "0.0.0",
				"delegatedOperatorJobTemplates.template.jobAnnotations.annotation-1":               "annotation-1-value",
				"delegatedOperatorJobTemplates.template.nodeSelector.node-selector-1":              "node-selector-1-value",
				"delegatedOperatorJobTemplates.template.labels.labels-1":                           "label-1-value",
				"delegatedOperatorJobTemplates.template.podAnnotations.pod-annotation-1":           "pod-annotation-1-value",
				"delegatedOperatorJobTemplates.template.podSecurityContext.runAsUser":              "1000",
				"delegatedOperatorJobTemplates.template.resources.cpu":                             "2",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretName":           "secret-name",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretKey":            "secret-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].key":                        "example-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].operator":                   "Exists",
				"delegatedOperatorJobTemplates.template.tolerations[0].effect":                     "NoSchedule",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].mountPath":                 "/test-data-volume",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].name":                      "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].name":                           "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].hostPath.path":                  "/test-volume",

				// Override All Values From Template
				"delegatedOperatorJobTemplates.jobs.override-example.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "hostname",
				"delegatedOperatorJobTemplates.jobs.override-example.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "NotIn",
				"delegatedOperatorJobTemplates.jobs.override-example.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "override",
				"delegatedOperatorJobTemplates.jobs.override-example.backoffLimit":                              "10",
				"delegatedOperatorJobTemplates.jobs.override-example.ttlSecondsAfterFinished":                   "20",
				"delegatedOperatorJobTemplates.jobs.override-example.activeDeadlineSeconds":                     "30",
				"delegatedOperatorJobTemplates.jobs.override-example.completions":                               "40",
				"delegatedOperatorJobTemplates.jobs.override-example.containerSecurityContext.runAsGroup":       "4000",
				"delegatedOperatorJobTemplates.jobs.override-example.env.FIFTYONE_DELEGATED_OPERATION_LOG_PATH": "/tmp/foo-override",
				"delegatedOperatorJobTemplates.jobs.override-example.env.ADDITIONAL_ENV_VAR":                    "an-env-var-override",
				"delegatedOperatorJobTemplates.jobs.override-example.image.pullPolicy":                          "Always",
				"delegatedOperatorJobTemplates.jobs.override-example.image.repository":                          "eu.gcr.io",
				"delegatedOperatorJobTemplates.jobs.override-example.tag":                                       "1.1.1",
				"delegatedOperatorJobTemplates.jobs.override-example.jobAnnotations.annotation-1":               "annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.nodeSelector.node-selector-1":              "node-selector-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.labels.labels-1":                           "label-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podAnnotations.pod-annotation-1":           "pod-annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podSecurityContext.runAsUser":              "3000",
				"delegatedOperatorJobTemplates.jobs.override-example.resources.cpu":                             "20",
				"delegatedOperatorJobTemplates.jobs.override-example.secretEnv.SECRET_ENV.secretName":           "secret-name-override",
				"delegatedOperatorJobTemplates.jobs.override-example.secretEnv.SECRET_ENV.secretKey":            "secret-key-override",
				"delegatedOperatorJobTemplates.jobs.override-example.tolerations[0].key":                        "example-key-override",
				"delegatedOperatorJobTemplates.jobs.override-example.tolerations[0].operator":                   "NotExists",
				"delegatedOperatorJobTemplates.jobs.override-example.tolerations[0].effect":                     "Schedule",
				"delegatedOperatorJobTemplates.jobs.override-example.volumeMounts[0].mountPath":                 "/test-data-volume-override",
				"delegatedOperatorJobTemplates.jobs.override-example.volumeMounts[0].name":                      "test-volume-override",
				"delegatedOperatorJobTemplates.jobs.override-example.volumes[0].name":                           "test-volume-override",
				"delegatedOperatorJobTemplates.jobs.override-example.volumes[0].hostPath.path":                  "/test-volume-override",

				// Override Some Values From Template
				"delegatedOperatorJobTemplates.jobs.override-example-cascading.backoffLimit":            "10",
				"delegatedOperatorJobTemplates.jobs.override-example-cascading.ttlSecondsAfterFinished": "20",
				"delegatedOperatorJobTemplates.jobs.override-example-cascading.activeDeadlineSeconds":   "30",
				"delegatedOperatorJobTemplates.jobs.override-example-cascading.completions":             "40",
			},
			func(data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"id":      strings.ToLower(random.UniqueId()),
					"command": "fiftyone",
					"args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default-override-template-values.yaml": "cpu-default.yaml",
					"test_data/delegated-operator-job-configmap_test/expected-override-example-template.yaml":            "override-example.yaml",
					"test_data/delegated-operator-job-configmap_test/expected-override-example-cascading-template.yaml":  "override-example-cascading.yaml",
				}

				s.Equal(len(tests), len(data), "Number of test entries should be equal to number of data entries.")

				for expectedFile, actualYamlKey := range tests {
					expectedJobJinja := s.loadTestFile(expectedFile, chartVersion.(string))
					expectedJobYaml, err := convertJinjaToYAML(expectedJobJinja, jinjaArgs)
					s.NoError(err)

					actualJobYaml, err := convertJinjaToYAML(data[actualYamlKey], jinjaArgs)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(expectedJobYaml), &expectedJobConfig)
					s.NoError(err)

					err = yaml.Unmarshal([]byte(actualJobYaml), &actualJobConfig)
					s.NoError(err)

					s.Equal(expectedJobConfig, actualJobConfig, "Jobs should be equal")
				}
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

			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(subT, output, &configMap)

			testCase.expected(configMap.Data)
		})
	}
}

func (s *doK8sConfigMapTemplateTest) loadTestFile(filename, chartVersion string) string {
	content, err := os.ReadFile(filename)
	s.NoError(err, "Failed to read test file: %s", filename)
	result := strings.ReplaceAll(string(content), "{{CHART_VERSION}}", chartVersion)
	result = strings.TrimSpace(result)
	return strings.TrimSpace(result)
}
