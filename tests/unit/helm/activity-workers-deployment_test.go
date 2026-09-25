//go:build kubeall || helm || unit || unitActivityWorkersDeployment
// +build kubeall helm unit unitActivityWorkersDeployment

package unit

import (
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

type activityWorkersDeploymentTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestActivityWorkersDeploymentTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &activityWorkersDeploymentTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates: []string{
			"templates/activity-workers-deployment.yaml",
		},
	})
}

// TestActivityWithoutQueueFailsRender pins the pairing guard. Activity on with
// the queue off renders producers and workers with nothing to connect to, so
// every event is dropped while the install reports healthy -- the failure this
// turns into a render error instead.
func (s *activityWorkersDeploymentTemplateTest) TestActivityWithoutQueueFailsRender() {
	options := &helm.Options{SetValues: map[string]string{
		"activitySettings.enabled": "true",
		"fiftyoneMq.enabled":       "false",
	}}

	_, err := helm.RenderTemplateE(
		s.T(), options, s.chartPath, s.releaseName,
		[]string{"templates/api-deployment.yaml"},
	)

	s.Require().Error(err, "activity without a queue must not render")
	s.Contains(
		err.Error(),
		"activitySettings.enabled is true but fiftyoneMq.enabled is false",
		"the error should name both settings so the fix is obvious",
	)
}

// TestQueueWithoutActivityRenders pins the deliberate asymmetry: a reachable
// queue is not consent to emit, which is why the gate is a flag rather than an
// inference from FIFTYONE_MQ_REDIS_URL. This direction must stay legal.
func (s *activityWorkersDeploymentTemplateTest) TestQueueWithoutActivityRenders() {
	options := &helm.Options{SetValues: map[string]string{
		"activitySettings.enabled": "false",
		"fiftyoneMq.enabled":       "true",
	}}

	_, err := helm.RenderTemplateE(
		s.T(), options, s.chartPath, s.releaseName,
		[]string{"templates/api-deployment.yaml"},
	)

	s.NoError(err, "queue on with activity off is a supported configuration")
}

// activityEnabled returns SetValues enabling the workers with the values the
// template requires. Tests set these explicitly rather than leaning on chart
// defaults so they keep asserting the same thing if defaults change.
func activityEnabled(values map[string]string) map[string]string {
	out := map[string]string{
		"activitySettings.enabled": "true",
		"activitySettings.orgId":   "test-org",
		"fiftyoneMq.enabled":       "true",
	}
	for k, v := range values {
		out[k] = v
	}
	return out
}

// renderWorkers renders the multi-document workers template into one
// Deployment per document.
func (s *activityWorkersDeploymentTemplateTest) renderWorkers(
	values map[string]string,
) []appsv1.Deployment {
	options := &helm.Options{SetValues: values}
	output := helm.RenderTemplate(
		s.T(), options, s.chartPath, s.releaseName, s.templates,
	)
	docs := strings.Split(output, "\n---")
	deployments := []appsv1.Deployment{}
	for _, doc := range docs {
		if !strings.Contains(doc, "kind: Deployment") {
			continue
		}
		var d appsv1.Deployment
		helm.UnmarshalK8SYaml(s.T(), doc, &d)
		deployments = append(deployments, d)
	}
	return deployments
}

func envByName(container corev1.Container) map[string]corev1.EnvVar {
	byName := map[string]corev1.EnvVar{}
	for _, e := range container.Env {
		byName[e.Name] = e
	}
	return byName
}

// The prune worker enforces the storage size cap. Without it only the
// time-based TTL bounds the event store, so it renders by default now that
// the image pin carries its entrypoint.
func (s *activityWorkersDeploymentTemplateTest) TestPruneWorkerEnabledByDefault() {
	deployments := s.renderWorkers(activityEnabled(nil))
	names := []string{}
	for _, d := range deployments {
		names = append(names, d.ObjectMeta.Name)
	}
	s.Len(deployments, 3, "expected ingest, prune, rollup; got %v", names)
	s.Contains(strings.Join(names, ","), "prune")
}

func (s *activityWorkersDeploymentTemplateTest) TestPruneWorkerOptOut() {
	deployments := s.renderWorkers(activityEnabled(map[string]string{
		"activitySettings.workers.prune.enabled": "false",
	}))
	names := []string{}
	for _, d := range deployments {
		names = append(names, d.ObjectMeta.Name)
	}
	s.Len(deployments, 2, "expected ingest, rollup only; got %v", names)
	for _, d := range deployments {
		s.NotContains(d.ObjectMeta.Name, "prune")
	}
}

// An unset org id renders, and renders no org var at all. It is an override,
// not a requirement: the producers stamp the authenticated organization from
// the request, and teams-api reads the organization for the dataset and
// sample counts from CAS. The var must be absent rather than an empty string,
// which the workers cannot tell apart from a deliberate blank.
func (s *activityWorkersDeploymentTemplateTest) TestOrgIdOmittedWhenUnset() {
	values := activityEnabled(nil)
	delete(values, "activitySettings.orgId")
	options := &helm.Options{SetValues: values}
	_, err := helm.RenderTemplateE(
		s.T(), options, s.chartPath, s.releaseName, s.templates,
	)
	s.NoError(err, "an unset orgId must not fail the render")

	deployments := s.renderWorkers(values)
	s.Len(deployments, 3)
	for _, d := range deployments {
		env := envByName(d.Spec.Template.Spec.Containers[0])
		_, ok := env["FIFTYONE_ACTIVITY_ORG_ID"]
		s.False(
			ok,
			"%s must not carry FIFTYONE_ACTIVITY_ORG_ID when orgId is unset",
			d.ObjectMeta.Name,
		)
	}
}

