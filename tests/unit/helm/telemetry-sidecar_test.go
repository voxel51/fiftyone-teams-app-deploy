//go:build kubeall || helm || unit || unitTelemetrySidecar
// +build kubeall helm unit unitTelemetrySidecar

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

type telemetrySidecarTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
}

func TestTelemetrySidecarTemplate(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &telemetrySidecarTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   helmChartPath,
		releaseName: "fiftyone-test",
		namespace:   "fiftyone-" + strings.ToLower(random.UniqueId()),
	})
}

// findSidecar locates the telemetry-sidecar container in a slice of containers.
func findSidecar(containers []corev1.Container) *corev1.Container {
	for i, c := range containers {
		if c.Name == "telemetry-sidecar" {
			return &containers[i]
		}
	}
	return nil
}

// TestShareProcessNamespaceEnabledByDefault asserts that the api, app,
// plugins, and delegated-operator deployments all opt into PID-namespace
// sharing when telemetry is enabled (the default). The sidecar relies on
// /proc/<pid>/fd/1 access in the target container's PID namespace, so
// dropping this would silently break log capture.
func (s *telemetrySidecarTemplateTest) TestShareProcessNamespaceEnabledByDefault() {
	cases := []struct {
		template string
		values   map[string]string
	}{
		{"templates/api-deployment.yaml", nil},
		{"templates/app-deployment.yaml", nil},
		{
			"templates/plugins-deployment.yaml",
			map[string]string{"pluginsSettings.enabled": "true"},
		},
		{
			"templates/delegated-operator-instance-deployment.yaml",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled": "true",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		s.Run(tc.template, func() {
			options := &helm.Options{SetValues: tc.values}
			output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName,
				[]string{tc.template})

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(s.T(), output, &deployment)

			s.Require().NotNil(deployment.Spec.Template.Spec.ShareProcessNamespace,
				"shareProcessNamespace should be set on %s", tc.template)
			s.True(*deployment.Spec.Template.Spec.ShareProcessNamespace,
				"shareProcessNamespace should be true on %s", tc.template)
		})
	}
}

// TestShareProcessNamespaceDisabledWithTelemetryOff asserts that
// shareProcessNamespace is NOT set on workloads when telemetry is
// disabled — otherwise we'd be punching an unnecessary hole in pod
// isolation for users who opt out of telemetry.
func (s *telemetrySidecarTemplateTest) TestShareProcessNamespaceDisabledWithTelemetryOff() {
	cases := []struct {
		template string
		values   map[string]string
	}{
		{"templates/api-deployment.yaml", nil},
		{"templates/app-deployment.yaml", nil},
		{
			"templates/plugins-deployment.yaml",
			map[string]string{"pluginsSettings.enabled": "true"},
		},
		{
			"templates/delegated-operator-instance-deployment.yaml",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled": "true",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		s.Run(tc.template, func() {
			options := &helm.Options{SetValues: disableTelemetry(tc.values)}
			output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName,
				[]string{tc.template})

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(s.T(), output, &deployment)

			if deployment.Spec.Template.Spec.ShareProcessNamespace != nil {
				s.False(*deployment.Spec.Template.Spec.ShareProcessNamespace,
					"shareProcessNamespace should not be true on %s when telemetry is disabled",
					tc.template)
			}
		})
	}
}

// TestSidecarSecurityContext asserts the sidecar drops all default caps,
// disables privilege escalation, and only adds SYS_PTRACE on executor (DO)
// sidecars where py-spy crash-stack archives are load-bearing.
func (s *telemetrySidecarTemplateTest) TestSidecarSecurityContext() {
	cases := []struct {
		template string
		values   map[string]string
		executor bool
	}{
		{"templates/api-deployment.yaml", nil, false},
		{"templates/app-deployment.yaml", nil, false},
		{
			"templates/plugins-deployment.yaml",
			map[string]string{"pluginsSettings.enabled": "true"},
			false,
		},
		{
			"templates/delegated-operator-instance-deployment.yaml",
			map[string]string{
				"delegatedOperatorDeployments.deployments.teamsDoCpuDefault.enabled": "true",
			},
			true,
		},
	}

	for _, tc := range cases {
		tc := tc
		s.Run(tc.template, func() {
			options := &helm.Options{SetValues: tc.values}
			output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName,
				[]string{tc.template})

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(s.T(), output, &deployment)

			sidecar := findSidecar(deployment.Spec.Template.Spec.Containers)
			s.Require().NotNil(sidecar, "telemetry-sidecar container should be injected into %s", tc.template)

			sc := sidecar.SecurityContext
			s.Require().NotNil(sc, "telemetry-sidecar should have a securityContext")

			s.Require().NotNil(sc.AllowPrivilegeEscalation, "allowPrivilegeEscalation should be set")
			s.False(*sc.AllowPrivilegeEscalation, "sidecar must not allow privilege escalation")

			s.Require().NotNil(sc.Capabilities, "Capabilities should be set on sidecar")
			s.Contains(sc.Capabilities.Drop, corev1.Capability("ALL"),
				"sidecar must drop all default capabilities")

			var hasPtrace bool
			for _, capability := range sc.Capabilities.Add {
				if capability == "SYS_PTRACE" {
					hasPtrace = true
					break
				}
			}
			if tc.executor {
				s.True(hasPtrace, "executor sidecar must add SYS_PTRACE for py-spy crash archives")
			} else {
				s.False(hasPtrace, "service-mode sidecar must not add SYS_PTRACE")
			}
		})
	}
}
