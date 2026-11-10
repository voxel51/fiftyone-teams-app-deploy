//go:build kubeall || helm || unit || unitInferenceEnv
// +build kubeall helm unit unitInferenceEnv

package unit

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/noirbizarre/gonja"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// The inference values wire FiftyOne to the fiftyone-inference chart.
type inferenceEnvTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestInferenceEnvTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &inferenceEnvTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

var inferenceOn = map[string]string{
	"inference.url":         "http://inference:8080",
	"inference.tokenSecret": "inference-token",
}

var labelerEnv = map[string]string{
	"LABELER_SERVICE_URL":       "http://inference-labeler:8080",
	"LABELER_SERVICE_AUTH":      "static-token",
	"LABELER_TEST_TRANSPORT":    "operator",
	"LABELER_RUN_MODE":          "job",
	"LABELER_DELEGATION_TARGET": "agentic-labeler",
}

// Every container that talks to the inference service: its template, the
// values that render it, the prefix of its own env and the ConfigMap key of
// a job or service template.
var inferenceComponents = []struct {
	name     string
	template string
	values   map[string]string
	prefix   string
	dataKey  string
}{
	{
		"fiftyone-app",
		"templates/app-deployment.yaml",
		map[string]string{},
		"appSettings",
		"",
	},
	{
		"teams-plugins",
		"templates/plugins-deployment.yaml",
		map[string]string{"pluginsSettings.enabled": "true"},
		"pluginsSettings",
		"",
	},
	{
		"delegated-operator-deployment",
		"templates/delegated-operator-instance-deployment.yaml",
		map[string]string{"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled": "true"},
		"delegatedOperatorDeployments.deployments.teamsDoCpuDefault",
		"",
	},
	{
		"delegated-operator-job-template",
		"templates/delegated-operator-job-configmap.yaml",
		map[string]string{"delegatedOperatorJobTemplates.jobs.cpuDefault.enabled": "true"},
		"delegatedOperatorJobTemplates.jobs.cpuDefault",
		"cpuDefault.yaml",
	},
	{
		"service-orchestrator",
		"templates/delegated-operator-job-configmap.yaml",
		map[string]string{},
		"delegatedOperatorJobTemplates.serviceOrchestrators.cpuServiceOrc",
		"cpuServiceOrc.yaml",
	},
}

// containerEnv renders the template and returns the env of its first
// container, the main one with telemetry off.
func (s *inferenceEnvTemplateTest) containerEnv(template, dataKey string, values map[string]string) []corev1.EnvVar {
	options := &helm.Options{SetValues: disableTelemetry(values)}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{template})

	if dataKey == "" {
		for _, doc := range splitYAMLDocs(output) {
			if strings.Contains(doc, "kind: Deployment") {
				var d appsv1.Deployment
				helm.UnmarshalK8SYaml(s.T(), doc, &d)
				return d.Spec.Template.Spec.Containers[0].Env
			}
		}
		s.FailNow("no Deployment rendered by " + template)
	}

	var configMap corev1.ConfigMap
	helm.UnmarshalK8SYaml(s.T(), output, &configMap)
	s.Require().Contains(configMap.Data, dataKey)
	tpl, err := gonja.FromString(configMap.Data[dataKey])
	s.Require().NoError(err)
	// Placeholders for both the job and the service broker variables
	rendered, err := tpl.Execute(map[string]interface{}{
		"_id": "task1", "_name": "task1", "_namespace": "ns", "_image": "img:tag",
		"_command": "fiftyone", "_args": []string{"test"}, "_env": []map[string]interface{}{},
		"_port": 8000, "_health_path": "", "_health_port": 8000,
	})
	s.Require().NoError(err)
	// A job template renders a Job, a service template a bare Pod
	if strings.Contains(rendered, "kind: Job") {
		var job batchv1.Job
		helm.UnmarshalK8SYaml(s.T(), rendered, &job)
		return job.Spec.Template.Spec.Containers[0].Env
	}
	var pod corev1.Pod
	helm.UnmarshalK8SYaml(s.T(), rendered, &pod)
	return pod.Spec.Containers[0].Env
}

