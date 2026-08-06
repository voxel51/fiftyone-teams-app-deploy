//go:build kubeall || helm || unit || unitFiftyoneMqRedisDeployment
// +build kubeall helm unit unitFiftyoneMqRedisDeployment

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

	appsv1 "k8s.io/api/apps/v1"
)

type fiftyoneMqRedisDeploymentTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestFiftyoneMqRedisDeploymentTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &fiftyoneMqRedisDeploymentTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates: []string{
			"templates/fiftyone-mq-redis-deployment.yaml",
		},
	})
}

// mqRedisEnabled returns SetValues with the two gates the bundled fiftyone-mq
// Redis requires. Tests set these explicitly rather than leaning on the chart
// defaults so they keep asserting the same thing if the shipped default for
// `fiftyoneMq.enabled` changes.
func mqRedisEnabled(values map[string]string) map[string]string {
	out := map[string]string{
		"fiftyoneMq.enabled":       "true",
		"fiftyoneMq.redis.enabled": "true",
	}
	for k, v := range values {
		out[k] = v
	}
	return out
}

// mqRedisArgValue returns the value following the given redis-server flag in
// the container args list. redis-server takes its config as `--flag value`
// pairs, so asserting on the element after the flag is what actually pins the
// setting; a bare Contains() would pass even if the value landed under the
// wrong flag.
func mqRedisArgValue(args []string, flag string) (string, bool) {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1], true
		}
	}
	return "", false
}

// renderDeployment renders the Deployment template and unmarshals it into the
// typed apps/v1 struct.
func (s *fiftyoneMqRedisDeploymentTemplateTest) renderDeployment(options *helm.Options) appsv1.Deployment {
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)
	return deployment
}

func (s *fiftyoneMqRedisDeploymentTemplateTest) TestMetadata() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	expectedName := fmt.Sprintf("%s-fiftyone-mq-redis", s.releaseName)
	s.Equal(expectedName, deployment.ObjectMeta.Name, "Deployment name should be release-prefixed")
	s.True(strings.HasPrefix(deployment.ObjectMeta.Name, s.releaseName+"-"),
		"Deployment name must carry the release-name prefix so two releases can share a namespace")
	s.Equal("fiftyone-teams", deployment.ObjectMeta.Namespace,
		"Deployment namespace should default to fiftyone-teams")
	s.Equal("fiftyone-mq-redis", deployment.ObjectMeta.Labels["app.kubernetes.io/name"])
	s.Equal(s.releaseName, deployment.ObjectMeta.Labels["app.kubernetes.io/instance"])
	s.Equal("fiftyone-mq-redis", deployment.ObjectMeta.Labels["app.voxel51.com/component"])
}

func (s *fiftyoneMqRedisDeploymentTemplateTest) TestNamespaceOverride() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"namespace.name": "my-ns",
	})}

	deployment := s.renderDeployment(options)
	s.Equal("my-ns", deployment.ObjectMeta.Namespace)
}

// TestMqDisabled ensures the bundled Redis is not rendered when fiftyone-mq is
// off, even with the redis sub-toggle on.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestMqDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"fiftyoneMq.enabled":       "false",
		"fiftyoneMq.redis.enabled": "true",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

// TestRedisDisabled ensures the redis sub-toggle alone suppresses the bundled
// Deployment while fiftyone-mq stays enabled.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestRedisDisabled() {
	options := &helm.Options{SetValues: map[string]string{
		"fiftyoneMq.enabled":       "true",
		"fiftyoneMq.redis.enabled": "false",
	}}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

// TestExternalUrlSkipsBundled ensures that pointing fiftyone-mq at an
// operator-managed Redis suppresses the bundled Deployment. Rendering both
// would leave an unused Redis pod running while producers and workers talk to
// the external instance.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestExternalUrlSkipsBundled() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.external.url": "redis://my-managed-redis:6379/0",
	})}

	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, s.templates)
	s.ErrorContains(err, "could not find template")
}

