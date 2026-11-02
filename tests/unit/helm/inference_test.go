//go:build kubeall || helm || unit || unitInference
// +build kubeall helm unit unitInference

package unit

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const inferenceChartPath = "../../../helm/fiftyone-inference/"

const inferenceToken = "test-0123456789abcdef0123456789abcdef"

type inferenceTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestInferenceTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(inferenceChartPath)
	require.NoError(t, err)

	suite.Run(t, &inferenceTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "inference",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

// The value every render needs: the service refuses to start without a token
func withToken(extra map[string]string) map[string]string {
	values := map[string]string{"sharedToken.value": inferenceToken}
	for key, value := range extra {
		values[key] = value
	}
	return values
}

// An OpenAI compatible backend beside the default model server
func withLlm(extra map[string]string) map[string]string {
	values := map[string]string{
		"backends[0].kind":                                 "openai",
		"backends[0].id":                                   "openai",
		"backends[0].baseUrl":                              "https://api.openai.com/v1",
		"backends[0].auth.mode":                            "bearer-token",
		"backends[0].auth.tokenEnv":                        "OPENAI_API_KEY",
		"backends[0].models[0].name":                       "qwen3-vl",
		"backends[0].models[0].tasks[0]":                   "generation",
		"backends[0].models[0].contextWindow":              "131072",
		"backends[0].models[0].limits.maxImagesPerRequest": "64",
		"backends[0].models[1].name":                       "other",
		"backends[0].models[1].tasks[0]":                   "generation",
		"backends[0].models[0].wireFormat":                 "openai-chat",
		"backends[0].models[1].wireFormat":                 "openai-chat",
		"secretEnv.OPENAI_API_KEY.secretName":              "openai",
		"secretEnv.OPENAI_API_KEY.secretKey":               "apiKey",
	}
	for key, value := range extra {
		values[key] = value
	}
	return withToken(values)
}

func (s *inferenceTemplateTest) render(values map[string]string, template string) string {
	options := &helm.Options{SetValues: values}
	return helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{template})
}

func (s *inferenceTemplateTest) deployment(values map[string]string, template string) appsv1.Deployment {
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), s.render(values, template), &deployment)
	return deployment
}

func backendsOf(raw string) []map[string]any {
	var backends []map[string]any
	if err := yaml.Unmarshal([]byte(raw), &backends); err != nil {
		panic(err)
	}
	return backends
}

func envOf(container corev1.Container) map[string]corev1.EnvVar {
	env := map[string]corev1.EnvVar{}
	for _, variable := range container.Env {
		env[variable.Name] = variable
	}
	return env
}

func (s *inferenceTemplateTest) TestRefusesWhatCannotWork() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"noToken",
			nil,
			"sharedToken.secretName or sharedToken.value is required",
		},
		{
			// The service trims, so whitespace does not count
			"shortToken",
			map[string]string{"sharedToken.value": "  0123456789abcdef0123456789abcde  "},
			"sharedToken.value must be at least 32 characters",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, []string{})

			s.ErrorContains(err, testCase.expected)
		})
	}
}

func (s *inferenceTemplateTest) TestTokenFromAValue() {
	var secret corev1.Secret
	helm.UnmarshalK8SYaml(s.T(), s.render(withToken(nil), "templates/secret.yaml"), &secret)
	s.Equal("inference-token", secret.Name)
	s.Equal(inferenceToken, secret.StringData["token"])

	deployment := s.deployment(withToken(nil), "templates/deployment.yaml")
	token := envOf(deployment.Spec.Template.Spec.Containers[0])["AUTH_SHARED_TOKEN"]
	s.Equal("inference-token", token.ValueFrom.SecretKeyRef.Name)
	s.Equal("token", token.ValueFrom.SecretKeyRef.Key)

	// A new token restarts the pod, which reads it at start only
	other := s.deployment(withToken(map[string]string{"sharedToken.value": inferenceToken + "x"}), "templates/deployment.yaml")
	s.NotEqual(deployment.Spec.Template.Annotations["checksum/token"], other.Spec.Template.Annotations["checksum/token"])
}

