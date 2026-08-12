//go:build kubeall || helm || unit || unitActivityProducerEnv
// +build kubeall helm unit unitActivityProducerEnv

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

// Producer-side regression tests for the activity/queue env contract.
// The workflow producers (teams-plugins) and the delegated operators emit
// activity events; a producer without the queue URL drops every emit
// SILENTLY, and one without the org id emits events the tenant-scoped
// readers never return — both failure modes are invisible at install time,
// which is why they get template-level tests.
type activityProducerEnvTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestActivityProducerEnvTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &activityProducerEnvTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

func (s *activityProducerEnvTemplateTest) renderFirstDeployment(
	template string,
	values map[string]string,
) appsv1.Deployment {
	options := &helm.Options{SetValues: values}
	output := helm.RenderTemplate(
		s.T(), options, s.chartPath, s.releaseName, []string{template},
	)
	// Some templates are multi-document; the first Deployment is enough
	// for env assertions (all instances share the env helper).
	for _, doc := range strings.Split(output, "\n---") {
		if !strings.Contains(doc, "kind: Deployment") {
			continue
		}
		var d appsv1.Deployment
		helm.UnmarshalK8SYaml(s.T(), doc, &d)
		return d
	}
	s.FailNow("no Deployment rendered by " + template)
	return appsv1.Deployment{}
}

func producerEnv(d appsv1.Deployment) map[string]corev1.EnvVar {
	byName := map[string]corev1.EnvVar{}
	for _, c := range d.Spec.Template.Spec.Containers {
		for _, e := range c.Env {
			byName[e.Name] = e
		}
	}
	return byName
}

var activityOn = map[string]string{
	"activitySettings.enabled": "true",
	"activitySettings.orgId":   "test-org",
	"fiftyoneMq.enabled":       "true",
}

// Every workload that carries the gated activity env, and the values it
// needs before its Deployment renders at all.
var activityProducerTemplates = []struct {
	name     string
	template string
	values   map[string]string
}{
	{
		"teams-api",
		"templates/api-deployment.yaml",
		map[string]string{},
	},
	{
		"fiftyone-app",
		"templates/app-deployment.yaml",
		map[string]string{},
	},
	{
		"teams-plugins",
		"templates/plugins-deployment.yaml",
		map[string]string{"pluginsSettings.enabled": "true"},
	},
	{
		"delegated-operators",
		"templates/delegated-operator-instance-deployment.yaml",
		map[string]string{
			"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled": "true",
		},
	},
}

// FIFTYONE_ACTIVITY_ENABLED is the single producer-side gate: emit, flush,
// and the operator mutation capture all no-op without it, checked before any
// queue client is constructed. Two properties matter and neither is visible
// at install time, so both get template-level tests:
//
//  1. It reaches EVERY workload that already gets the gated activity env. A
//     producer that emits but is not told it is enabled drops silently.
//  2. It is ABSENT at default values. This is the half that regresses: the
//     var replaced "is FIFTYONE_MQ_REDIS_URL set" as the gate precisely
//     because that signal could not distinguish unset from explicitly-local,
//     so a default render that leaked the flag would recreate the bug it
//     was added to fix. Absent and "false" mean the same thing to the
//     workload, so the chart renders nothing rather than "false".
func (s *activityProducerEnvTemplateTest) TestActivityEnabledPresentWhenEnabled() {
	for _, tc := range activityProducerTemplates {
		s.Run(tc.name, func() {
			values := map[string]string{}
			for k, v := range tc.values {
				values[k] = v
			}
			for k, v := range activityOn {
				values[k] = v
			}
			d := s.renderFirstDeployment(tc.template, disableTelemetry(values))
			env := producerEnv(d)

			e, ok := env["FIFTYONE_ACTIVITY_ENABLED"]
			s.True(ok, tc.name+" producer needs the activity gate")
			s.Equal("true", e.Value, tc.name+" gate must be the string \"true\"")
		})
	}
}

func (s *activityProducerEnvTemplateTest) TestActivityEnabledAbsentWhenDisabled() {
	for _, tc := range activityProducerTemplates {
		s.Run(tc.name, func() {
			// Chart defaults for activitySettings/fiftyoneMq — only the
			// values that make the Deployment render are set.
			values := map[string]string{}
			for k, v := range tc.values {
				values[k] = v
			}
			d := s.renderFirstDeployment(tc.template, disableTelemetry(values))
			env := producerEnv(d)

			_, ok := env["FIFTYONE_ACTIVITY_ENABLED"]
			s.False(ok, tc.name+" must not carry the activity gate at defaults")
		})
	}
}