// TestReplicasAndStrategy pins the single-replica Recreate rollout. The queue
// is a single non-clustered Redis, so a RollingUpdate would briefly run two
// pods backed by different in-memory state.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestReplicasAndStrategy() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	s.Require().NotNil(deployment.Spec.Replicas, "replicas should be set explicitly")
	s.EqualValues(1, *deployment.Spec.Replicas, "bundled queue Redis should run exactly one replica")
	s.Equal(appsv1.RecreateDeploymentStrategyType, deployment.Spec.Strategy.Type)
}

func (s *fiftyoneMqRedisDeploymentTemplateTest) TestSelectorMatchesPodLabels() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	s.Require().NotNil(deployment.Spec.Selector)
	s.Equal("fiftyone-mq-redis", deployment.Spec.Selector.MatchLabels["app.kubernetes.io/name"])
	s.Equal(s.releaseName, deployment.Spec.Selector.MatchLabels["app.kubernetes.io/instance"])

	// Every selector label must also be on the pod template, or the Deployment
	// is rejected by the API server.
	podLabels := deployment.Spec.Template.ObjectMeta.Labels
	for k, v := range deployment.Spec.Selector.MatchLabels {
		s.Equal(v, podLabels[k], "pod template must carry selector label %s", k)
	}
}

func (s *fiftyoneMqRedisDeploymentTemplateTest) TestContainer() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	containers := deployment.Spec.Template.Spec.Containers
	s.Require().Len(containers, 1, "Deployment should have exactly one container")
	s.Equal("redis", containers[0].Name)
	s.Equal("redis:7-alpine", containers[0].Image)

	s.Require().Len(containers[0].Ports, 1, "container should expose exactly one port")
	s.EqualValues(6379, containers[0].Ports[0].ContainerPort)
}

func (s *fiftyoneMqRedisDeploymentTemplateTest) TestImageOverride() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.image": "my-registry/redis:custom",
	})}

	deployment := s.renderDeployment(options)
	s.Equal("my-registry/redis:custom", deployment.Spec.Template.Spec.Containers[0].Image)
}

func (s *fiftyoneMqRedisDeploymentTemplateTest) TestMaxmemoryDefault() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	s.Contains(args, "redis-server")
	value, ok := mqRedisArgValue(args, "--maxmemory")
	s.Require().True(ok, "--maxmemory should be passed to redis-server, got args %v", args)
	s.Equal("200mb", value, "--maxmemory should default to 200mb")
}

func (s *fiftyoneMqRedisDeploymentTemplateTest) TestMaxmemoryOverride() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.maxmemory": "1gb",
	})}

	deployment := s.renderDeployment(options)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	value, ok := mqRedisArgValue(args, "--maxmemory")
	s.Require().True(ok, "--maxmemory should be passed to redis-server, got args %v", args)
	s.Equal("1gb", value, "--maxmemory should honor fiftyoneMq.redis.maxmemory")
}

// TestMaxmemoryPolicyIsNoeviction is the load-bearing assertion for this
// Deployment. BullMQ stores each queued job as its own Redis key; under any
// evicting `maxmemory-policy` (allkeys-lru, volatile-*, ...) Redis silently
// drops those keys when it hits the memory ceiling, so jobs vanish with no
// error surfaced to producers or workers. The policy must therefore stay
// `noeviction` — Redis then rejects writes once full, which surfaces as a
// visible enqueue error instead of silent data loss. It is deliberately NOT
// exposed as a value.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestMaxmemoryPolicyIsNoeviction() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	policy, ok := mqRedisArgValue(args, "--maxmemory-policy")
	s.Require().True(ok, "--maxmemory-policy should be passed to redis-server, got args %v", args)
	s.Equal("noeviction", policy,
		"maxmemory-policy must be noeviction; an evicting policy silently discards queued BullMQ jobs")
}

// TestMaxmemoryPolicyNotOverridable guards the invariant above against a
// values key being wired up later: even if someone sets a policy value, the
// rendered args must still say noeviction.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestMaxmemoryPolicyNotOverridable() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.maxmemoryPolicy": "allkeys-lru",
	})}

	deployment := s.renderDeployment(options)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	policy, ok := mqRedisArgValue(args, "--maxmemory-policy")
	s.Require().True(ok, "--maxmemory-policy should be passed to redis-server, got args %v", args)
	s.Equal("noeviction", policy,
		"maxmemory-policy must not be overridable to an evicting policy")
	s.NotContains(args, "allkeys-lru")
}
