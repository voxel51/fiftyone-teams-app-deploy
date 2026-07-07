//go:build kubeall || helm || unit || unitServiceOrchestrator
// +build kubeall helm unit unitServiceOrchestrator

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

func seedJobEnv(job batchv1.Job) map[string]string {
	env := map[string]string{}
	for _, envVar := range job.Spec.Template.Spec.Containers[0].Env {
		env[envVar.Name] = envVar.Value
	}
	return env
}

func (s *seedOrchestratorsJobTemplateTest) TestGating() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected string // empty => the template must not render
	}{
		{
			"defaultValues",
			nil,
			"",
		},
		{
			// Seeding is independent of serviceOrchestrator: it covers any
			// orchestrator environment, so its own flag alone enables it.
			"enabledAloneWithoutServiceOrchestrator",
			map[string]string{
				"seedOrchestrators.enabled": "true",
			},
			fmt.Sprintf("%s-fiftyone-teams-app-seed-orchestrators", s.releaseName),
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
			"seedOrchestrators.enabled": "true",
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

func (s *seedOrchestratorsJobTemplateTest) TestOrchestratorsEnv() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"seedOrchestrators.enabled": "true",
		}),
		SetJsonValues: map[string]string{
			"seedOrchestrators.orchestrators": `[
        {
          "instance_id": "service-orchestrator-cpu",
          "description": "Service orchestrator (CPU)",
          "environment": "kubernetes-service",
          "config": {
            "namespace": "svc-ns",
            "execution_tmpl_uri": "/opt/service-pod-template/pod.yaml.j2",
            "resource_requests": {"cpu": "2", "memory": "8Gi"}
          },
          "secrets": {"kube_config": ""},
          "available_operators": ["@voxel51/operators/run_service"]
        },
        {
          "instance_id": "kubernetes-cpu",
          "description": "Kubernetes CPU",
          "environment": "kubernetes",
          "config": {
            "image": "",
            "execution_tmpl_uri": "/tmp/do-targets/teamsDoK8sCpu.yaml",
            "namespace": "svc-ns"
          },
          "secrets": {"kube_config": ""}
        }
      ]`,
		},
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), output, &job)

	env := seedJobEnv(job)

	// The orchestrator list round-trips as JSON the seeding script consumes
	var orchestrators []map[string]interface{}
	err := json.Unmarshal([]byte(env["ORCHESTRATORS"]), &orchestrators)
	s.NoError(err, "ORCHESTRATORS should be valid JSON")
	s.Require().Len(orchestrators, 2)
	s.Equal("service-orchestrator-cpu", orchestrators[0]["instance_id"])
	config := orchestrators[0]["config"].(map[string]interface{})
	s.Equal("/opt/service-pod-template/pod.yaml.j2", config["execution_tmpl_uri"])
	// The empty-string image sentinel survives for the script to fill from
	// DEFAULT_WORKER_IMAGE
	kubernetesConfig := orchestrators[1]["config"].(map[string]interface{})
	s.Equal("", kubernetesConfig["image"])

	// DEFAULT_WORKER_IMAGE defaults to the delegated-operator worker image
	// at the chart's appVersion
	cInfo, err := chartInfo(s.T(), s.chartPath)
	s.NoError(err)
	appVersion, exists := cInfo["appVersion"]
	s.True(exists)
	s.Equal(
		fmt.Sprintf("voxel51/fiftyone-teams-cv-full:%s", appVersion),
		env["DEFAULT_WORKER_IMAGE"],
	)
}

func (s *seedOrchestratorsJobTemplateTest) TestImageFollowsDelegatedOperatorTemplate() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"seedOrchestrators.enabled":                        "true",
			"delegatedOperatorJobTemplates.template.image.tag": "v9.9.9",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), output, &job)

	container := job.Spec.Template.Spec.Containers[0]
	s.Equal("voxel51/fiftyone-teams-cv-full:v9.9.9", container.Image)
	s.Equal(
		"voxel51/fiftyone-teams-cv-full:v9.9.9",
		seedJobEnv(job)["DEFAULT_WORKER_IMAGE"],
	)
}

func (s *seedOrchestratorsJobTemplateTest) TestImageOverride() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"seedOrchestrators.enabled":          "true",
			"seedOrchestrators.image.repository": "custom/seeder",
			"seedOrchestrators.image.tag":        "v1.2.3",
		}),
	}

	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var job batchv1.Job
	helm.UnmarshalK8SYaml(s.T(), output, &job)

	s.Equal("custom/seeder:v1.2.3", job.Spec.Template.Spec.Containers[0].Image)
}

func (s *seedOrchestratorsJobTemplateTest) TestPodSpec() {
	options := &helm.Options{
		SetValues: disableTelemetry(map[string]string{
			"seedOrchestrators.enabled": "true",
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
	s.Contains(container.Args[0], "DEFAULT_WORKER_IMAGE")
}
