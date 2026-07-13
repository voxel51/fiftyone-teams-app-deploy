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

	gruntworkTesting "github.com/gruntwork-io/terratest/modules/testing"

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

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

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

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

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

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

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

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}
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
				"app.voxel51.com/component":    "on-demand-delegated-operators",
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

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}
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
		expected func(subT gruntworkTesting.TestingT, data map[string]string)
	}{
		{
			"defaultValues",
			nil,
			func(subT gruntworkTesting.TestingT, data map[string]string) {
				s.Empty(data, "Data should be empty")
			},
		},
		{
			"defaultValuesCpuEnabled",
			map[string]string{
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			func(subT gruntworkTesting.TestingT, data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"_id":      strings.ToLower(random.UniqueId()),
					"_command": "fiftyone",
					"_args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default.yaml": "cpuDefault.yaml",
				}

				s.Equal(len(tests), len(data), "Number of test entries should be equal to number of data entries.")

				for expectedFile, actualYamlKey := range tests {
					expectedJobJinja := s.loadTestFile(expectedFile, chartVersion.(string))
					expectedJobYaml, err := convertJinjaToYAML(expectedJobJinja, jinjaArgs)
					s.NoError(err)

					actualJobYaml, err := convertJinjaToYAML(data[actualYamlKey], jinjaArgs)
					s.NoError(err)

					helm.UnmarshalK8SYaml(subT, expectedJobYaml, &expectedJobConfig)

					helm.UnmarshalK8SYaml(subT, actualJobYaml, &actualJobConfig)

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
				"delegatedOperatorJobTemplates.template.image.tag":                                 "0.0.0",
				"delegatedOperatorJobTemplates.template.jobAnnotations.annotation-1":               "annotation-1-value",
				"delegatedOperatorJobTemplates.template.nodeSelector.node-selector-1":              "node-selector-1-value",
				"delegatedOperatorJobTemplates.template.labels.labels-1":                           "label-1-value",
				"delegatedOperatorJobTemplates.template.podAnnotations.pod-annotation-1":           "pod-annotation-1-value",
				"delegatedOperatorJobTemplates.template.podSecurityContext.runAsUser":              "1000",
				"delegatedOperatorJobTemplates.template.resources.requests.cpu":                    "2",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretName":           "secret-name",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretKey":            "secret-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].key":                        "example-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].operator":                   "Exists",
				"delegatedOperatorJobTemplates.template.tolerations[0].effect":                     "NoSchedule",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].mountPath":                 "/test-data-volume",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].name":                      "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].name":                           "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].hostPath.path":                  "/test-volume",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":                             "nil",
			},
			func(subT gruntworkTesting.TestingT, data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"_id":      strings.ToLower(random.UniqueId()),
					"_command": "fiftyone",
					"_args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default-override-template-values.yaml": "cpuDefault.yaml",
				}

				s.Equal(len(tests), len(data), "Number of test entries should be equal to number of data entries.")

				for expectedFile, actualYamlKey := range tests {
					expectedJobJinja := s.loadTestFile(expectedFile, chartVersion.(string))
					expectedJobYaml, err := convertJinjaToYAML(expectedJobJinja, jinjaArgs)
					s.NoError(err)

					actualJobYaml, err := convertJinjaToYAML(data[actualYamlKey], jinjaArgs)
					s.NoError(err)

					helm.UnmarshalK8SYaml(subT, expectedJobYaml, &expectedJobConfig)

					helm.UnmarshalK8SYaml(subT, actualJobYaml, &actualJobConfig)

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
				"delegatedOperatorJobTemplates.template.image.tag":                                 "0.0.0",
				"delegatedOperatorJobTemplates.template.jobAnnotations.annotation-1":               "annotation-1-value",
				"delegatedOperatorJobTemplates.template.nodeSelector.node-selector-1":              "node-selector-1-value",
				"delegatedOperatorJobTemplates.template.labels.labels-1":                           "label-1-value",
				"delegatedOperatorJobTemplates.template.podAnnotations.pod-annotation-1":           "pod-annotation-1-value",
				"delegatedOperatorJobTemplates.template.podSecurityContext.runAsUser":              "1000",
				"delegatedOperatorJobTemplates.template.resources.requests.cpu":                    "2",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretName":           "secret-name",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretKey":            "secret-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].key":                        "example-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].operator":                   "Exists",
				"delegatedOperatorJobTemplates.template.tolerations[0].effect":                     "NoSchedule",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].mountPath":                 "/test-data-volume",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].name":                      "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].name":                           "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].hostPath.path":                  "/test-volume",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":                             "nil",

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
				"delegatedOperatorJobTemplates.jobs.override-example.image.tag":                                 "1.1.1",
				"delegatedOperatorJobTemplates.jobs.override-example.jobAnnotations.annotation-1":               "annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.nodeSelector.node-selector-1":              "node-selector-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.labels.labels-1":                           "label-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podAnnotations.pod-annotation-1":           "pod-annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podSecurityContext.runAsUser":              "3000",
				"delegatedOperatorJobTemplates.jobs.override-example.resources.requests.cpu":                    "20",
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
			func(subT gruntworkTesting.TestingT, data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"_id":      strings.ToLower(random.UniqueId()),
					"_command": "fiftyone",
					"_args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default-override-template-values.yaml": "cpuDefault.yaml",
					"test_data/delegated-operator-job-configmap_test/expected-override-example-template.yaml":            "override-example.yaml",
				}

				s.Equal(len(tests), len(data), "Number of test entries should be equal to number of data entries.")

				for expectedFile, actualYamlKey := range tests {
					expectedJobJinja := s.loadTestFile(expectedFile, chartVersion.(string))
					expectedJobYaml, err := convertJinjaToYAML(expectedJobJinja, jinjaArgs)
					s.NoError(err)

					actualJobYaml, err := convertJinjaToYAML(data[actualYamlKey], jinjaArgs)
					s.NoError(err)

					helm.UnmarshalK8SYaml(subT, expectedJobYaml, &expectedJobConfig)

					helm.UnmarshalK8SYaml(subT, actualJobYaml, &actualJobConfig)

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
				"delegatedOperatorJobTemplates.template.image.tag":                                 "0.0.0",
				"delegatedOperatorJobTemplates.template.jobAnnotations.annotation-1":               "annotation-1-value",
				"delegatedOperatorJobTemplates.template.nodeSelector.node-selector-1":              "node-selector-1-value",
				"delegatedOperatorJobTemplates.template.labels.labels-1":                           "label-1-value",
				"delegatedOperatorJobTemplates.template.podAnnotations.pod-annotation-1":           "pod-annotation-1-value",
				"delegatedOperatorJobTemplates.template.podSecurityContext.runAsUser":              "1000",
				"delegatedOperatorJobTemplates.template.resources.requests.cpu":                    "2",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretName":           "secret-name",
				"delegatedOperatorJobTemplates.template.secretEnv.SECRET_ENV.secretKey":            "secret-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].key":                        "example-key",
				"delegatedOperatorJobTemplates.template.tolerations[0].operator":                   "Exists",
				"delegatedOperatorJobTemplates.template.tolerations[0].effect":                     "NoSchedule",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].mountPath":                 "/test-data-volume",
				"delegatedOperatorJobTemplates.template.volumeMounts[0].name":                      "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].name":                           "test-volume",
				"delegatedOperatorJobTemplates.template.volumes[0].hostPath.path":                  "/test-volume",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":                             "nil",

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
				"delegatedOperatorJobTemplates.jobs.override-example.image.tag":                                 "1.1.1",
				"delegatedOperatorJobTemplates.jobs.override-example.jobAnnotations.annotation-1":               "annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.nodeSelector.node-selector-1":              "node-selector-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.labels.labels-1":                           "label-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podAnnotations.pod-annotation-1":           "pod-annotation-1-value-override",
				"delegatedOperatorJobTemplates.jobs.override-example.podSecurityContext.runAsUser":              "3000",
				"delegatedOperatorJobTemplates.jobs.override-example.resources.requests.cpu":                    "20",
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
			func(subT gruntworkTesting.TestingT, data map[string]string) {
				var expectedJobConfig batchv1.Job
				var actualJobConfig batchv1.Job

				jinjaArgs := map[string]interface{}{
					"id":      strings.ToLower(random.UniqueId()),
					"command": "fiftyone",
					"args":    []string{"test", "arg1"},
				}

				tests := map[string]string{
					"test_data/delegated-operator-job-configmap_test/expected-cpu-default-override-template-values.yaml": "cpuDefault.yaml",
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

					helm.UnmarshalK8SYaml(subT, expectedJobYaml, &expectedJobConfig)

					helm.UnmarshalK8SYaml(subT, actualJobYaml, &actualJobConfig)

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

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

			output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

			var configMap corev1.ConfigMap
			helm.UnmarshalK8SYaml(subT, output, &configMap)

			testCase.expected(subT, configMap.Data)
		})
	}
}

