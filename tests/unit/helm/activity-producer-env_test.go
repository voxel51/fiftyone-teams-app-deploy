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
