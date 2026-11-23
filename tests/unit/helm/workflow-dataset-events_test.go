//go:build kubeall || helm || unit || unitWorkflowDatasetEvents
// +build kubeall helm unit unitWorkflowDatasetEvents

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

// Template tests for the workflow dataset-event contract.
//
// `workflowSettings.datasetEvents.enabled` reaches five workloads: the two
// Activity Core workers that publish sample changes, the trigger worker that
// posts them, teams-api which relays them, and the plugin servers which offer
// the stages. If one of them did not render, a workflow could be started
// that never receives anything, and nothing shows at install time.
type workflowDatasetEventsTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestWorkflowDatasetEventsTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &workflowDatasetEventsTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

const (
	workflowTopicEnv         = "FIFTYONE_ACTIVITY_WORKFLOWS_TOPIC_ENABLED"
	workflowTriggerURIEnv    = "FIFTYONE_ACTIVITY_TRIGGER_API_URI"
	workflowTriggerKeyEnv    = "FIFTYONE_ACTIVITY_TRIGGER_API_KEY"
	workflowRelayURIEnv      = "FIFTYONE_WORKFLOW_TRIGGER_API_URI"
	workflowDatasetEventsEnv = "FIFTYONE_WORKFLOW_DATASET_EVENTS_ENABLED"
	triggerWorkerCommand     = "fiftyone-activity-trigger-worker"
	workersTemplate          = "templates/activity-workers-deployment.yaml"
)

var workflowDatasetEventsOn = map[string]string{
	"workflowSettings.datasetEvents.enabled": "true",
}

func mergeValues(values ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, v := range values {
		for k, val := range v {
			merged[k] = val
		}
	}
	return merged
}

func containerEnv(container corev1.Container) map[string]corev1.EnvVar {
	byName := map[string]corev1.EnvVar{}
	for _, e := range container.Env {
		byName[e.Name] = e
	}
	return byName
}

// workers renders the Activity Core workers and returns each one's container
// by its entrypoint.
func (s *workflowDatasetEventsTemplateTest) workers(
	values ...map[string]string,
) map[string]corev1.Container {
	options := &helm.Options{SetValues: mergeValues(values...)}
	output := helm.RenderTemplate(
		s.T(), options, s.chartPath, s.releaseName, []string{workersTemplate},
	)

	byCommand := map[string]corev1.Container{}
	for _, doc := range strings.Split(output, "\n---") {
		if !strings.Contains(doc, "kind: Deployment") {
			continue
		}
		var d appsv1.Deployment
		helm.UnmarshalK8SYaml(s.T(), doc, &d)
		container := d.Spec.Template.Spec.Containers[0]
		byCommand[container.Command[0]] = container
	}
	return byCommand
}

