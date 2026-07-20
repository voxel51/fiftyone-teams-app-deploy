//go:build kubeall || helm || unit || unitSeedOrchestrators
// +build kubeall helm unit unitSeedOrchestrators

package unit

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	batchv1 "k8s.io/api/batch/v1"
)

type seedOrchestratorsJobTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestSeedOrchestratorsJobTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &seedOrchestratorsJobTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
		templates:   []string{"templates/seed-orchestrators-job.yaml"},
	})
}

// seedJobOrchestrators unmarshals the derived ORCHESTRATORS env var from
// the rendered job.
func (s *seedOrchestratorsJobTemplateTest) seedJobOrchestrators(job batchv1.Job) []map[string]interface{} {
	s.T().Helper()
	for _, envVar := range job.Spec.Template.Spec.Containers[0].Env {
		if envVar.Name == "ORCHESTRATORS" {
			var orchestrators []map[string]interface{}
			err := json.Unmarshal([]byte(envVar.Value), &orchestrators)
			s.Require().NoError(err, "ORCHESTRATORS should be valid JSON")
			return orchestrators
		}
	}
	s.Require().Fail("ORCHESTRATORS env var not found")
	return nil
}

// TestGating verifies that the job renders exactly when at least one
// enabled jobs/services entry resolves registerOrchestrator to true (its
// own key when present, otherwise template.registerOrchestrator).
func (s *seedOrchestratorsJobTemplateTest) TestGating() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string // empty => the template must not render
	}{
		{
			// No entries at all: nothing to register
			"defaultValues",
			nil,
			"",
		},
		{
			// Entries register by default (template.registerOrchestrator
			// defaults to true)
			"entriesRegisterByDefault",
			map[string]string{
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":      "nil",
				"delegatedOperatorJobTemplates.services.cpuServices.unused": "nil",
			},
			fmt.Sprintf("%s-fiftyone-teams-app-seed-orchestrators", s.releaseName),
		},
		{
			// A template-level opt-out disables seeding for all entries
			"templateLevelOptOut",
			map[string]string{
				"delegatedOperatorJobTemplates.template.registerOrchestrator": "false",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":        "nil",
				"delegatedOperatorJobTemplates.services.cpuServices.unused":   "nil",
			},
			"",
		},
		{
			// An entry-level opt-in beats the template-level opt-out
			"entryTrueOverridesTemplateFalse",
			map[string]string{
				"delegatedOperatorJobTemplates.template.registerOrchestrator":        "false",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.registerOrchestrator": "true",
			},
			fmt.Sprintf("%s-fiftyone-teams-app-seed-orchestrators", s.releaseName),
		},
		{
			// An entry-level opt-out beats the template-level default
			"entryFalseOverridesTemplateDefault",
			map[string]string{
				"delegatedOperatorJobTemplates.jobs.cpuDefault.registerOrchestrator": "false",
			},
			"",
		},
		{
			// A disabled entry is not registered even when it opts in
			"disabledEntryNotRegistered",
			map[string]string{
				"delegatedOperatorJobTemplates.jobs.cpuDefault.registerOrchestrator": "true",
				"delegatedOperatorJobTemplates.jobs.cpuDefault.enabled":              "false",
			},
			"",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

			options := &helm.Options{SetValues: disableTelemetry(testCase.values)}

			if testCase.expected == "" {
				_, err := helm.RenderTemplateE(subT, options, s.chartPath, s.releaseName, s.templates)
				s.ErrorContains(err, "could not find template templates/seed-orchestrators-job.yaml in chart")
			} else {
				output := helm.RenderTemplate(subT, options, s.chartPath, s.releaseName, s.templates)

				var job batchv1.Job
				helm.UnmarshalK8SYaml(subT, output, &job)

				s.Equal(testCase.expected, job.ObjectMeta.Name, "Name should be set")
			}
		})
	}
}

func (s *seedOrchestratorsJobTemplateTest) TestHelmHooks() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"delegatedOperatorJobTemplates.jobs.cpuDefault.registerOrchestrator": "true",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), output, &job)

	s.Equal(
		"post-install,post-upgrade",
		job.ObjectMeta.Annotations["helm.sh/hook"],
	)
	s.Equal(
		"before-hook-creation,hook-succeeded",
		job.ObjectMeta.Annotations["helm.sh/hook-delete-policy"],
	)
}

