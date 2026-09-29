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
	corev1 "k8s.io/api/core/v1"
)

type fiftyoneMqRedisDeploymentTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
	pvcTemplate string
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
		pvcTemplate: "templates/fiftyone-mq-redis-pvc.yaml",
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

// renderPvc renders the PVC template and unmarshals it into the typed struct.
func (s *fiftyoneMqRedisDeploymentTemplateTest) renderPvc(options *helm.Options) corev1.PersistentVolumeClaim {
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{s.pvcTemplate})
	var pvc corev1.PersistentVolumeClaim
	helm.UnmarshalK8SYaml(s.T(), output, &pvc)
	return pvc
}

// assertPvcNotRendered confirms the PVC template produces no output for the
// given values. Helm errors with "could not find template" when --show-only
// targets a template that renders empty.
func (s *fiftyoneMqRedisDeploymentTemplateTest) assertPvcNotRendered(options *helm.Options, msg string) {
	_, err := helm.RenderTemplateE(s.T(), options, s.chartPath, s.releaseName, []string{s.pvcTemplate})
	s.ErrorContains(err, "could not find template", msg)
}

// TestAofEnabled is the other load-bearing assertion for this Deployment.
// BullMQ keeps queued jobs as ordinary Redis keys, so a memory-only Redis comes
// back empty after a restart and every unconsumed job is gone. `appendonly yes`
// makes Redis replay its append-only file on startup instead.
//
// This does not make the pipeline durable end to end (FOEPD-4411) — it bounds
// the restart window only.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestAofEnabled() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	appendonly, ok := mqRedisArgValue(args, "--appendonly")
	s.Require().True(ok, "--appendonly should be passed to redis-server, got args %v", args)
	s.Equal("yes", appendonly,
		"appendonly must be yes; a memory-only queue Redis loses every queued job on restart")
}

// TestAppendfsyncIsEverysec pins the flush cadence. The Redis default for
// appendfsync is already everysec, but leaving it implicit means an image or
// config default change could silently widen the loss window (`no` defers
// flushing to the OS, up to ~30s of writes).
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestAppendfsyncIsEverysec() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	deployment := s.renderDeployment(options)

	args := deployment.Spec.Template.Spec.Containers[0].Args
	appendfsync, ok := mqRedisArgValue(args, "--appendfsync")
	s.Require().True(ok, "--appendfsync should be passed to redis-server, got args %v", args)
	s.Equal("everysec", appendfsync,
		"appendfsync should be everysec, bounding loss to ~1s of writes")
}

// TestAofDirIsMountedPath guards the pairing between --dir and the volume
// mount. If they drift apart the AOF lands on the container filesystem, where
// it is lost even on a container restart — with no error to indicate that
// persistence is not working.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestAofDirIsMountedPath() {
	for _, persistence := range []string{"false", "true"} {
		options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
			"fiftyoneMq.redis.persistence.enabled": persistence,
		})}

		deployment := s.renderDeployment(options)
		container := deployment.Spec.Template.Spec.Containers[0]

		dir, ok := mqRedisArgValue(container.Args, "--dir")
		s.Require().True(ok, "--dir should be passed to redis-server, got args %v", container.Args)

		s.Require().Len(container.VolumeMounts, 1,
			"expected exactly one volumeMount with persistence.enabled=%s", persistence)
		s.Equal("redis-data", container.VolumeMounts[0].Name)
		s.Equal(dir, container.VolumeMounts[0].MountPath,
			"--dir must match the redis-data mountPath with persistence.enabled=%s", persistence)
	}
}

// TestPersistenceEnabledByDefault pins the shipped default: queued activity
// must survive a pod reschedule, so a plain install provisions the PVC.
// Disabling it is the explicit opt-out, covered below.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestPersistenceEnabledByDefault() {
	options := &helm.Options{SetValues: mqRedisEnabled(nil)}

	pvc := s.renderPvc(options)
	s.Equal(
		fmt.Sprintf("%s-fiftyone-mq-redis-data", s.releaseName),
		pvc.ObjectMeta.Name,
		"PVC should be rendered by default",
	)

	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1, "Deployment should have exactly one volume")
	s.Equal("redis-data", volumes[0].Name)
	s.Require().NotNil(volumes[0].PersistentVolumeClaim,
		"redis-data should be PVC-backed by default so the AOF survives a pod reschedule")
	s.Nil(volumes[0].EmptyDir)
}

// TestPersistenceOptOutUsesEmptyDir covers the explicit disable: no PVC, so
// the chart still installs on a cluster with no default StorageClass. The AOF
// then survives a container restart but not a pod reschedule.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestPersistenceOptOutUsesEmptyDir() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.persistence.enabled": "false",
	})}

	s.assertPvcNotRendered(options, "PVC should not be rendered when persistence is disabled")

	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1, "Deployment should have exactly one volume")
	s.Equal("redis-data", volumes[0].Name)
	s.NotNil(volumes[0].EmptyDir,
		"redis-data should fall back to an emptyDir when persistence is disabled")
	s.Nil(volumes[0].PersistentVolumeClaim,
		"redis-data must not reference a PVC when persistence is disabled")
}