// workloadEnv renders one single-Deployment template and returns its
// containers' env by name.
func (s *workflowDatasetEventsTemplateTest) workloadEnv(
	template string,
	values ...map[string]string,
) map[string]corev1.EnvVar {
	options := &helm.Options{SetValues: disableTelemetry(mergeValues(values...))}
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

// Off is the default, and off renders nothing: no trigger worker, and no
// worker publishing to a topic that nobody reads.
func (s *workflowDatasetEventsTemplateTest) TestNothingRendersByDefault() {
	workers := s.workers()

	_, ok := workers[triggerWorkerCommand]
	s.False(ok, "the trigger worker must not render by default")
	for command, container := range workers {
		_, ok := containerEnv(container)[workflowTopicEnv]
		s.False(ok, command+" must not publish for workflows by default")
	}

	for _, tc := range []struct {
		template string
		values   map[string]string
	}{
		{"templates/api-deployment.yaml", map[string]string{}},
		{"templates/app-deployment.yaml", map[string]string{}},
		{
			"templates/plugins-deployment.yaml",
			map[string]string{"pluginsSettings.enabled": "true"},
		},
	} {
		s.Run(tc.template, func() {
			env := s.workloadEnv(tc.template, tc.values)

			_, ok := env[workflowRelayURIEnv]
			s.False(ok)
			_, ok = env[workflowDatasetEventsEnv]
			s.False(ok)
		})
	}
}

// Only the workers that see sample changes publish them: ingest, for events
// that travel the queue, and rollup, which runs the database watch.
func (s *workflowDatasetEventsTemplateTest) TestIngestAndRollupPublishWhenEnabled() {
	workers := s.workers(workflowDatasetEventsOn)

	for command, publishes := range map[string]bool{
		"fiftyone-activity-ingest-worker": true,
		"fiftyone-activity-rollup-worker": true,
		"fiftyone-activity-prune-worker":  false,
		triggerWorkerCommand:              false,
	} {
		s.Run(command, func() {
			container, ok := workers[command]
			s.Require().True(ok, command+" must render")

			env, ok := containerEnv(container)[workflowTopicEnv]
			s.Equal(publishes, ok)
			if publishes {
				s.Equal("true", env.Value)
			}
		})
	}
}

// The trigger worker posts to the relay on teams-api, and names itself with
// the deployment's auth secret, read from the chart's secret and never
// rendered as a value.
func (s *workflowDatasetEventsTemplateTest) TestTriggerWorkerPostsToTheRelay() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"defaultApiService",
			map[string]string{},
			"http://teams-api:80/internal/workflow-triggers",
		},
		{
			"overrideApiService",
			map[string]string{
				"apiSettings.service.name": "test-teams-api",
				"apiSettings.service.port": "8001",
			},
			"http://test-teams-api:8001/internal/workflow-triggers",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workers := s.workers(workflowDatasetEventsOn, tc.values)
			container, ok := workers[triggerWorkerCommand]
			s.Require().True(ok, "the trigger worker must render")
			env := containerEnv(container)

			s.Equal(tc.expected, env[workflowTriggerURIEnv].Value)

			key := env[workflowTriggerKeyEnv]
			s.Empty(key.Value)
			s.Require().NotNil(key.ValueFrom)
			s.Require().NotNil(key.ValueFrom.SecretKeyRef)
			s.Equal("fiftyone-teams-secrets", key.ValueFrom.SecretKeyRef.Name)
			s.Equal("fiftyoneAuthSecret", key.ValueFrom.SecretKeyRef.Key)

			// It reads subscriptions and the queue like every other worker.
			s.Contains(env, "FIFTYONE_ACTIVITY_MONGO_URI")
			s.Contains(env, "FIFTYONE_DATABASE_NAME")
			s.Contains(env, "FIFTYONE_MQ_REDIS_URL")
		})
	}
}

// teams-api relays to whichever service runs the plugins, and the plugin
// servers offer the stages; neither carries the other's variable.
func (s *workflowDatasetEventsTemplateTest) TestApiRelaysAndPluginServersOfferWhenEnabled() {
	api := s.workloadEnv("templates/api-deployment.yaml", workflowDatasetEventsOn)
	s.Equal("http://fiftyone-app:80", api[workflowRelayURIEnv].Value)
	_, ok := api[workflowDatasetEventsEnv]
	s.False(ok)

	dedicated := map[string]string{"pluginsSettings.enabled": "true"}
	api = s.workloadEnv(
		"templates/api-deployment.yaml", workflowDatasetEventsOn, dedicated,
	)
	s.Equal("http://teams-plugins:80", api[workflowRelayURIEnv].Value)

	for _, tc := range []struct {
		template string
		values   map[string]string
	}{
		{"templates/app-deployment.yaml", map[string]string{}},
		{"templates/plugins-deployment.yaml", dedicated},
	} {
		s.Run(tc.template, func() {
			env := s.workloadEnv(tc.template, workflowDatasetEventsOn, tc.values)

			s.Equal("true", env[workflowDatasetEventsEnv].Value)
			_, ok := env[workflowRelayURIEnv]
			s.False(ok)
		})
	}
}

// Dataset events come from Activity Core's database watch. With either off
// the stages would be offered and never fire, so the render fails instead.
func (s *workflowDatasetEventsTemplateTest) TestRenderFailsWithoutWhatFeedsIt() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			"activityOff",
			map[string]string{
				"activitySettings.enabled": "false",
				"fiftyoneMq.enabled":       "false",
			},
			"activitySettings.enabled is false",
		},
		{
			"watchOff",
			map[string]string{"activitySettings.watch.enabled": "false"},
			"activitySettings.watch.enabled is false",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			options := &helm.Options{
				SetValues: disableTelemetry(
					mergeValues(workflowDatasetEventsOn, tc.values),
				),
			}
			_, err := helm.RenderTemplateE(
				s.T(), options, s.chartPath, s.releaseName,
				[]string{"templates/api-deployment.yaml"},
			)
			s.Require().Error(err)
			s.Contains(err.Error(), tc.expected)
		})
	}
}

// The trigger worker has no switch of its own to turn it on.
func (s *workflowDatasetEventsTemplateTest) TestTriggerWorkerNeedsTheSwitch() {
	workers := s.workers(map[string]string{
		"activitySettings.workers.trigger.enabled": "true",
	})

	_, ok := workers[triggerWorkerCommand]
	s.False(ok)
}
