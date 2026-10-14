//go:build kubeall || helm || unit || unitActivityArchiveEnv
// +build kubeall helm unit unitActivityArchiveEnv

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
)

// The deployment-level Events archive configuration. Two properties are
// invisible at install time and so get template-level tests:
//
//  1. teams-api (the archiver) and the prune worker read the SAME storage
//     budget. If the prune worker's cap sat below the archiver's space
//     threshold, the prune worker would hit its cap first and only warn.
//  2. The byte count renders as an integer. Helm reads YAML numbers as
//     float64, so a naive template prints 1.073741824e+10, which the Python
//     side cannot parse.
type activityArchiveEnvTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestActivityArchiveEnvTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &activityArchiveEnvTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

func (s *activityArchiveEnvTemplateTest) deployments(
	template string,
	values map[string]string,
) map[string]appsv1.Deployment {
	options := &helm.Options{SetValues: disableTelemetry(values)}
	output := helm.RenderTemplate(
		s.T(), options, s.chartPath, s.releaseName, []string{template},
	)
	byName := map[string]appsv1.Deployment{}
	for _, doc := range strings.Split(output, "\n---") {
		if !strings.Contains(doc, "kind: Deployment") {
			continue
		}
		var d appsv1.Deployment
		helm.UnmarshalK8SYaml(s.T(), doc, &d)
		byName[d.Name] = d
	}
	return byName
}

func withActivity(extra map[string]string) map[string]string {
	values := map[string]string{}
	for k, v := range activityOn {
		values[k] = v
	}
	for k, v := range extra {
		values[k] = v
	}
	return values
}

func (s *activityArchiveEnvTemplateTest) apiEnv(values map[string]string) map[string]string {
	env := map[string]string{}
	for _, d := range s.deployments("templates/api-deployment.yaml", values) {
		for k, v := range producerEnv(d) {
			env[k] = v.Value
		}
	}
	return env
}

func (s *activityArchiveEnvTemplateTest) TestDefaultBudgetIsTenGiBAsAnInteger() {
	env := s.apiEnv(withActivity(nil))
	s.Equal("10737418240", env["FIFTYONE_ACTIVITY_MAX_STORAGE_BYTES"])
	// Nothing else renders at defaults: the API's own defaults apply.
	for _, name := range []string{
		"FIFTYONE_ACTIVITY_ARCHIVE_BUCKET_PATH",
		"FIFTYONE_ACTIVITY_ARCHIVE_RETENTION_MODE",
		"FIFTYONE_ACTIVITY_ARCHIVE_TRIGGER",
		"FIFTYONE_ACTIVITY_ARCHIVE_AFTER_DAYS",
		"FIFTYONE_ACTIVITY_ARCHIVE_SETTINGS_LOCKED",
	} {
		_, ok := env[name]
		s.False(ok, name+" must not render at defaults")
	}
}

func (s *activityArchiveEnvTemplateTest) TestConfiguredValuesReachTeamsApi() {
	env := s.apiEnv(withActivity(map[string]string{
		"activitySettings.archive.retentionMode":   "archive",
		"activitySettings.archive.bucketPath":      "gs://archive",
		"activitySettings.archive.trigger":         "space",
		"activitySettings.archive.afterDays":       "30",
		"activitySettings.archive.maxStorageBytes": "68719476736",
		"activitySettings.archive.lockSettings":    "true",
	}))
	s.Equal("archive", env["FIFTYONE_ACTIVITY_ARCHIVE_RETENTION_MODE"])
	s.Equal("gs://archive", env["FIFTYONE_ACTIVITY_ARCHIVE_BUCKET_PATH"])
	s.Equal("space", env["FIFTYONE_ACTIVITY_ARCHIVE_TRIGGER"])
	s.Equal("30", env["FIFTYONE_ACTIVITY_ARCHIVE_AFTER_DAYS"])
	s.Equal("68719476736", env["FIFTYONE_ACTIVITY_MAX_STORAGE_BYTES"])
	s.Equal("true", env["FIFTYONE_ACTIVITY_ARCHIVE_SETTINGS_LOCKED"])
}

func (s *activityArchiveEnvTemplateTest) TestPruneWorkerCarriesTheSameBudget() {
	values := withActivity(map[string]string{
		"activitySettings.archive.maxStorageBytes": "68719476736",
	})
	var prune *appsv1.Deployment
	for name, d := range s.deployments("templates/activity-workers-deployment.yaml", values) {
		if strings.Contains(name, "prune") {
			d := d
			prune = &d
		}
	}
	s.Require().NotNil(prune, "the prune worker renders by default")
	s.Equal(
		"68719476736",
		producerEnv(*prune)["FIFTYONE_ACTIVITY_MAX_STORAGE_BYTES"].Value,
	)
	s.Equal(
		s.apiEnv(values)["FIFTYONE_ACTIVITY_MAX_STORAGE_BYTES"],
		producerEnv(*prune)["FIFTYONE_ACTIVITY_MAX_STORAGE_BYTES"].Value,
	)
}

func (s *activityArchiveEnvTemplateTest) TestNothingRendersWithActivityOff() {
	env := s.apiEnv(map[string]string{
		"activitySettings.archive.bucketPath": "gs://archive",
	})
	for name := range env {
		s.False(
			strings.HasPrefix(name, "FIFTYONE_ACTIVITY_ARCHIVE") ||
				name == "FIFTYONE_ACTIVITY_MAX_STORAGE_BYTES",
			name+" must not render when Activity Core is off",
		)
	}
}