func (s *inferenceTemplateTest) TestTokenFromAnExistingSecret() {
	values := map[string]string{"sharedToken.secretName": "inference-shared"}
	deployment := s.deployment(values, "templates/deployment.yaml")

	token := envOf(deployment.Spec.Template.Spec.Containers[0])["AUTH_SHARED_TOKEN"]
	s.Equal("inference-shared", token.ValueFrom.SecretKeyRef.Name)
	s.Empty(deployment.Spec.Template.Annotations)
	_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: values}, s.chartPath, s.releaseName, []string{"templates/secret.yaml"})
	s.ErrorContains(err, "could not find template templates/secret.yaml in chart")
}

func (s *inferenceTemplateTest) TestLicenseModeWithTheModelServer() {
	deployment := s.deployment(withToken(nil), "templates/deployment.yaml")
	spec := deployment.Spec.Template.Spec
	container := spec.Containers[0]
	env := envOf(container)

	s.Equal(int32(1), *deployment.Spec.Replicas)
	s.Equal("voxel51/inference:v2.25.1", container.Image)
	s.Equal("shared-token", env["AUTH_MODE"].Value)
	s.Equal("license", env["BILLING_MODE"].Value)
	s.Equal("/opt/fiftyone/licenses/fiftyone-license", env["LICENSE_KEY_FILE_PATHS"].Value)
	s.Equal("memory", env["STORE_MODE"].Value)
	s.Equal("67108864", env["MAX_BODY_BYTES"].Value)
	s.Equal("8080", env["PORT"].Value)
	// The shipped model server backend, pointed at the model server chart
	backends := backendsOf(env["BACKENDS"].Value)
	s.Len(backends, 1)
	s.Equal("model-server", backends[0]["id"])
	s.Equal("model-server", backends[0]["kind"])
	s.Equal([]any{map[string]any{"url": "http://model-server-0:8080"}}, backends[0]["endpoints"])

	// The license is mounted where LICENSE_KEY_FILE_PATHS points
	s.Equal("/opt/fiftyone/licenses", container.VolumeMounts[0].MountPath)
	s.True(container.VolumeMounts[0].ReadOnly)
	sources := spec.Volumes[0].Projected.Sources
	s.Len(sources, 1)
	s.Equal("fiftyone-license", sources[0].Secret.Name)
	s.Equal("license", sources[0].Secret.Items[0].Key)
	s.Equal("fiftyone-license", sources[0].Secret.Items[0].Path)
	// No token for the model servers unless one is set
	s.Equal(map[string]any{"mode": "none"}, backends[0]["auth"])
	s.NotContains(env, "MODEL_SERVER_TOKEN")

	// The fiftyone-model-server chart admits callers with this label
	s.Equal("inference", deployment.Spec.Template.Labels["app"])
	s.Equal("/healthz", container.LivenessProbe.HTTPGet.Path)
	s.Equal("/readyz", container.ReadinessProbe.HTTPGet.Path)
	s.Equal(int64(10), container.Lifecycle.PreStop.Sleep.Seconds)
	s.Equal(int64(360), *spec.TerminationGracePeriodSeconds)
	s.False(*spec.AutomountServiceAccountToken)
}