// TestPersistenceEnabledRendersPvcAndBindsIt covers the opt-in path where the
// chart provisions the claim itself.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestPersistenceEnabledRendersPvcAndBindsIt() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.persistence.enabled":      "true",
		"fiftyoneMq.redis.persistence.size":         "5Gi",
		"fiftyoneMq.redis.persistence.storageClass": "gp3",
	})}

	pvc := s.renderPvc(options)

	expectedName := fmt.Sprintf("%s-fiftyone-mq-redis-data", s.releaseName)
	s.Equal(expectedName, pvc.ObjectMeta.Name, "PVC name should be release-prefixed")
	s.Equal("fiftyone-teams", pvc.ObjectMeta.Namespace)
	s.Equal("fiftyone-mq-redis", pvc.ObjectMeta.Labels["app.voxel51.com/component"])
	s.Require().Len(pvc.Spec.AccessModes, 1)
	s.Equal(corev1.ReadWriteOnce, pvc.Spec.AccessModes[0])
	storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	s.Equal("5Gi", storage.String())
	s.Require().NotNil(pvc.Spec.StorageClassName, "storageClassName should be set")
	s.Equal("gp3", *pvc.Spec.StorageClassName)

	// The Deployment must bind the claim the chart just created.
	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1)
	s.Require().NotNil(volumes[0].PersistentVolumeClaim,
		"redis-data should reference a PVC when persistence is enabled")
	s.Equal(expectedName, volumes[0].PersistentVolumeClaim.ClaimName,
		"the Deployment must bind the chart-generated claim name")
	s.Nil(volumes[0].EmptyDir)
}

// TestPersistenceDefaultsSizeAndStorageClass leaves size/storageClass unset so
// the claim takes the chart default size and the cluster's default
// StorageClass.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestPersistenceDefaultsSizeAndStorageClass() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.persistence.enabled": "true",
	})}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, []string{s.pvcTemplate})
	s.NotContains(output, "storageClassName",
		"storageClassName should be omitted entirely so the cluster default applies")

	pvc := s.renderPvc(options)
	storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	s.Equal("1Gi", storage.String(), "size should default to 1Gi")
}

// TestExistingClaimSkipsPvcAndBindsNamedClaim covers clusters with no dynamic
// provisioner, where an operator pre-creates the claim.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestExistingClaimSkipsPvcAndBindsNamedClaim() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.persistence.enabled":       "true",
		"fiftyoneMq.redis.persistence.existingClaim": "my-prebuilt-queue-pvc",
	})}

	s.assertPvcNotRendered(options, "chart must not create a PVC when existingClaim is set")

	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1)
	s.Require().NotNil(volumes[0].PersistentVolumeClaim)
	s.Equal("my-prebuilt-queue-pvc", volumes[0].PersistentVolumeClaim.ClaimName,
		"claimName should be the user-supplied existingClaim, not the chart-generated name")
	s.Nil(volumes[0].EmptyDir)
}

// TestPersistenceDisabledBeatsExistingClaim ensures persistence.enabled stays
// the kill-switch: opting out yields an emptyDir rather than a stale claim
// mount left behind in values.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestPersistenceDisabledBeatsExistingClaim() {
	options := &helm.Options{SetValues: mqRedisEnabled(map[string]string{
		"fiftyoneMq.redis.persistence.enabled":       "false",
		"fiftyoneMq.redis.persistence.existingClaim": "my-prebuilt-queue-pvc",
	})}

	s.assertPvcNotRendered(options, "PVC should not render when persistence.enabled=false")

	deployment := s.renderDeployment(options)
	volumes := deployment.Spec.Template.Spec.Volumes
	s.Require().Len(volumes, 1)
	s.NotNil(volumes[0].EmptyDir,
		"persistence.enabled=false should win and yield an emptyDir")
	s.Nil(volumes[0].PersistentVolumeClaim,
		"redis-data must not reference the existingClaim when persistence is disabled")
}

// TestPvcFollowsBundledRedisGates ensures the PVC is suppressed by every gate
// that suppresses the bundled Deployment. A PVC rendered without its Redis
// would bind storage for a pod that does not exist.
func (s *fiftyoneMqRedisDeploymentTemplateTest) TestPvcFollowsBundledRedisGates() {
	testCases := []struct {
		name   string
		values map[string]string
	}{
		{
			"mqDisabled",
			map[string]string{
				"fiftyoneMq.enabled":                   "false",
				"fiftyoneMq.redis.enabled":             "true",
				"fiftyoneMq.redis.persistence.enabled": "true",
			},
		},
		{
			"redisDisabled",
			map[string]string{
				"fiftyoneMq.enabled":                   "true",
				"fiftyoneMq.redis.enabled":             "false",
				"fiftyoneMq.redis.persistence.enabled": "true",
			},
		},
		{
			"externalUrlSet",
			map[string]string{
				"fiftyoneMq.enabled":                   "true",
				"fiftyoneMq.redis.enabled":             "true",
				"fiftyoneMq.redis.persistence.enabled": "true",
				"fiftyoneMq.redis.external.url":        "redis://my-managed-redis:6379/0",
			},
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			options := &helm.Options{SetValues: testCase.values}
			s.assertPvcNotRendered(options,
				"PVC must not render when the bundled Redis is suppressed")
		})
	}
}
