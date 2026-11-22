//go:build kubeall || helm || unit || unitWorkflowSchedulesEnv
// +build kubeall helm unit unitWorkflowSchedulesEnv

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

// Template tests for the workflow schedule env contract.
//
// `workflowSettings.schedules.enabled` drives two variables on different
// workloads: the URI that both turns on teams-api's schedule sweep and tells
// it where the plugin server is, and the flag that makes the plugin server
// offer the *On Schedule* stage. If only one of them rendered, a schedule
// could be started that nothing fires, or fired at a server that refuses it,
// and neither shows at install time.
type workflowSchedulesEnvTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestWorkflowSchedulesEnvTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &workflowSchedulesEnvTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

const (
	workflowScheduleURIEnv     = "FIFTYONE_WORKFLOW_SCHEDULE_API_URI"
	workflowScheduleEnabledEnv = "FIFTYONE_WORKFLOW_SCHEDULE_ENABLED"
)

var workflowSchedulesOn = map[string]string{
	"workflowSettings.schedules.enabled": "true",
}

// The workloads the switch reaches, and the values each needs before its
// Deployment renders at all.
var workflowScheduleWorkloads = []struct {
	name     string
	template string
	values   map[string]string
	// The variable this workload carries when schedules are on.
	carries string
}{
	{
		"teams-api",
		"templates/api-deployment.yaml",
		map[string]string{},
		workflowScheduleURIEnv,
	},
	{
		"fiftyone-app",
		"templates/app-deployment.yaml",
		map[string]string{},
		workflowScheduleEnabledEnv,
	},
	{
		"teams-plugins",
		"templates/plugins-deployment.yaml",
		map[string]string{"pluginsSettings.enabled": "true"},
		workflowScheduleEnabledEnv,
	},
}

// workflowScheduleEnv renders one workload and returns its containers' env
// by name.
func (s *workflowSchedulesEnvTemplateTest) workflowScheduleEnv(
	template string,
	values ...map[string]string,
) map[string]corev1.EnvVar {
	merged := map[string]string{}
	for _, v := range values {
		for k, val := range v {
			merged[k] = val
		}
	}
	options := &helm.Options{SetValues: disableTelemetry(merged)}
	output := helm.RenderTemplate(
		s.T(), options, s.chartPath, s.releaseName, []string{template},
	)

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(s.T(), output, &deployment)

	byName := map[string]corev1.EnvVar{}
	for _, c := range deployment.Spec.Template.Spec.Containers {
		for _, e := range c.Env {
			byName[e.Name] = e
		}
	}
	return byName
}

// Off is the default, and off renders nothing: an empty value would still
// be a second place for the setting to live.
func (s *workflowSchedulesEnvTemplateTest) TestAbsentByDefault() {
	for _, tc := range workflowScheduleWorkloads {
		s.Run(tc.name, func() {
			env := s.workflowScheduleEnv(tc.template, tc.values)

			_, ok := env[workflowScheduleURIEnv]
			s.False(ok, tc.name+" must not carry the sweep URI by default")
			_, ok = env[workflowScheduleEnabledEnv]
			s.False(ok, tc.name+" must not carry the stage flag by default")
		})
	}
}

func (s *workflowSchedulesEnvTemplateTest) TestAbsentWhenDisabled() {
	off := map[string]string{"workflowSettings.schedules.enabled": "false"}
	for _, tc := range workflowScheduleWorkloads {
		s.Run(tc.name, func() {
			env := s.workflowScheduleEnv(tc.template, tc.values, off)

			_, ok := env[workflowScheduleURIEnv]
			s.False(ok, tc.name+" must not carry the sweep URI when disabled")
			_, ok = env[workflowScheduleEnabledEnv]
			s.False(ok, tc.name+" must not carry the stage flag when disabled")
		})
	}
}

// Each workload gets its own variable and not the other's: teams-api sweeps,
// the plugin servers offer the stage.
func (s *workflowSchedulesEnvTemplateTest) TestEachWorkloadCarriesItsVariableWhenEnabled() {
	for _, tc := range workflowScheduleWorkloads {
		s.Run(tc.name, func() {
			env := s.workflowScheduleEnv(
				tc.template, tc.values, workflowSchedulesOn,
			)

			_, hasURI := env[workflowScheduleURIEnv]
			_, hasFlag := env[workflowScheduleEnabledEnv]
			s.Equal(tc.carries == workflowScheduleURIEnv, hasURI,
				tc.name+" sweep URI")
			s.Equal(tc.carries == workflowScheduleEnabledEnv, hasFlag,
				tc.name+" stage flag")
		})
	}
}

func (s *workflowSchedulesEnvTemplateTest) TestStageFlagIsTheStringTrue() {
	for _, tc := range workflowScheduleWorkloads {
		if tc.carries != workflowScheduleEnabledEnv {
			continue
		}
		s.Run(tc.name, func() {
			env := s.workflowScheduleEnv(
				tc.template, tc.values, workflowSchedulesOn,
			)
			s.Equal("true", env[workflowScheduleEnabledEnv].Value)
		})
	}
}

// The sweep posts to whichever service runs the plugins: the dedicated
// teams-plugins service when it is enabled, fiftyone-app otherwise.
func (s *workflowSchedulesEnvTemplateTest) TestSweepURIFollowsThePluginServer() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"builtinPlugins",
			map[string]string{},
			"http://fiftyone-app:80",
		},
		{
			"dedicatedPlugins",
			map[string]string{"pluginsSettings.enabled": "true"},
			"http://teams-plugins:80",
		},
		{
			"overrideAppService",
			map[string]string{
				"appSettings.service.name": "test-fiftyone-app",
				"appSettings.service.port": "5151",
			},
			"http://test-fiftyone-app:5151",
		},
		{
			"overridePluginsService",
			map[string]string{
				"pluginsSettings.enabled":      "true",
				"pluginsSettings.service.name": "test-teams-plugins",
				"pluginsSettings.service.port": "5152",
			},
			"http://test-teams-plugins:5152",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			env := s.workflowScheduleEnv(
				"templates/api-deployment.yaml", tc.values, workflowSchedulesOn,
			)
			s.Equal(tc.expected, env[workflowScheduleURIEnv].Value)
		})
	}
}

// The sweep must reach the same server teams-app sends operator calls to, so
// both read one helper.
func (s *workflowSchedulesEnvTemplateTest) TestSweepURIMatchesTeamsAppPluginURL() {
	for _, plugins := range []string{"false", "true"} {
		s.Run("pluginsEnabled="+plugins, func() {
			values := map[string]string{"pluginsSettings.enabled": plugins}
			api := s.workflowScheduleEnv(
				"templates/api-deployment.yaml", values, workflowSchedulesOn,
			)
			teamsApp := s.workflowScheduleEnv(
				"templates/teams-app-deployment.yaml", values, workflowSchedulesOn,
			)
			s.Equal(
				teamsApp["FIFTYONE_TEAMS_PLUGIN_URL"].Value,
				api[workflowScheduleURIEnv].Value,
			)
		})
	}
}

// teams-app and teams-cas take no part in schedules.
func (s *workflowSchedulesEnvTemplateTest) TestOtherWorkloadsCarryNeither() {
	for _, template := range []string{
		"templates/teams-app-deployment.yaml",
		"templates/cas-deployment.yaml",
	} {
		s.Run(template, func() {
			env := s.workflowScheduleEnv(template, workflowSchedulesOn)

			_, ok := env[workflowScheduleURIEnv]
			s.False(ok)
			_, ok = env[workflowScheduleEnabledEnv]
			s.False(ok)
		})
	}
}