// renderJob renders a single job from the rendered ConfigMap data and
// returns it as a batchv1.Job for introspection. The job template body
// in the ConfigMap is Jinja, which we render with placeholder values so
// the result is valid YAML.
func (s *doK8sConfigMapTemplateTest) renderJob(data map[string]string, key string) batchv1.Job {
	s.T().Helper()
	s.Require().Contains(data, key, "ConfigMap should contain key %q", key)

	jinjaArgs := map[string]interface{}{
		"_id":      "test-id",
		"_command": "fiftyone",
		"_args":    []string{"test"},
	}
	jobYAML, err := convertJinjaToYAML(data[key], jinjaArgs)
	s.Require().NoError(err)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), jobYAML, &job)
	return job
}

func (s *doK8sConfigMapTemplateTest) TestTelemetrySocketInjection() {
	const socketName = "telemetry-socket"
	jobKey := "cpuDefault.yaml"

	// helper: count entries with the given name
	countByName := func(volumes []corev1.Volume, name string) int {
		n := 0
		for _, v := range volumes {
			if v.Name == name {
				n++
			}
		}
		return n
	}
	countMountsByName := func(mounts []corev1.VolumeMount, name string) int {
		n := 0
		for _, m := range mounts {
			if m.Name == name {
				n++
			}
		}
		return n
	}

	testCases := []struct {
		name      string
		values    map[string]string
		expectVol int // expected occurrences of telemetry-socket volume
		expectMnt int // expected occurrences of telemetry-socket mount
	}{
		{
			name: "autoInjectsWhenUserHasNoTelemetrySocket",
			values: map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			expectVol: 1,
			expectMnt: 1,
		},
		{
			name: "doesNotDuplicateWhenUserDeclaresTelemetrySocket",
			values: map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":                     "nil",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.volumes[0].name":            socketName,
				"delegatedOperatorJobTemplates.jobs.cpuDefault.volumes[0].emptyDir.medium": "Memory",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.volumeMounts[0].name":       socketName,
				"delegatedOperatorJobTemplates.jobs.cpuDefault.volumeMounts[0].mountPath":  "/custom/path",
			},
			expectVol: 1,
			expectMnt: 1,
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

			job := s.renderJob(configMap.Data, jobKey)

			s.Equal(
				testCase.expectVol,
				countByName(job.Spec.Template.Spec.Volumes, socketName),
				"telemetry-socket volume count mismatch",
			)
			s.Require().NotEmpty(job.Spec.Template.Spec.Containers, "expected at least one container")
			s.Equal(
				testCase.expectMnt,
				countMountsByName(job.Spec.Template.Spec.Containers[0].VolumeMounts, socketName),
				"telemetry-socket volumeMount count mismatch",
			)
		})
	}
}

