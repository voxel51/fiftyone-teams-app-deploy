package unit

import corev1 "k8s.io/api/core/v1"

// findContainer returns a pointer to the container with the given name, or nil
// if no container matches. Shared by tests that introspect rendered pod specs.
func findContainer(containers []corev1.Container, name string) *corev1.Container {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}

// findVolume returns a pointer to the pod volume with the given name, or nil.
func findVolume(volumes []corev1.Volume, name string) *corev1.Volume {
	for i := range volumes {
		if volumes[i].Name == name {
			return &volumes[i]
		}
	}
	return nil
}

// findVolumeMount returns a pointer to the volume mount with the given name, or nil.
func findVolumeMount(mounts []corev1.VolumeMount, name string) *corev1.VolumeMount {
	for i := range mounts {
		if mounts[i].Name == name {
			return &mounts[i]
		}
	}
	return nil
}

// envValue returns the value of the env var with the given name and whether it
// was found.
func envValue(env []corev1.EnvVar, name string) (string, bool) {
	for _, e := range env {
		if e.Name == name {
			return e.Value, true
		}
	}
	return "", false
}