func (s *inferenceTemplateTest) TestSettings() {
	deployment := s.deployment(withToken(map[string]string{
		"image.repository":               "registry.example/inference",
		"image.tag":                      "v1",
		"imagePullSecrets[0].name":       "registry",
		"licenseSecrets[0]":              "org-a-license",
		"licenseSecrets[1]":              "org-b-license",
		"storeMode":                      "mongo",
		"maxBodyBytes":                   "134217728",
		"env.LOG_LEVEL":                  "debug",
		"secretEnv.MONGO_URI.secretName": "mongo",
		"secretEnv.MONGO_URI.secretKey":  "uri",
		"resources.limits.memory":        "1Gi",
	}), "templates/deployment.yaml")
	spec := deployment.Spec.Template.Spec
	container := spec.Containers[0]
	env := envOf(container)

	s.Equal("registry.example/inference:v1", container.Image)
	s.Equal("registry", spec.ImagePullSecrets[0].Name)
	// One license per org, each file at its own path
	s.Equal("/opt/fiftyone/licenses/org-a-license,/opt/fiftyone/licenses/org-b-license", env["LICENSE_KEY_FILE_PATHS"].Value)
	sources := spec.Volumes[0].Projected.Sources
	s.Len(sources, 2)
	for i, name := range []string{"org-a-license", "org-b-license"} {
		s.Equal(name, sources[i].Secret.Name)
		s.Equal("license", sources[i].Secret.Items[0].Key)
		s.Equal(name, sources[i].Secret.Items[0].Path)
	}
	s.Equal("mongo", env["STORE_MODE"].Value)
	s.Equal("134217728", env["MAX_BODY_BYTES"].Value)
	s.Equal("debug", env["LOG_LEVEL"].Value)
	s.Equal("mongo", env["MONGO_URI"].ValueFrom.SecretKeyRef.Name)
	s.Equal("uri", env["MONGO_URI"].ValueFrom.SecretKeyRef.Key)
	s.Equal(resource.MustParse("1Gi"), container.Resources.Limits[corev1.ResourceMemory])
}

func (s *inferenceTemplateTest) TestNoLabelerWithoutAnLlm() {
	for _, template := range []string{"templates/labeler-deployment.yaml", "templates/labeler-service.yaml", "templates/labeler-networkpolicy.yaml"} {
		_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: withToken(nil)}, s.chartPath, s.releaseName, []string{template})
		s.ErrorContains(err, "could not find template "+template+" in chart", template)
	}
}

func (s *inferenceTemplateTest) TestLlmBackend() {
	deployment := s.deployment(withLlm(nil), "templates/deployment.yaml")
	env := envOf(deployment.Spec.Template.Spec.Containers[0])

	s.Contains(env["BACKENDS"].Value, "kind: openai")
	s.Contains(env["BACKENDS"].Value, "tokenEnv: OPENAI_API_KEY")
	s.Equal("openai", env["OPENAI_API_KEY"].ValueFrom.SecretKeyRef.Name)
	s.Equal("apiKey", env["OPENAI_API_KEY"].ValueFrom.SecretKeyRef.Key)
}

func (s *inferenceTemplateTest) TestLabelerWithAnLlm() {
	deployment := s.deployment(withLlm(nil), "templates/labeler-deployment.yaml")
	spec := deployment.Spec.Template.Spec
	container := spec.Containers[0]
	env := envOf(container)

	s.Equal("inference-labeler", deployment.Name)
	s.Equal(int32(1), *deployment.Spec.Replicas)
	s.Equal("voxel51/agentic-labeler-cpu:v2.25.1", container.Image)
	s.Equal("shared", env["LABELER_SERVICE_MODE"].Value)
	s.Equal("http://inference:8080", env["FIFTYONE_INFERENCE_URL"].Value)
	s.Equal("inference-token", env["FIFTYONE_INFERENCE_TOKEN"].ValueFrom.SecretKeyRef.Name)
	s.Equal("token", env["FIFTYONE_INFERENCE_TOKEN"].ValueFrom.SecretKeyRef.Key)
	// It checks the shared token its callers send
	s.Equal("shared-token", env["LABELER_AUTH_MODE"].Value)
	s.Equal("1", env["LABELER_MAX_IN_FLIGHT_TESTS"].Value)
	// The token is sent only when the engine config names no key of its own
	s.NotContains(env, "FIFTYONE_INFERENCE_AUTH")
	s.Equal("/healthz", container.LivenessProbe.HTTPGet.Path)
	s.Equal("/readyz", container.ReadinessProbe.HTTPGet.Path)
	s.Equal(int64(180), *spec.TerminationGracePeriodSeconds)
	s.Equal(int64(5), container.Lifecycle.PreStop.Sleep.Seconds)
	s.False(*spec.AutomountServiceAccountToken)
	s.True(*spec.SecurityContext.RunAsNonRoot)
	s.Equal(int64(1000), *spec.SecurityContext.RunAsUser)
	s.Equal(int64(1000), *spec.SecurityContext.RunAsGroup)
	// A new token restarts the labeler with the inference service
	s.NotEmpty(deployment.Spec.Template.Annotations["checksum/token"])

	var service corev1.Service
	helm.UnmarshalK8SYaml(s.T(), s.render(withLlm(nil), "templates/labeler-service.yaml"), &service)
	s.Equal("inference-labeler", service.Name)
	s.Equal(int32(8080), service.Spec.Ports[0].Port)
	s.Equal(deployment.Spec.Selector.MatchLabels, service.Spec.Selector)

	var policy networkingv1.NetworkPolicy
	helm.UnmarshalK8SYaml(s.T(), s.render(withLlm(nil), "templates/labeler-networkpolicy.yaml"), &policy)
	s.Equal(deployment.Spec.Selector.MatchLabels, policy.Spec.PodSelector.MatchLabels)
	// Every pod in this namespace, and nothing else, on the service port only
	s.Equal([]networkingv1.PolicyType{networkingv1.PolicyTypeIngress}, policy.Spec.PolicyTypes)
	s.Len(policy.Spec.Ingress, 1)
	s.Len(policy.Spec.Ingress[0].From, 1)
	s.Empty(policy.Spec.Ingress[0].From[0].PodSelector.MatchLabels)
	s.Nil(policy.Spec.Ingress[0].From[0].NamespaceSelector)
	s.Len(policy.Spec.Ingress[0].Ports, 1)
	s.Equal(int32(8080), policy.Spec.Ingress[0].Ports[0].Port.IntVal)
}

