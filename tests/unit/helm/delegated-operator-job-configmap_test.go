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
	"k8s.io/apimachinery/pkg/api/resource"
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

// TestTelemetryRedisUrlEnv asserts FIFTYONE_TELEMETRY_REDIS_URL is wired on
// both the workload container and the native-sidecar initContainer in the
// rendered DO Job. External URL must point at the user-supplied managed Redis;
// otherwise the bundled in-cluster Service URL must be release- and
// namespace-scoped. Regression: the DO templates env-vars helper previously
// built the URL inline via printf rather than going through the
// telemetry.redis.url helper, which silently ignored external.url.
func (s *doK8sConfigMapTemplateTest) TestTelemetryRedisUrlEnv() {
	jobKey := "cpuDefault.yaml"

	testCases := []struct {
		name        string
		values      map[string]string
		expectedURL string
	}{
		{
			"bundledRedisUrl",
			map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			fmt.Sprintf("redis://%s-telemetry-redis.fiftyone-teams.svc.cluster.local:6379", s.releaseName),
		},
		{
			"externalRedisUrl",
			map[string]string{
				"telemetry.enabled":                                    "true",
				"telemetry.redis.external.url":                         "redis://my-managed-redis:6379",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			"redis://my-managed-redis:6379",
		},
	}

	findEnv := func(envs []corev1.EnvVar, name string) *corev1.EnvVar {
		for i, ev := range envs {
			if ev.Name == name {
				return &envs[i]
			}
		}
		return nil
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

			s.Require().NotEmpty(job.Spec.Template.Spec.Containers, "expected at least one workload container")
			for _, container := range job.Spec.Template.Spec.Containers {
				ev := findEnv(container.Env, "FIFTYONE_TELEMETRY_REDIS_URL")
				s.Require().NotNil(ev,
					"FIFTYONE_TELEMETRY_REDIS_URL should be set on workload container %s", container.Name)
				s.Equal(testCase.expectedURL, ev.Value,
					"FIFTYONE_TELEMETRY_REDIS_URL on workload container %s should match expected URL",
					container.Name)
			}

			// The native-sidecar lives under initContainers (with
			// restartPolicy: Always) so a long-running sidecar does not block
			// Job completion.
			var sidecar *corev1.Container
			for i, c := range job.Spec.Template.Spec.InitContainers {
				if c.Name == "telemetry-sidecar" {
					sidecar = &job.Spec.Template.Spec.InitContainers[i]
					break
				}
			}
			s.Require().NotNil(sidecar, "telemetry-sidecar initContainer should be injected")
			ev := findEnv(sidecar.Env, "FIFTYONE_TELEMETRY_REDIS_URL")
			s.Require().NotNil(ev, "FIFTYONE_TELEMETRY_REDIS_URL should be set on telemetry-sidecar")
			s.Equal(testCase.expectedURL, ev.Value,
				"FIFTYONE_TELEMETRY_REDIS_URL on telemetry-sidecar should match expected URL")
		})
	}
}

// TestShareProcessNamespace asserts the rendered DO Job pod opts into
// PID-namespace sharing when telemetry is on — the native-sidecar reads
// /proc/<pid>/fd/1 in the workload's PID namespace and would silently fail to
// capture logs without this. With telemetry off, the share must not be set.
func (s *doK8sConfigMapTemplateTest) TestShareProcessNamespace() {
	jobKey := "cpuDefault.yaml"

	testCases := []struct {
		name              string
		values            map[string]string
		expectShareNSTrue bool
	}{
		{
			"telemetryEnabled",
			map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			true,
		},
		{
			"telemetryDisabled",
			map[string]string{
				"telemetry.enabled": "false",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			false,
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

			share := job.Spec.Template.Spec.ShareProcessNamespace
			if testCase.expectShareNSTrue {
				s.Require().NotNil(share, "shareProcessNamespace should be set on Job pod")
				s.True(*share, "shareProcessNamespace should be true on Job pod")
			} else {
				if share != nil {
					s.False(*share, "shareProcessNamespace should not be true when telemetry is disabled")
				}
			}
		})
	}
}

// findInitContainerByName returns a pointer into the slice for the first
// initContainer with the given name, or nil. Mirrors findContainerByName
// (defined in telemetry_sidecar_helpers_test.go) for spec.initContainers.
func findInitContainerByName(containers []corev1.Container, name string) *corev1.Container {
	for i, c := range containers {
		if c.Name == name {
			return &containers[i]
		}
	}
	return nil
}

func (s *doK8sConfigMapTemplateTest) TestSidecarContainerImage() {
	jobKey := "cpuDefault.yaml"

	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)
	chartAppVersion, exists := cInfo["appVersion"]
	s.True(exists, "failed to get app version from chart info")

	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultValues",
			map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			fmt.Sprintf("voxel51/telemetry-sidecar:%s", chartAppVersion),
		},
		{
			"overrideImageRepositoryAndTag",
			map[string]string{
				"telemetry.enabled":                                    "true",
				"telemetry.sidecar.image.repository":                   "my-registry/telemetry-sidecar",
				"telemetry.sidecar.image.tag":                          "v1.2.3",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			"my-registry/telemetry-sidecar:v1.2.3",
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
			sidecar := findInitContainerByName(job.Spec.Template.Spec.InitContainers, "telemetry-sidecar")
			s.Require().NotNil(sidecar, "telemetry-sidecar initContainer should be injected")
			s.Equal(testCase.expected, sidecar.Image)
		})
	}
}

