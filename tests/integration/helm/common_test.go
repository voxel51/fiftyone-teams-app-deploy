//go:build docker || helm || integration || integrationHelmInternalAuth || integrationHelmLegacyAuth
// +build docker helm integration integrationHelmInternalAuth integrationHelmLegacyAuth

package integration

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	chartPath           = "../../../helm/fiftyone-teams-app/"
	integrationValues   = "../../fixtures/helm/integration_values.yaml"
	licenseFileInternal = "../../fixtures/helm/internal-license.key"
	licenseFileLegacy   = "../../fixtures/helm/legacy-license.key"
	// for minikube, where node count is 1, we don't need ReadWriteMany and NFS
	persistentVolumeYaml = `---
    apiVersion: v1
    kind: PersistentVolume
    metadata:
      name: pv0001
    spec:
      accessModes:
        - ReadWriteOnce
        - ReadOnlyMany
      capacity:
        storage: 100Mi
      hostPath:
        path: /data/pv0001/
`
	// for minikube, where node count is 1, we don't need ReadWriteMany and NFS
	persistentVolumeClaimYaml = `---
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: pv0001claim
    spec:
      accessModes:
        - ReadWriteOnce
        - ReadOnlyMany
      resources:
        requests:
          storage: 100Mi
`

	// License File Secret
	licenseFileSecretTemplateYaml = `---
    apiVersion: v1
    kind: Secret
    metadata:
      name: fiftyone-license
    type: Opaque
    data:
      license: `
)

type serviceValidations struct {
	name             string
	url              string
	responsePayload  string
	httpResponseCode int
	log              string
}

func validate_endpoint(t *testing.T, url string, expectedBody string, expectedStatus int) {
	maxRetries := 10
	timeBetweenRetries := 3 * time.Second
	http_helper.HttpGetWithRetry(
		t,
		url,
		&tls.Config{
			InsecureSkipVerify: true, // Skip verify because we use self-signed certificates created with cert-manager
		},
		expectedStatus,
		expectedBody,
		maxRetries,
		timeBetweenRetries,
	)
}

// This wrapper is used for test consistency between compose tests
func get_logs(t *testing.T, options *k8s.KubectlOptions, pod *corev1.Pod, containerName string) string {
	output := k8s.GetPodLogs(t, options, pod, containerName)
	return output
}

// Reused from https://github.com/gruntwork-io/terratest/blob/df790dab719e1120f14b4e82c1c8d1150537c026/modules/k8s/tunnel.go#L55-L62
// makeLabels is a helper to format a map of label key and value pairs into a single string for use as a selector.
func makeLabels(labels map[string]string) string {
	out := []string{}
	for key, value := range labels {
		out = append(out, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(out, ",")
}

func getBase64EncodedStringOfFile(filePath string) string {
	b, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Print(err)
	}
	sEnc := base64.StdEncoding.EncodeToString(b)
	return sEnc
}

func waitForTeamsApi(t *testing.T, kubectlOptions *k8s.KubectlOptions, maxRetries int, sleepBetweenRetries time.Duration, deployment *appsv1.Deployment, expected serviceValidations) error {
	statusMsg := fmt.Sprint("Wait for teams-api to start successfully.")
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		maxRetries,
		sleepBetweenRetries,
		func() (string, error) {
			k8s.WaitUntilDeploymentAvailable(t, kubectlOptions, deployment.Name, maxRetries, sleepBetweenRetries)
			selectorLabelsPods := makeLabels(deployment.Spec.Selector.MatchLabels)
			listOptions := metav1.ListOptions{LabelSelector: selectorLabelsPods}
			pods := k8s.ListPods(t, kubectlOptions, listOptions)
			started := false
			for _, pod := range pods {
				started = strings.Contains(get_logs(t, kubectlOptions, &pod, ""), expected.log)
			}
			if !started {
				return "", errors.New("team-api not available")
			}
			return "teams-api is now available", nil
		},
	)
	if err != nil {
		logger.Logf(t, "Timedout waiting for Pod to be provisioned: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}