func (s *inferenceTemplateTest) TestLabelerSettings() {
	deployment := s.deployment(withLlm(map[string]string{
		"sharedToken.value":                       "",
		"sharedToken.secretName":                  "inference-shared",
		"labeler.image.repository":                "registry.example/labeler",
		"labeler.image.tag":                       "v1",
		"labeler.env.LABELER_MAX_IN_FLIGHT_TESTS": "2",
		"labeler.env.LABELER_FETCH_ALLOWED_HOSTS": `["minio.internal"]`,
		"labeler.resources.limits.memory":         "4Gi",
	}), "templates/labeler-deployment.yaml")
	container := deployment.Spec.Template.Spec.Containers[0]
	env := envOf(container)

	s.Equal("registry.example/labeler:v1", container.Image)
	s.Equal("inference-shared", env["FIFTYONE_INFERENCE_TOKEN"].ValueFrom.SecretKeyRef.Name)
	s.Equal("2", env["LABELER_MAX_IN_FLIGHT_TESTS"].Value)
	s.Equal(`["minio.internal"]`, env["LABELER_FETCH_ALLOWED_HOSTS"].Value)
	s.Equal(resource.MustParse("4Gi"), container.Resources.Limits[corev1.ResourceMemory])
	s.Len(container.Env, 8)
}

// Any OpenAI compatible LLM: Chat Completions through the inference service,
// on the first model of the first LLM backend, with the shared token
func (s *inferenceTemplateTest) TestLabelerOnChatCompletions() {
	for name, values := range map[string]map[string]string{
		"compatible":     withLlm(nil),
		"filesNotHosted": withLlm(map[string]string{"backends[0].files": "true"}),
		"hostedNoFiles":  withLlm(map[string]string{"backends[0].openaiHosted": "true"}),
		"secondLlm": withLlm(map[string]string{
			"backends[1].kind":               "openai",
			"backends[1].id":                 "later",
			"backends[1].baseUrl":            "https://llm.example/v1",
			"backends[1].models[0].name":     "later-model",
			"backends[1].models[0].tasks[0]": "generation",
		}),
		// Only the backend the labeler uses must declare models
		"laterWithoutModels": withLlm(map[string]string{
			"backends[1].kind":    "openai",
			"backends[1].id":      "later",
			"backends[1].baseUrl": "https://llm.example/v1",
		}),
	} {
		env := envOf(s.deployment(values, "templates/labeler-deployment.yaml").Spec.Template.Spec.Containers[0])

		s.Equal("/app/configs/remote-vllm-openai.json", env["LABELER_CONFIG_FILE"].Value, name)
		var overlay map[string]any
		s.Require().NoError(json.Unmarshal([]byte(env["LABELER_ENGINE_CONFIG"].Value), &overlay), name)
		s.Equal(map[string]any{
			"base_url_env":      "LABELER_INFERENCE_BASE_URL",
			"api_key_env":       nil,
			"model":             "qwen3-vl",
			"max_pixels":        float64(1048576),
			"structured_output": "auto",
			"max_images":        float64(64),
		}, overlay, name)
		// The labeler's image token estimate is Gemma's: no context window
		s.NotContains(env, "LABELER_MODEL_LIMITS", name)
	}
}