// TestDerivedOrchestrators verifies the registration derived for a job and
// a service entry: instance_id from the entry name, environment by map,
// execution_tmpl_uri pointing at the entry's file in the do-templates
// mount, the merged worker image for jobs, and available_operators pinned
// to run_service for services only.
func (s *seedOrchestratorsJobTemplateTest) TestDerivedOrchestrators() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"delegatedOperatorJobTemplates.jobs.cpuDefault.unused":                     "nil",
			"delegatedOperatorJobTemplates.services.cpuServices.description":           "Service orchestrator (CPU)",
			"delegatedOperatorJobTemplates.services.unregistered.registerOrchestrator": "false",
			"delegatedOperatorJobTemplates.jobs.disabledJob.registerOrchestrator":      "true",
			"delegatedOperatorJobTemplates.jobs.disabledJob.enabled":                   "false",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), output, &job)

	orchestrators := s.seedJobOrchestrators(job)
	s.Require().Len(orchestrators, 2)

	// Sorted by map key within each of jobs then services
	jobOrc := orchestrators[0]
	s.Equal("cpuDefault", jobOrc["instance_id"])
	s.Equal("kubernetes", jobOrc["environment"])
	s.Equal("Chart-managed job orchestrator cpuDefault", jobOrc["description"])
	jobConfig := jobOrc["config"].(map[string]interface{})
	s.Equal("/tmp/do-targets/cpuDefault.yaml", jobConfig["execution_tmpl_uri"])
	s.Equal(s.namespaceFromConfig(jobConfig), jobConfig["namespace"])
	// Job orchestrators omit available_operators so the app's Refresh
	// action owns the discovered list
	_, hasAvailableOperators := jobOrc["available_operators"]
	s.False(hasAvailableOperators)
	s.Equal(map[string]interface{}{"kube_config": ""}, jobOrc["secrets"])

	// The job registration tracks the merged worker image at the chart's
	// appVersion
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)
	appVersion, exists := cInfo["appVersion"]
	s.True(exists)
	s.Equal(
		fmt.Sprintf("voxel51/fiftyone-teams-cv-full:%s", appVersion),
		jobConfig["image"],
	)

	serviceOrc := orchestrators[1]
	s.Equal("cpuServices", serviceOrc["instance_id"])
	s.Equal("kubernetes-service", serviceOrc["environment"])
	s.Equal("Service orchestrator (CPU)", serviceOrc["description"])
	serviceConfig := serviceOrc["config"].(map[string]interface{})
	s.Equal("/tmp/do-targets/cpuServices.yaml", serviceConfig["execution_tmpl_uri"])
	// Service orchestrators run only the run_service operator
	s.Equal(
		[]interface{}{"@voxel51/operators/run_service"},
		serviceOrc["available_operators"],
	)
	// Services have no chart-owned image: the broker provides it per task
	_, hasImage := serviceConfig["image"]
	s.False(hasImage)
}

// namespaceFromConfig asserts the config carries a non-empty namespace and
// returns it (the chart default is fiftyone-teams).
func (s *seedOrchestratorsJobTemplateTest) namespaceFromConfig(config map[string]interface{}) interface{} {
	s.T().Helper()
	namespace, ok := config["namespace"]
	s.Require().True(ok, "config should carry a namespace")
	s.Require().NotEmpty(namespace)
	return namespace
}

// TestImages verifies the seeding container runs the teams-api image (the
// script only needs pymongo; the worker image is multi-GB), while the
// derived job registrations still track the delegated-operator worker
// image the rendered Job templates actually run. Distinct sentinel tags
// prove each image flows from its own value.
func (s *seedOrchestratorsJobTemplateTest) TestImages() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"delegatedOperatorJobTemplates.jobs.cpuDefault.registerOrchestrator": "true",
			"delegatedOperatorJobTemplates.template.image.tag":                   "v9.9.9",
			"apiSettings.image.tag": "v8.8.8",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), output, &job)

	container := job.Spec.Template.Spec.Containers[0]
	s.Equal("voxel51/fiftyone-teams-api:v8.8.8", container.Image)

	orchestrators := s.seedJobOrchestrators(job)
	s.Require().Len(orchestrators, 1)
	config := orchestrators[0]["config"].(map[string]interface{})
	s.Equal("voxel51/fiftyone-teams-cv-full:v9.9.9", config["image"])
}

func (s *seedOrchestratorsJobTemplateTest) TestPodSpec() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"delegatedOperatorJobTemplates.jobs.cpuDefault.registerOrchestrator": "true",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), output, &job)

	s.Equal(int32(2), *job.Spec.BackoffLimit)
	s.Equal(int32(600), *job.Spec.TTLSecondsAfterFinished)

	podSpec := job.Spec.Template.Spec

	// Non-root, matching the delegated-operator defaults for the same image
	s.Require().NotNil(podSpec.SecurityContext)
	s.True(*podSpec.SecurityContext.RunAsNonRoot)
	s.Equal(int64(1000), *podSpec.SecurityContext.RunAsUser)

	container := podSpec.Containers[0]
	s.Require().NotNil(container.SecurityContext)
	s.True(*container.SecurityContext.ReadOnlyRootFilesystem)
	s.False(*container.SecurityContext.AllowPrivilegeEscalation)

	// The seeding script arrives inline via .Files.Get
	s.Equal([]string{"python", "-c"}, container.Command)
	s.Require().Len(container.Args, 1)
	s.Contains(container.Args[0], "ORCHESTRATORS")
	s.Contains(container.Args[0], "instance_id")
}
