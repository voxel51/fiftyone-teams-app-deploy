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

// The prune worker's entrypoint first ships in fiftyone-activity >=0.0.26
// (FOEPD-4410); until the image pin is bumped, rendering its Deployment by
// default would CrashLoopBackOff on every install that enables activity.
func (s *activityWorkersDeploymentTemplateTest) TestPruneWorkerDisabledByDefault() {
	deployments := s.renderWorkers(activityEnabled(nil))
	names := []string{}
	for _, d := range deployments {
		names = append(names, d.ObjectMeta.Name)
	}
	s.Len(deployments, 3, "expected ingest, rollup, snapshot only; got %v", names)
	for _, d := range deployments {
		s.NotContains(d.ObjectMeta.Name, "prune")
	}
}

func (s *activityWorkersDeploymentTemplateTest) TestPruneWorkerOptIn() {
	deployments := s.renderWorkers(activityEnabled(map[string]string{
		"activitySettings.workers.prune.enabled": "true",
	}))
	s.Len(deployments, 4)
}

// Rollups are org-scoped: an empty org id yields a "successful" install of a
// silently empty pipeline, so the template must refuse to render without one.
func (s *activityWorkersDeploymentTemplateTest) TestOrgIdRequired() {
	values := activityEnabled(nil)
	delete(values, "activitySettings.orgId")
	options := &helm.Options{SetValues: values}
	_, err := helm.RenderTemplateE(
		s.T(), options, s.chartPath, s.releaseName, s.templates,
	)
	s.Error(err)
	s.Contains(err.Error(), "activitySettings.orgId is required")
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

		_, ok = env["FIFTYONE_MQ_REDIS_URL"]
		s.True(ok, "workers consume the fiftyone-mq queue")
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