// The labeler refuses every test with examples unless it can size them
func (s *inferenceTemplateTest) TestLabelerLimitsFromTheModel() {
	overlayOf := func(env map[string]corev1.EnvVar) map[string]any {
		var overlay map[string]any
		s.Require().NoError(json.Unmarshal([]byte(env["LABELER_ENGINE_CONFIG"].Value), &overlay))
		return overlay
	}

	// The image count is the only limit: a declared contextWindow is not
	// passed on, as the labeler's image token estimate is Gemma's
	env := envOf(s.deployment(withLlm(nil), "templates/labeler-deployment.yaml").Spec.Template.Spec.Containers[0])
	s.Equal(float64(64), overlayOf(env)["max_images"])
	s.NotContains(env, "LABELER_MODEL_LIMITS")
}

func (s *inferenceTemplateTest) TestRefusesWhatTheLabelerCannotUse() {
	// A context window alone leaves the request size unbounded
	noLimits := withLlm(nil)
	delete(noLimits, "backends[0].models[0].limits.maxImagesPerRequest")
	strings := withLlm(nil)
	for _, key := range []string{"name", "tasks[0]", "wireFormat", "contextWindow", "limits.maxImagesPerRequest"} {
		delete(strings, "backends[0].models[0]."+key)
	}
	strings["backends[0].models[0]"] = "qwen3-vl"
	notAList := withLlm(nil)
	for key := range notAList {
		if len(key) > len("backends[0].models") && key[:len("backends[0].models")] == "backends[0].models" {
			delete(notAList, key)
		}
	}
	notAList["backends[0].models"] = "qwen3-vl"
	solFiles := map[string]string{
		"backends[0].openaiHosted":         "true",
		"backends[0].files":                "true",
		"backends[0].models[1].name":       "gpt-6-sol-responses",
		"backends[0].models[1].wireFormat": "openai-responses",
	}
	solWrongFormat := withLlm(solFiles)
	solWrongFormat["backends[0].models[1].wireFormat"] = "openai-chat"
	solNoGeneration := withLlm(solFiles)
	solNoGeneration["backends[0].models[1].tasks[0]"] = "embedding"

	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{"noLimits", noLimits, "model qwen3-vl of backend openai declares no limits.maxImagesPerRequest"},
		{"modelAsString", strings, "backend openai: each model must be an object with a name"},
		{"modelsNotAList", notAList, "backend openai: models must be a list"},
		{"chatWrongFormat", withLlm(map[string]string{"backends[0].models[0].wireFormat": "openai-responses"}), "model qwen3-vl of backend openai needs wireFormat: openai-chat"},
		{"chatNoFormat", withLlm(map[string]string{"backends[0].models[0].wireFormat": ""}), "model qwen3-vl of backend openai needs wireFormat: openai-chat"},
		{"chatNoGeneration", withLlm(map[string]string{"backends[0].models[0].tasks[0]": "embedding"}), "model qwen3-vl of backend openai needs generation in its tasks"},
		{"solWrongFormat", solWrongFormat, "model gpt-6-sol-responses of backend openai needs wireFormat: openai-responses"},
		{"solNoGeneration", solNoGeneration, "model gpt-6-sol-responses of backend openai needs generation in its tasks"},
		{"filesWithoutSol", withLlm(map[string]string{
			"backends[0].openaiHosted": "true",
			"backends[0].files":        "true",
		}), "backend openai is OpenAI with files, so the labeler calls its model gpt-6-sol-responses"},
	}
	for _, testCase := range testCases {
		_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: testCase.values}, s.chartPath, s.releaseName, []string{})
		s.ErrorContains(err, testCase.expected, testCase.name)
	}
}