// Explicitly disabling activity while the queue Redis is enabled must still
// leave the gate off. This is the disagreement the flag exists to prevent:
// the queue being reachable is not consent to emit.
func (s *activityProducerEnvTemplateTest) TestActivityEnabledOffWithQueueOn() {
	values := map[string]string{
		"activitySettings.enabled": "false",
		"activitySettings.orgId":   "test-org",
		"fiftyoneMq.enabled":       "true",
	}
	d := s.renderFirstDeployment(
		"templates/api-deployment.yaml", disableTelemetry(values),
	)
	env := producerEnv(d)

	_, ok := env["FIFTYONE_MQ_REDIS_URL"]
	s.True(ok, "queue URL still renders from fiftyoneMq.enabled")
	_, ok = env["FIFTYONE_ACTIVITY_ENABLED"]
	s.False(ok, "a reachable queue must not imply the activity gate")
	_, ok = env["FIFTYONE_ACTIVITY_ORG_ID"]
	s.False(ok, "org id stays gated on activitySettings.enabled")
}

// teams-app is the frontend, not a producer: it must never be handed the
// gate, otherwise the env list becomes a fourth signal that can disagree.
func (s *activityProducerEnvTemplateTest) TestTeamsAppNeverCarriesActivityGate() {
	values := map[string]string{}
	for k, v := range activityOn {
		values[k] = v
	}
	d := s.renderFirstDeployment(
		"templates/teams-app-deployment.yaml", disableTelemetry(values),
	)
	_, ok := producerEnv(d)["FIFTYONE_ACTIVITY_ENABLED"]
	s.False(ok, "teams-app is not an activity producer")
}

// The worker Deployments only render when activity is enabled, but they emit
// derived rollup/snapshot events through the same gated path, so they need
// the flag too.
func (s *activityProducerEnvTemplateTest) TestActivityWorkersCarryTheGate() {
	values := map[string]string{}
	for k, v := range activityOn {
		values[k] = v
	}
	d := s.renderFirstDeployment(
		"templates/activity-workers-deployment.yaml", disableTelemetry(values),
	)
	s.Equal("true", producerEnv(d)["FIFTYONE_ACTIVITY_ENABLED"].Value)
}

// teams-plugins runs the workflow producers in dedicated-plugins
// deployments — it must carry the queue URL and org id.
func (s *activityProducerEnvTemplateTest) TestPluginsProducerEnv() {
	values := map[string]string{"pluginsSettings.enabled": "true"}
	for k, v := range activityOn {
		values[k] = v
	}
	d := s.renderFirstDeployment(
		"templates/plugins-deployment.yaml", disableTelemetry(values),
	)
	env := producerEnv(d)

	_, ok := env["FIFTYONE_MQ_REDIS_URL"]
	s.True(ok, "plugins producer needs the queue URL")
	s.Equal("test-org", env["FIFTYONE_ACTIVITY_ORG_ID"].Value)

	// Co-located default: the readers fall back to FIFTYONE_DATABASE_NAME
	// on their own; pinning FIFTYONE_ACTIVITY_MONGO_DB here would defeat
	// that fallback and can diverge from what the workers resolve.
	_, ok = env["FIFTYONE_ACTIVITY_MONGO_DB"]
	s.False(ok, "no dedicated DB configured -> no override env")
}

func (s *activityProducerEnvTemplateTest) TestPluginsProducerDedicatedDbEnv() {
	values := map[string]string{
		"pluginsSettings.enabled":         "true",
		"activitySettings.mongo.database": "dedicated_activity",
	}
	for k, v := range activityOn {
		values[k] = v
	}
	d := s.renderFirstDeployment(
		"templates/plugins-deployment.yaml", disableTelemetry(values),
	)
	env := producerEnv(d)
	s.Equal("dedicated_activity", env["FIFTYONE_ACTIVITY_MONGO_DB"].Value)
}

// Regression: the delegated-operator instance deployments' queue env was
// nested inside the telemetry toggle, so telemetry-off deployments dropped
// every DO emit. The queue/activity env must render with telemetry OFF.
func (s *activityProducerEnvTemplateTest) TestDelegatedOperatorEnvWithTelemetryDisabled() {
	values := map[string]string{
		"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled": "true",
	}
	for k, v := range activityOn {
		values[k] = v
	}
	d := s.renderFirstDeployment(
		"templates/delegated-operator-instance-deployment.yaml",
		disableTelemetry(values),
	)
	env := producerEnv(d)

	_, ok := env["FIFTYONE_MQ_REDIS_URL"]
	s.True(ok, "DO queue URL must not depend on the telemetry toggle")
	s.Equal("test-org", env["FIFTYONE_ACTIVITY_ORG_ID"].Value)
}
