package unit

import "strings"

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