// Values in labeler.env win over the ones the chart sets
func (s *inferenceTemplateTest) TestLabelerEnvWins() {
	options := &helm.Options{
		SetValues: withLlm(map[string]string{"labeler.env.LABELER_CONFIG_FILE": "/app/configs/mine.json"}),
		SetJsonValues: map[string]string{
			"labeler.env.LABELER_ENGINE_CONFIG": `"{\"model\":\"mine\"}"`,
			"labeler.env.LABELER_MODEL_LIMITS":  `"{\"context_length\":1}"`,
		},
	}
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{"templates/labeler-deployment.yaml"}), &deployment)
	env := envOf(deployment.Spec.Template.Spec.Containers[0])

	s.Equal("/app/configs/mine.json", env["LABELER_CONFIG_FILE"].Value)
	s.Equal(`{"model":"mine"}`, env["LABELER_ENGINE_CONFIG"].Value)
	s.Equal(`{"context_length":1}`, env["LABELER_MODEL_LIMITS"].Value)
}

// Pod settings of both pods: hardened by default, scheduling from values
func (s *inferenceTemplateTest) TestPodSettings() {
	for _, template := range []string{"templates/deployment.yaml", "templates/labeler-deployment.yaml"} {
		spec := s.deployment(withLlm(nil), template).Spec.Template.Spec
		s.True(*spec.SecurityContext.RunAsNonRoot, template)
		s.Equal(int64(1000), *spec.SecurityContext.RunAsUser, template)
		s.Equal(int64(1000), *spec.SecurityContext.RunAsGroup, template)
		s.False(*spec.Containers[0].SecurityContext.AllowPrivilegeEscalation, template)
		s.Empty(spec.NodeSelector, template)
		s.Empty(spec.Tolerations, template)
		s.Nil(spec.Affinity, template)

		spec = s.deployment(withLlm(map[string]string{
			"podSecurityContext.runAsNonRoot":        "false",
			"podSecurityContext.fsGroup":             "1000",
			"securityContext.readOnlyRootFilesystem": "true",
			"nodeSelector.pool":                      "cpu",
			"tolerations[0].key":                     "dedicated",
			"tolerations[0].operator":                "Exists",
			"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":      "pool",
			"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator": "Exists",
			"imagePullSecrets[0].name": "registry",
			"image.pullPolicy":         "Always",
		}), template).Spec.Template.Spec
		// Never root, whatever the values say
		s.True(*spec.SecurityContext.RunAsNonRoot, template)
		s.Equal(int64(1000), *spec.SecurityContext.FSGroup, template)
		s.True(*spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem, template)
		s.False(*spec.Containers[0].SecurityContext.AllowPrivilegeEscalation, template)
		s.Equal("cpu", spec.NodeSelector["pool"], template)
		s.Equal("dedicated", spec.Tolerations[0].Key, template)
		s.Equal("pool", spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, template)
		s.Equal("registry", spec.ImagePullSecrets[0].Name, template)
		s.Equal(corev1.PullAlways, spec.Containers[0].ImagePullPolicy, template)
	}
}

// OpenAI's own API with files: gpt-6-sol on the Responses API, as shipped
func (s *inferenceTemplateTest) TestLabelerOnOpenAIWithFiles() {
	for name, values := range map[string]map[string]string{
		"only": withLlm(map[string]string{
			"backends[0].openaiHosted":         "true",
			"backends[0].files":                "true",
			"backends[0].models[1].name":       "gpt-6-sol-responses",
			"backends[0].models[1].wireFormat": "openai-responses",
		}),
		// Wins over an earlier compatible backend
		"second": withLlm(map[string]string{
			"backends[1].kind":                 "openai",
			"backends[1].id":                   "openai",
			"backends[1].baseUrl":              "https://api.openai.com/v1",
			"backends[1].openaiHosted":         "true",
			"backends[1].files":                "true",
			"backends[1].models[0].name":       "gpt-6-sol-responses",
			"backends[1].models[0].wireFormat": "openai-responses",
			"backends[1].models[0].tasks[0]":   "generation",
		}),
	} {
		env := envOf(s.deployment(values, "templates/labeler-deployment.yaml").Spec.Template.Spec.Containers[0])

		s.Equal("/app/configs/remote-gpt-6-sol.json", env["LABELER_CONFIG_FILE"].Value, name)
		s.NotContains(env, "LABELER_ENGINE_CONFIG", name)
	}
}