func (s *doK8sConfigMapTemplateTest) TestSidecarContainerResourceRequirements() {
	jobKey := "cpuDefault.yaml"

	testCases := []struct {
		name     string
		values   map[string]string
		expected func(r corev1.ResourceRequirements)
	}{
		{
			"defaultValues",
			map[string]string{
				"telemetry.enabled": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			func(r corev1.ResourceRequirements) {
				expected := corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    resource.MustParse("100m"),
						"memory": resource.MustParse("512Mi"),
					},
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("100m"),
						"memory": resource.MustParse("512Mi"),
					},
				}
				s.Equal(expected, r, "default sidecar resources should match chart defaults")
			},
		},
		{
			"overrideResources",
			map[string]string{
				"telemetry.enabled":                                    "true",
				"telemetry.sidecar.resources.limits.cpu":               "200m",
				"telemetry.sidecar.resources.limits.memory":            "1Gi",
				"telemetry.sidecar.resources.requests.cpu":             "200m",
				"telemetry.sidecar.resources.requests.memory":          "768Mi",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
			},
			func(r corev1.ResourceRequirements) {
				expected := corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    resource.MustParse("200m"),
						"memory": resource.MustParse("1Gi"),
					},
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("200m"),
						"memory": resource.MustParse("768Mi"),
					},
				}
				s.Equal(expected, r)
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

			job := s.renderJob(configMap.Data, jobKey)
			sidecar := findInitContainerByName(job.Spec.Template.Spec.InitContainers, "telemetry-sidecar")
			s.Require().NotNil(sidecar, "telemetry-sidecar initContainer should be injected")
			testCase.expected(sidecar.Resources)
		})
	}
}

// TestSidecarContainerSecurityContext asserts the DO Job's native-sidecar
// drops all default capabilities, disables privilege escalation, AND adds
// SYS_PTRACE — DO Jobs are executor pods where the sidecar archives py-spy
// crash stacks from the workload's PID namespace.
func (s *doK8sConfigMapTemplateTest) TestSidecarContainerSecurityContext() {
	jobKey := "cpuDefault.yaml"

	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
		"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
	}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	job := s.renderJob(configMap.Data, jobKey)
	sidecar := findInitContainerByName(job.Spec.Template.Spec.InitContainers, "telemetry-sidecar")
	s.Require().NotNil(sidecar, "telemetry-sidecar initContainer should be injected")

	sc := sidecar.SecurityContext
	s.Require().NotNil(sc, "telemetry-sidecar should have a securityContext")
	s.Require().NotNil(sc.AllowPrivilegeEscalation)
	s.False(*sc.AllowPrivilegeEscalation, "sidecar must not allow privilege escalation")
	s.Require().NotNil(sc.Capabilities)
	s.Contains(sc.Capabilities.Drop, corev1.Capability("ALL"),
		"sidecar must drop all default capabilities")

	var hasPtrace bool
	for _, capability := range sc.Capabilities.Add {
		if capability == "SYS_PTRACE" {
			hasPtrace = true
			break
		}
	}
	s.True(hasPtrace, "executor sidecar must add SYS_PTRACE for py-spy crash archives")
}

// TestSidecarContainerRestartPolicyAndProbe asserts the DO Job native-sidecar
// is configured to NOT block Job completion (restartPolicy: Always makes it a
// native-sidecar initContainer) and that the workload waits for the agent
// socket before starting (readinessProbe).
func (s *doK8sConfigMapTemplateTest) TestSidecarContainerRestartPolicyAndProbe() {
	jobKey := "cpuDefault.yaml"

	options := &helm.Options{SetValues: map[string]string{
		"telemetry.enabled": "true",
		"delegatedOperatorJobTemplates.jobs.cpuDefault.unused": "nil",
	}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)

	job := s.renderJob(configMap.Data, jobKey)
	sidecar := findInitContainerByName(job.Spec.Template.Spec.InitContainers, "telemetry-sidecar")
	s.Require().NotNil(sidecar, "telemetry-sidecar initContainer should be injected")

	// native-sidecar: initContainer with restartPolicy=Always so it runs
	// alongside the workload but does not block Job completion.
	s.Require().NotNil(sidecar.RestartPolicy, "native-sidecar must set restartPolicy")
	s.Equal(corev1.ContainerRestartPolicyAlways, *sidecar.RestartPolicy)

	s.Require().NotNil(sidecar.ReadinessProbe, "native-sidecar must have a readinessProbe")
	s.Require().NotNil(sidecar.ReadinessProbe.Exec, "readinessProbe must use exec")
	s.NotEmpty(sidecar.ReadinessProbe.Exec.Command, "readinessProbe exec command must be set")
}

func (s *doK8sConfigMapTemplateTest) loadTestFile(filename, chartVersion string) string {
	content, err := os.ReadFile(filename)
	s.NoError(err, "Failed to read test file: %s", filename)
	result := strings.ReplaceAll(string(content), "{{CHART_VERSION}}", chartVersion)
	result = strings.TrimSpace(result)
	return strings.TrimSpace(result)
}
