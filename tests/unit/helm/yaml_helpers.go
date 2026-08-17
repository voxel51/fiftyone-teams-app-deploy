package unit

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// expectedMqRedisURL is the bundled queue URL the `fiftyone-mq.redis.url`
// helper renders for a release in the default namespace.
func expectedMqRedisURL(releaseName string) string {
	return fmt.Sprintf(
		"redis://%s-fiftyone-mq-redis.fiftyone-teams.svc.cluster.local:6379/0",
		releaseName,
	)
}

// requireMqRedisURL asserts that a workload carries the queue URL AND that
// its value is the one the helper renders.
//
// Asserting only the NAME is not enough: a producer handed an empty
// FIFTYONE_MQ_REDIS_URL, or one wired to the wrong Service, drops every
// emit silently — the same invisible failure the name-only check was
// written to catch, so it has to check the value the workload actually
// connects to.
func requireMqRedisURL(
	t require.TestingT,
	env map[string]corev1.EnvVar,
	releaseName string,
	msg string,
) {
	v, ok := env["FIFTYONE_MQ_REDIS_URL"]
	require.True(t, ok, msg)
	require.Empty(t, v.ValueFrom,
		"the queue URL is a literal value, not a secret/config reference")
	require.Equal(t, expectedMqRedisURL(releaseName), v.Value, msg)
}

// splitYAMLDocs splits a multi-doc YAML string on `---` separators.
// Empty/whitespace-only docs are dropped.
func splitYAMLDocs(s string) []string {
	out := []string{}
	for _, raw := range strings.Split(s, "\n---") {
		doc := strings.TrimSpace(raw)
		if doc == "" {
			continue
		}
		out = append(out, doc)
	}
	return out
}

// disableTelemetry returns a copy of the given helm SetValues map with
// `telemetry.enabled=false` injected as a default. Caller-supplied keys
// (including `telemetry.enabled=true`) win over the default. Use this in
// pre-existing tests that assert non-telemetry deployment shape so the
// chart's new `telemetry.enabled=true` default doesn't add a sidecar
// container / env var / volume that breaks their expectations.
func disableTelemetry(values map[string]string) map[string]string {
	out := map[string]string{"telemetry.enabled": "false"}
	for k, v := range values {
		out[k] = v
	}
	return out
}

// disableDefaultServiceOrchestrators returns a copy of the given helm
// SetValues map that removes the chart's default serviceOrchestrators
// entries (cpuServiceOrc, gpuServiceOrc). Use this in tests that assert
// exact counts or the absence of orchestrator-derived resources, so the
// shipped defaults don't skew their expectations.
func disableDefaultServiceOrchestrators(values map[string]string) map[string]string {
	out := map[string]string{
		"delegatedOperatorJobTemplates.serviceOrchestrators.cpuServiceOrc": "null",
		"delegatedOperatorJobTemplates.serviceOrchestrators.gpuServiceOrc": "null",
	}
	for k, v := range values {
		out[k] = v
	}
	return out
}