func (s *inferenceTemplateTest) TestRefusesAnLlmWithNoModels() {
	values := withLlm(nil)
	delete(values, "backends[0].models[0].name")
	delete(values, "backends[0].models[0].tasks[0]")
	delete(values, "backends[0].models[1].name")
	delete(values, "backends[0].models[1].tasks[0]")
	delete(values, "backends[0].models[0].wireFormat")
	delete(values, "backends[0].models[1].wireFormat")
	delete(values, "backends[0].models[0].contextWindow")
	delete(values, "backends[0].models[0].limits.maxImagesPerRequest")
	_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: values}, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "backend openai declares no models")
}

func (s *inferenceTemplateTest) TestInTheReleaseNamespace() {
	options := &helm.Options{SetValues: withLlm(nil)}
	for _, template := range []string{"templates/secret.yaml", "templates/deployment.yaml", "templates/service.yaml", "templates/labeler-deployment.yaml", "templates/labeler-service.yaml", "templates/labeler-networkpolicy.yaml"} {
		output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{template}, "--namespace", s.namespace)
		var object corev1.ConfigMap
		helm.UnmarshalK8SYaml(s.T(), output, &object)
		s.Equal(s.namespace, object.Namespace, template)
	}
}

func (s *inferenceTemplateTest) TestServiceTheTeamsAppCalls() {
	var service corev1.Service
	helm.UnmarshalK8SYaml(s.T(), s.render(withToken(nil), "templates/service.yaml"), &service)
	deployment := s.deployment(withToken(nil), "templates/deployment.yaml")

	s.Equal("inference", service.Name)
	s.Equal(int32(8080), service.Spec.Ports[0].Port)
	s.Equal("http", service.Spec.Ports[0].TargetPort.StrVal)
	s.Equal(deployment.Spec.Selector.MatchLabels, service.Spec.Selector)
}