// TestTelemetryDisabledOmitsSidecar verifies that when telemetry is disabled
// the rendered job carries neither the telemetry-sidecar initContainer nor the
// telemetry-socket volume/mount. The full disabled-telemetry job shape is also
// covered by golden files in TestData; this is the focused negative assertion
// living next to the positive cases above.
func (s *doK8sConfigMapTemplateTest) TestTelemetryDisabledOmitsSidecar() {
	const socketName = "telemetry-socket"
	jobKey := "cpuDefault.yaml"

	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "false",
		"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
	}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	job := s.renderJob(configMap.Data, jobKey)

	s.Nil(findContainer(job.Spec.Template.Spec.InitContainers, "telemetry-sidecar"),
		"telemetry-sidecar initContainer should be absent when telemetry is disabled")
	s.Nil(findVolume(job.Spec.Template.Spec.Volumes, socketName),
		"telemetry-socket volume should be absent when telemetry is disabled")
	s.Require().NotEmpty(job.Spec.Template.Spec.Containers, "expected at least one container")
	s.Nil(findVolumeMount(job.Spec.Template.Spec.Containers[0].VolumeMounts, socketName),
		"telemetry-socket volumeMount should be absent when telemetry is disabled")
}

// TestTelemetrySidecarGpuEnv verifies that when a delegated-operator job's
// executor requests a GPU (via resources.limits or resources.requests), its
// telemetry native-sidecar is given the NVIDIA_* env vars needed to read GPU
// metrics, without the sidecar requesting its own nvidia.com/gpu allocation.
// When no GPU is requested, the sidecar receives no NVIDIA_* env vars.
func (s *doK8sConfigMapTemplateTest) TestTelemetrySidecarGpuEnv() {
	const gpuResource = "nvidia.com/gpu"
	jobKey := "cpuDefault.yaml"

	// gpuKey escapes the dot in nvidia.com/gpu so helm --set treats the whole
	// string as a single map key rather than a nested path.
	gpuKey := func(section string) string {
		return "delegatedOperatorJobTemplates.jobs.cpuDefault.resources." +
			section + ".nvidia\\.com/gpu"
	}

	testCases := []struct {
		name      string
		values    map[string]string
		expectGpu bool
	}{
		{
			name: "gpuInLimitsExposesEnvToSidecar",
			values: map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
				gpuKey("limits"): "1",
			},
			expectGpu: true,
		},
		{
			name: "gpuInRequestsExposesEnvToSidecar",
			values: map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
				gpuKey("requests"): "1",
			},
			expectGpu: true,
		},
		{
			name: "noGpuOmitsEnvFromSidecar",
			values: map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			expectGpu: false,
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

			job := s.renderJob(configMap.Data, jobKey)

			// The native-sidecar runs as an initContainer (restartPolicy: Always).
			sidecar := findContainer(job.Spec.Template.Spec.InitContainers, "telemetry-sidecar")
			s.Require().NotNil(sidecar, "telemetry-sidecar initContainer not found")

			visibleDevices, hasVisibleDevices := envValue(sidecar.Env, "NVIDIA_VISIBLE_DEVICES")
			driverCaps, hasDriverCaps := envValue(sidecar.Env, "NVIDIA_DRIVER_CAPABILITIES")

			if testCase.expectGpu {
				s.True(hasVisibleDevices, "sidecar should have NVIDIA_VISIBLE_DEVICES")
				s.Equal("all", visibleDevices, "NVIDIA_VISIBLE_DEVICES value mismatch")
				s.True(hasDriverCaps, "sidecar should have NVIDIA_DRIVER_CAPABILITIES")
				s.Equal("compute,utility", driverCaps, "NVIDIA_DRIVER_CAPABILITIES value mismatch")

				// The sidecar reads the executor's GPU; it must not request its own.
				_, limitsHasGpu := sidecar.Resources.Limits[corev1.ResourceName(gpuResource)]
				_, requestsHasGpu := sidecar.Resources.Requests[corev1.ResourceName(gpuResource)]
				s.False(limitsHasGpu, "sidecar must not request nvidia.com/gpu in limits")
				s.False(requestsHasGpu, "sidecar must not request nvidia.com/gpu in requests")
			} else {
				s.False(hasVisibleDevices, "sidecar should not have NVIDIA_VISIBLE_DEVICES when no GPU requested")
				s.False(hasDriverCaps, "sidecar should not have NVIDIA_DRIVER_CAPABILITIES when no GPU requested")
			}
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

// brokerServiceVars returns the per-task variables the kubernetes-service
// broker renders into a service pod template. The sizing vars (_cpu,
// _memory, ...) are part of the broker's contract but chart templates
// deliberately ignore them — sizing comes from the values merge.
func brokerServiceVars() map[string]interface{} {
	return map[string]interface{}{
		"_id":                "task1",
		"_name":              "svc-task1",
		"_namespace":         "svc-ns",
		"_image":             "registry:5000/svc:tag",
		"_command":           "python",
		"_args":              []string{"-m", "some.module"},
		"_env":               []map[string]interface{}{{"name": "SVC_VAR", "value": "svc-val"}},
		"_port":              8000,
		"_health_path":       "/healthz",
		"_health_port":       8000,
		"_cpu":               "",
		"_memory":            "",
		"_ephemeral_storage": "",
		"_gpu_count":         0,
		"_gpu_type":          "",
	}
}

// renderServicePod renders a single service pod template from the rendered
// ConfigMap data with the given broker variables and returns it as a
// corev1.Pod for introspection.
func (s *doK8sConfigMapTemplateTest) renderServicePod(data map[string]string, key string, vars map[string]interface{}) corev1.Pod {
	s.T().Helper()
	s.Require().Contains(data, key, "ConfigMap should contain key %q", key)

	podYAML, err := convertJinjaToYAML(data[key], vars)
	s.Require().NoError(err)

	var pod corev1.Pod
	helm.UnmarshalK8SYaml(s.T(), podYAML, &pod)
	return pod
}

// TestServiceData verifies that entries under
// delegatedOperatorJobTemplates.services render jinja2 Pod manifests into
// the same ConfigMap as the job entries, shaped by the broker's per-task
// variables.
func (s *doK8sConfigMapTemplateTest) TestServiceData() {
	options := &helm.Options{SetValues: disableTelemetry(map[string]string{
		"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":      "nil",
		"delegatedOperatorJobTemplates.services.cpuServices.unused": "nil",
	})}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	// Jobs and services share the ConfigMap, one file per entry
	s.Contains(configMap.Data, "cpuDefault.yaml")
	s.Contains(configMap.Data, "cpuServices.yaml")

	pod := s.renderServicePod(configMap.Data, "cpuServices.yaml", brokerServiceVars())

	// Broker-owned fields
	s.Equal("svc-task1", pod.ObjectMeta.Name)
	s.Equal("svc-ns", pod.ObjectMeta.Namespace)
	s.Require().Len(pod.Spec.Containers, 1)
	container := pod.Spec.Containers[0]
	s.Equal("service", container.Name)
	s.Equal("registry:5000/svc:tag", container.Image)
	s.Equal([]string{"python"}, container.Command)
	s.Equal([]string{"-m", "some.module"}, container.Args)
	s.Require().Len(container.Ports, 1)
	s.Equal(int32(8000), container.Ports[0].ContainerPort)

	// Health-checked service gets a readiness probe on the health port
	s.Require().NotNil(container.ReadinessProbe)
	s.Equal("/healthz", container.ReadinessProbe.HTTPGet.Path)

	// Chart-owned fields, merged like jobs
	s.Equal("fiftyone-teams", pod.Spec.ServiceAccountName)
	s.Equal(
		"cpuServices",
		pod.ObjectMeta.Labels["app.voxel51.com/delegate-operator-template-name"],
	)
	s.Require().NotNil(pod.Spec.SecurityContext)
	s.True(*pod.Spec.SecurityContext.RunAsNonRoot)

	// Both the shared env wiring and the broker's per-task _env land
	apiURL, ok := envValue(container.Env, "API_URL")
	s.True(ok)
	s.Equal("http://teams-api:80", apiURL)
	internal, ok := envValue(container.Env, "FIFTYONE_INTERNAL_SERVICE")
	s.True(ok)
	s.Equal("true", internal)
	svcVar, ok := envValue(container.Env, "SVC_VAR")
	s.True(ok)
	s.Equal("svc-val", svcVar)
}

func (s *doK8sConfigMapTemplateTest) TestServiceWithoutHealthcheck() {
	options := &helm.Options{SetValues: disableTelemetry(map[string]string{
		"delegatedOperatorJobTemplates.services.cpuServices.unused": "nil",
	})}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	vars := brokerServiceVars()
	vars["_health_path"] = ""
	pod := s.renderServicePod(configMap.Data, "cpuServices.yaml", vars)

	s.Nil(pod.Spec.Containers[0].ReadinessProbe, "No probe without a health path")
}

func (s *doK8sConfigMapTemplateTest) TestServiceDisabled() {
	options := &helm.Options{SetValues: disableTelemetry(map[string]string{
		"delegatedOperatorJobTemplates.services.cpuServices.enabled": "false",
	})}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	s.NotContains(configMap.Data, "cpuServices.yaml")
}

// TestServiceResources verifies that service pod sizing comes from the
// chart's values merge (template merged with the entry's override), and
// that the broker's per-task sizing vars — an orchestrator record with
// resource_requests — are ignored: the deployment owns the pod shape,
// same as jobs.
func (s *doK8sConfigMapTemplateTest) TestServiceResources() {
	options := &helm.Options{SetValues: disableTelemetry(map[string]string{
		"delegatedOperatorJobTemplates.template.resources.requests.cpu":                "1",
		"delegatedOperatorJobTemplates.services.cpuServices.unused":                    "nil",
		"delegatedOperatorJobTemplates.services.gpuServices.resources.requests.cpu":    "2",
		"delegatedOperatorJobTemplates.services.gpuServices.resources.requests.memory": "8Gi",
	})}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	// Entry override merged over the template resources
	pod := s.renderServicePod(configMap.Data, "gpuServices.yaml", brokerServiceVars())
	requests := pod.Spec.Containers[0].Resources.Requests
	s.Equal("2", requests.Cpu().String())
	s.Equal("8Gi", requests.Memory().String())

	// Entry without an override inherits the template resources
	pod = s.renderServicePod(configMap.Data, "cpuServices.yaml", brokerServiceVars())
	s.Equal("1", pod.Spec.Containers[0].Resources.Requests.Cpu().String())

	// Broker sizing vars do not affect the chart-owned sizing
	vars := brokerServiceVars()
	vars["_cpu"] = "4"
	vars["_memory"] = "16Gi"
	vars["_ephemeral_storage"] = "9Gi"
	vars["_gpu_count"] = 1
	pod = s.renderServicePod(configMap.Data, "gpuServices.yaml", vars)
	requests = pod.Spec.Containers[0].Resources.Requests
	s.Equal("2", requests.Cpu().String())
	s.Equal("8Gi", requests.Memory().String())
	_, limitsHaveGpu := pod.Spec.Containers[0].Resources.Limits[corev1.ResourceName("nvidia.com/gpu")]
	s.False(limitsHaveGpu, "broker _gpu_count must not size the pod")
}

// TestServiceNodeSelector verifies the merged nodeSelector applies and the
// broker's gpu_type var is ignored: node targeting is chart-owned, same
// as sizing.
func (s *doK8sConfigMapTemplateTest) TestServiceNodeSelector() {
	options := &helm.Options{SetValues: disableTelemetry(map[string]string{
		"delegatedOperatorJobTemplates.services.gpuServices.nodeSelector.cloud\\.google\\.com/gke-accelerator": "nvidia-tesla-t4",
		"delegatedOperatorJobTemplates.services.gpuServices.nodeSelector.disktype":                             "ssd",
	})}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	pod := s.renderServicePod(configMap.Data, "gpuServices.yaml", brokerServiceVars())
	s.Equal("nvidia-tesla-t4", pod.Spec.NodeSelector["cloud.google.com/gke-accelerator"])
	s.Equal("ssd", pod.Spec.NodeSelector["disktype"])

	vars := brokerServiceVars()
	vars["_gpu_type"] = "nvidia-h100-80gb"
	pod = s.renderServicePod(configMap.Data, "gpuServices.yaml", vars)
	s.Equal("nvidia-tesla-t4", pod.Spec.NodeSelector["cloud.google.com/gke-accelerator"])
	s.Equal("ssd", pod.Spec.NodeSelector["disktype"])
}

// TestServiceTelemetry verifies the telemetry wiring on service pods: the
// native sidecar targeting the in-pod service launcher, the shared socket,
// and their absence when telemetry is disabled.
func (s *doK8sConfigMapTemplateTest) TestServiceTelemetry() {
	const socketName = "telemetry-socket"

	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
		"delegatedOperatorJobTemplates.services.cpuServices.unused": "nil",
	}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	pod := s.renderServicePod(configMap.Data, "cpuServices.yaml", brokerServiceVars())

	s.Require().NotNil(pod.Spec.ShareProcessNamespace)
	s.True(*pod.Spec.ShareProcessNamespace)

	sidecar := findContainer(pod.Spec.InitContainers, "telemetry-sidecar")
	s.Require().NotNil(sidecar, "telemetry-sidecar initContainer not found")
	targetContainer, ok := envValue(sidecar.Env, "TARGET_CONTAINER")
	s.True(ok)
	s.Equal("service", targetContainer)
	serviceType, ok := envValue(sidecar.Env, "SERVICE_TYPE")
	s.True(ok)
	s.Equal("service", serviceType)

	s.NotNil(findVolume(pod.Spec.Volumes, socketName))
	s.NotNil(findVolumeMount(pod.Spec.Containers[0].VolumeMounts, socketName))
	socket, ok := envValue(pod.Spec.Containers[0].Env, "TELEMETRY_SOCKET")
	s.True(ok)
	s.Equal("/tmp/telemetry/agent.sock", socket)

	// Disabled: no sidecar, no socket plumbing
	options = &helm.Options{SetValues: disableTelemetry(map[string]string{
		"delegatedOperatorJobTemplates.services.cpuServices.unused": "nil",
	})}
	output = helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	pod = s.renderServicePod(configMap.Data, "cpuServices.yaml", brokerServiceVars())
	s.Empty(pod.Spec.InitContainers, "No sidecar with telemetry disabled")
	s.Nil(findVolume(pod.Spec.Volumes, socketName))
	_, ok = envValue(pod.Spec.Containers[0].Env, "TELEMETRY_SOCKET")
	s.False(ok)
}