// The other half of the contract: when set, every worker carries it verbatim.
func (s *activityWorkersDeploymentTemplateTest) TestOrgIdRenderedWhenSet() {
	deployments := s.renderWorkers(activityEnabled(map[string]string{
		"activitySettings.orgId": "acme-org",
	}))
	s.Len(deployments, 3)
	for _, d := range deployments {
		env := envByName(d.Spec.Template.Spec.Containers[0])
		orgId, ok := env["FIFTYONE_ACTIVITY_ORG_ID"]
		s.True(ok, "%s must carry FIFTYONE_ACTIVITY_ORG_ID", d.ObjectMeta.Name)
		s.Equal("acme-org", orgId.Value)
		s.Nil(orgId.ValueFrom, "org id is a literal value, not a secret ref")
	}
}

// Workers have no FIFTYONE_DATABASE_* fallback of their own, so the chart
// must always resolve their Mongo database: the dedicated override when set,
// else the per-deployment FiftyOne database from the shared secret.
func (s *activityWorkersDeploymentTemplateTest) TestWorkerEnvDefaults() {
	deployments := s.renderWorkers(activityEnabled(nil))
	for _, d := range deployments {
		env := envByName(d.Spec.Template.Spec.Containers[0])

		s.Equal("test-org", env["FIFTYONE_ACTIVITY_ORG_ID"].Value)

		mongoDb, ok := env["FIFTYONE_ACTIVITY_MONGO_DB"]
		s.True(ok)
		s.NotNil(mongoDb.ValueFrom, "co-located default reads the secret")
		s.Equal(
			"fiftyoneDatabaseName",
			mongoDb.ValueFrom.SecretKeyRef.Key,
		)

		requireMqRedisURL(s.T(), env, s.releaseName,
			"workers consume the fiftyone-mq queue")
	}
}

func (s *activityWorkersDeploymentTemplateTest) TestWorkerEnvDedicatedDatabase() {
	deployments := s.renderWorkers(activityEnabled(map[string]string{
		"activitySettings.mongo.database": "dedicated_activity",
	}))
	for _, d := range deployments {
		env := envByName(d.Spec.Template.Spec.Containers[0])
		s.Equal("dedicated_activity", env["FIFTYONE_ACTIVITY_MONGO_DB"].Value)
		s.Nil(env["FIFTYONE_ACTIVITY_MONGO_DB"].ValueFrom)
	}
}

// A worker left parked on a dead Redis connection (the queue pod was
// replaced) stays Running and Ready while draining nothing -- the
// 2026-09-28 ephem incident. Every worker therefore gets the heartbeat file
// env and an exec liveness probe on its age by default.
func (s *activityWorkersDeploymentTemplateTest) TestWorkerLivenessDefaults() {
	deployments := s.renderWorkers(activityEnabled(nil))
	s.Len(deployments, 4)
	for _, d := range deployments {
		container := d.Spec.Template.Spec.Containers[0]
		env := envByName(container)
		s.Equal(
			"/tmp/fiftyone-mq-heartbeat",
			env["FIFTYONE_MQ_HEARTBEAT_FILE"].Value,
			"%s must tell the worker where to write its heartbeat",
			d.ObjectMeta.Name,
		)

		probe := container.LivenessProbe
		s.Require().NotNil(probe, "%s must have a liveness probe", d.ObjectMeta.Name)
		s.Require().NotNil(probe.Exec)
		s.Equal([]string{"sh", "-c"}, probe.Exec.Command[:2])
		script := probe.Exec.Command[2]
		s.Contains(script, `[ ! -f "$f" ]`,
			"a missing heartbeat file must pass, so older images are not restarted")
		s.Contains(script, "-lt 180 ]")
		s.Equal(int32(3), probe.FailureThreshold)
		s.Equal(int32(30), probe.InitialDelaySeconds)
		s.Equal(int32(30), probe.PeriodSeconds)
		s.Equal(int32(5), probe.TimeoutSeconds)
	}
}

func (s *activityWorkersDeploymentTemplateTest) TestWorkerLivenessOverrides() {
	deployments := s.renderWorkers(activityEnabled(map[string]string{
		"activitySettings.liveness.heartbeatFile":          "/var/run/hb",
		"activitySettings.liveness.maxHeartbeatAgeSeconds": "600",
		"activitySettings.liveness.periodSeconds":          "10",
	}))
	for _, d := range deployments {
		container := d.Spec.Template.Spec.Containers[0]
		s.Equal("/var/run/hb", envByName(container)["FIFTYONE_MQ_HEARTBEAT_FILE"].Value)
		s.Contains(container.LivenessProbe.Exec.Command[2], "-lt 600 ]")
		s.Equal(int32(10), container.LivenessProbe.PeriodSeconds)
	}
}

func (s *activityWorkersDeploymentTemplateTest) TestWorkerLivenessOptOut() {
	deployments := s.renderWorkers(activityEnabled(map[string]string{
		"activitySettings.liveness.enabled": "false",
	}))
	for _, d := range deployments {
		container := d.Spec.Template.Spec.Containers[0]
		s.Nil(container.LivenessProbe, "%s", d.ObjectMeta.Name)
		_, ok := envByName(container)["FIFTYONE_MQ_HEARTBEAT_FILE"]
		s.False(ok, "%s must not set the heartbeat file when disabled", d.ObjectMeta.Name)
	}
}
