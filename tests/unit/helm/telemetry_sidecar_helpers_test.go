//go:build kubeall || helm || unit || unitApiDeployment || unitAppDeployment || unitPluginsDeployment || unitConfigMap
// +build kubeall helm unit unitApiDeployment unitAppDeployment unitPluginsDeployment unitConfigMap

package unit

import (
	corev1 "k8s.io/api/core/v1"
)

// findContainerByName returns a pointer into the slice for the first container
// with the given name, or nil. Used to locate the auto-injected
// telemetry-sidecar in test assertions without depending on slice ordering.
func findContainerByName(containers []corev1.Container, name string) *corev1.Container {
	for i, c := range containers {
		if c.Name == name {
			return &containers[i]
		}
	}
	return nil
}