func withValues(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func envNamed(env []corev1.EnvVar, name string) []corev1.EnvVar {
	found := []corev1.EnvVar{}
	for _, e := range env {
		if e.Name == name {
			found = append(found, e)
		}
	}
	return found
}

func (s *inferenceEnvTemplateTest) TestAbsentByDefault() {
	for _, tc := range inferenceComponents {
		s.Run(tc.name, func() {
			env := s.containerEnv(tc.template, tc.dataKey, tc.values)

			s.Empty(envNamed(env, "FIFTYONE_INFERENCE_URL"))
			s.Empty(envNamed(env, "FIFTYONE_INFERENCE_TOKEN"))
			s.Empty(envNamed(env, "FIFTYONE_INFERENCE_TENANT_ID"))
			for name := range labelerEnv {
				s.Empty(envNamed(env, name), name)
			}
		})
	}
}

func (s *inferenceEnvTemplateTest) TestUrlAndToken() {
	for _, tc := range inferenceComponents {
		s.Run(tc.name, func() {
			env := s.containerEnv(tc.template, tc.dataKey, withValues(tc.values, inferenceOn))

			url := envNamed(env, "FIFTYONE_INFERENCE_URL")
			s.Require().Len(url, 1)
			s.Equal("http://inference:8080", url[0].Value)

			token := envNamed(env, "FIFTYONE_INFERENCE_TOKEN")
			s.Require().Len(token, 1)
			s.Require().NotNil(token[0].ValueFrom)
			s.Equal("inference-token", token[0].ValueFrom.SecretKeyRef.Name)
			s.Equal("token", token[0].ValueFrom.SecretKeyRef.Key)

			for name := range labelerEnv {
				s.Empty(envNamed(env, name), name+" needs inference.labeler.url")
			}
		})
	}
}

func (s *inferenceEnvTemplateTest) TestLabeler() {
	for _, tc := range inferenceComponents {
		s.Run(tc.name, func() {
			values := withValues(tc.values, inferenceOn, map[string]string{
				"inference.labeler.url": "http://inference-labeler:8080",
			})
			env := s.containerEnv(tc.template, tc.dataKey, values)

			for name, value := range labelerEnv {
				found := envNamed(env, name)
				s.Require().Len(found, 1, name)
				s.Equal(value, found[0].Value, name)
			}
		})
	}
}

// A component's own env and secretEnv win, and each name renders once.
func (s *inferenceEnvTemplateTest) TestOwnEnvWins() {
	for _, tc := range inferenceComponents {
		s.Run(tc.name, func() {
			values := withValues(tc.values, inferenceOn, map[string]string{
				"inference.labeler.url":                                      "http://inference-labeler:8080",
				tc.prefix + ".env.FIFTYONE_INFERENCE_URL":                    "http://other:8080",
				tc.prefix + ".env.LABELER_RUN_MODE":                          "local",
				tc.prefix + ".secretEnv.FIFTYONE_INFERENCE_TOKEN.secretName": "my-secret",
				tc.prefix + ".secretEnv.FIFTYONE_INFERENCE_TOKEN.secretKey":  "my-key",
			})
			env := s.containerEnv(tc.template, tc.dataKey, values)

			url := envNamed(env, "FIFTYONE_INFERENCE_URL")
			s.Require().Len(url, 1)
			s.Equal("http://other:8080", url[0].Value)

			mode := envNamed(env, "LABELER_RUN_MODE")
			s.Require().Len(mode, 1)
			s.Equal("local", mode[0].Value)

			token := envNamed(env, "FIFTYONE_INFERENCE_TOKEN")
			s.Require().Len(token, 1)
			s.Equal("my-secret", token[0].ValueFrom.SecretKeyRef.Name)
			s.Equal("my-key", token[0].ValueFrom.SecretKeyRef.Key)

			seen := map[string]bool{}
			for _, e := range env {
				s.False(seen[e.Name], "duplicate env "+e.Name)
				seen[e.Name] = true
			}
		})
	}
}

// The shared template env of delegated operators wins too
func (s *inferenceEnvTemplateTest) TestTemplateEnvWins() {
	for _, tc := range inferenceComponents[2:4] {
		s.Run(tc.name, func() {
			shared := strings.SplitN(tc.prefix, ".", 2)[0] + ".template"
			values := withValues(tc.values, inferenceOn, map[string]string{
				shared + ".env.FIFTYONE_INFERENCE_URL": "http://shared:8080",
			})
			url := envNamed(s.containerEnv(tc.template, tc.dataKey, values), "FIFTYONE_INFERENCE_URL")
			s.Require().Len(url, 1)
			s.Equal("http://shared:8080", url[0].Value)
		})
	}
}

// One org: the org of calls with no user
func (s *inferenceEnvTemplateTest) TestTenantId() {
	for _, tc := range inferenceComponents {
		s.Run(tc.name, func() {
			values := withValues(tc.values, inferenceOn, map[string]string{"inference.tenantId": "org-1"})
			tenant := envNamed(s.containerEnv(tc.template, tc.dataKey, values), "FIFTYONE_INFERENCE_TENANT_ID")
			s.Require().Len(tenant, 1)
			s.Equal("org-1", tenant[0].Value)
		})
	}
}

// The labeler calls the LLM through the inference service
func (s *inferenceEnvTemplateTest) TestLabelerFailsWithoutUrl() {
	for _, tc := range inferenceComponents {
		s.Run(tc.name, func() {
			values := withValues(tc.values, map[string]string{"inference.labeler.url": "http://inference-labeler:8080"})
			options := &helm.Options{SetValues: disableTelemetry(values)}
			_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, []string{tc.template})

			s.Require().Error(err)
			s.Contains(err.Error(), "inference.url is empty")
		})
	}
}

func (s *inferenceEnvTemplateTest) TestFailsWithoutTokenSecret() {
	for _, tc := range inferenceComponents {
		s.Run(tc.name, func() {
			values := withValues(tc.values, map[string]string{"inference.url": "http://inference:8080"})
			options := &helm.Options{SetValues: disableTelemetry(values)}
			_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, []string{tc.template})

			s.Require().Error(err)
			s.Contains(err.Error(), "inference.tokenSecret is empty")
		})
	}
}