// Every model the model server image serves, shipped with the chart
func (s *inferenceTemplateTest) TestShippedModelServerBackend() {
	backendsFor := func(values map[string]string, json map[string]string) []map[string]any {
		options := &helm.Options{SetValues: values, SetJsonValues: json}
		var deployment appsv1.Deployment
		helm.UnmarshalK8SYaml(s.T(), helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{"templates/deployment.yaml"}), &deployment)
		return backendsOf(envOf(deployment.Spec.Template.Spec.Containers[0])["BACKENDS"].Value)
	}

	shipped := backendsFor(withToken(nil), nil)[0]
	var names []string
	for _, model := range shipped["models"].([]any) {
		names = append(names, model.(map[string]any)["name"].(string))
	}
	s.Equal([]string{
		"qwen3-vl-embedding-2b-torch",
		"segment-anything-2-hiera-large-image-torch",
		"segment-anything-2-hiera-large-video-torch",
		"siglip-base-patch16-224-torch",
	}, names)
	s.Equal(map[string]any{"mode": "none"}, shipped["auth"])

	// One endpoint per pod of the model server release, in order
	pods := backendsFor(withToken(map[string]string{"modelServer.replicas": "3"}), nil)
	s.Equal([]any{
		map[string]any{"url": "http://model-server-0:8080"},
		map[string]any{"url": "http://model-server-1:8080"},
		map[string]any{"url": "http://model-server-2:8080"},
	}, pods[0]["endpoints"])
	named := backendsFor(withToken(map[string]string{"modelServer.name": "gpu", "modelServer.replicas": "2"}), nil)
	s.Equal([]any{map[string]any{"url": "http://gpu-0:8080"}, map[string]any{"url": "http://gpu-1:8080"}}, named[0]["endpoints"])

	// Listed URLs win, for model servers elsewhere
	several := backendsFor(withToken(map[string]string{"modelServer.replicas": "3"}), map[string]string{"modelServer.urls": `["http://box-a:8080","http://box-b:8080"]`})
	s.Equal([]any{map[string]any{"url": "http://box-a:8080"}, map[string]any{"url": "http://box-b:8080"}}, several[0]["endpoints"])

	// Yours come after it
	withYours := backendsFor(withLlm(nil), nil)
	s.Len(withYours, 2)
	s.Equal("model-server", withYours[0]["id"])
	s.Equal("openai", withYours[1]["id"])

	// Off: yours only, or an empty list the service accepts
	off := backendsFor(withLlm(map[string]string{"modelServer.enabled": "false"}), nil)
	s.Len(off, 1)
	s.Equal("openai", off[0]["id"])
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), s.render(withToken(map[string]string{"modelServer.enabled": "false"}), "templates/deployment.yaml"), &deployment)
	s.Equal("[]\n", envOf(deployment.Spec.Template.Spec.Containers[0])["BACKENDS"].Value)

	// No model server at all is refused
	_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: withToken(map[string]string{"modelServer.replicas": "0"})}, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "/modelServer/replicas")
	s.ErrorContains(err, "minimum")
	_, err = helm.RenderTemplateE(s.T(), &helm.Options{SetValues: withToken(nil), SetJsonValues: map[string]string{"modelServer.name": `""`}}, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "/modelServer/name")
	// Past the schema too
	_, err = helm.RenderTemplateE(s.T(), &helm.Options{SetValues: withToken(map[string]string{"modelServer.replicas": "0"})}, s.chartPath, s.releaseName, []string{}, "--skip-schema-validation")
	s.ErrorContains(err, "modelServer: no model server URL")

	// The service keeps the last backend of an id, so a second model-server
	// would take over the shipped models
	duplicate := withToken(map[string]string{
		"backends[0].kind":    "model-server",
		"backends[0].id":      "model-server",
		"backends[0].baseUrl": "http://box:8080",
	})
	_, err = helm.RenderTemplateE(s.T(), &helm.Options{SetValues: duplicate}, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "backends: the id model-server is the shipped model server backend's")
	duplicate["modelServer.enabled"] = "false"
	s.Equal("model-server", backendsFor(duplicate, nil)[0]["id"])

	// An empty URL is refused by the schema
	_, err = helm.RenderTemplateE(s.T(), &helm.Options{SetValues: withToken(nil), SetJsonValues: map[string]string{"modelServer.urls": `[""]`}}, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "minLength")
}

// The model servers' optional token, from their chart's Secret
func (s *inferenceTemplateTest) TestModelServerToken() {
	deployment := s.deployment(withToken(map[string]string{"modelServer.tokenSecret": "model-server-token"}), "templates/deployment.yaml")
	env := envOf(deployment.Spec.Template.Spec.Containers[0])

	s.Equal("model-server-token", env["MODEL_SERVER_TOKEN"].ValueFrom.SecretKeyRef.Name)
	s.Equal("token", env["MODEL_SERVER_TOKEN"].ValueFrom.SecretKeyRef.Key)
	s.Equal(map[string]any{"mode": "bearer-token", "tokenEnv": "MODEL_SERVER_TOKEN"}, backendsOf(env["BACKENDS"].Value)[0]["auth"])

	// No shipped backend, no token to send
	deployment = s.deployment(withToken(map[string]string{"modelServer.tokenSecret": "model-server-token", "modelServer.enabled": "false"}), "templates/deployment.yaml")
	s.NotContains(envOf(deployment.Spec.Template.Spec.Containers[0]), "MODEL_SERVER_TOKEN")
}

// At least one license
func (s *inferenceTemplateTest) TestRefusesNoLicense() {
	_, err := helm.RenderTemplateE(s.T(), &helm.Options{SetValues: withToken(nil), SetJsonValues: map[string]string{"licenseSecrets": "[]"}}, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "/licenseSecrets")

	// Nor an unnamed one
	_, err = helm.RenderTemplateE(s.T(), &helm.Options{SetValues: withToken(nil), SetJsonValues: map[string]string{"licenseSecrets": `[""]`}}, s.chartPath, s.releaseName, []string{})
	s.ErrorContains(err, "/licenseSecrets/0")
	s.ErrorContains(err, "minLength")
}
