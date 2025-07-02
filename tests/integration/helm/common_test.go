//go:build docker || helm || integration || integrationHelmInternalAuth || integrationHelmLegacyAuth || integrationHelmTopology
// +build docker helm integration integrationHelmInternalAuth integrationHelmLegacyAuth integrationHelmTopology

package integration

import (
	"crypto/tls"

	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
)

type NFSConfig struct {
	Server string
	Path   string
}

type PersistentVolume struct {
	Name             string
	AccessModes      []string
	Capacity         string
	StorageClassName string
	HostPath         string
	NFS              *NFSConfig // Pointer to allow for nil checking
}

type PersistentVolumeClaim struct {
	Name             string
	AccessModes      []string
	Capacity         string
	HostPath         string
	StorageClassName string
	VolumeName       string
}

const (
	chartPath           = "../../../helm/fiftyone-teams-app/"
	integrationValues   = "../../fixtures/helm/integration_values.yaml"
	licenseFileInternal = "../../fixtures/helm/internal-license.key"
	licenseFileLegacy   = "../../fixtures/helm/legacy-license.key"
	// for minikube, where node count is 1, we don't need ReadWriteMany and NFS
	persistentVolumeYamlTpl = `---
apiVersion: v1
kind: PersistentVolume
metadata:
    name: {{ .Name }}
spec:
    accessModes:
        - ReadWriteOnce
        - ReadOnlyMany
    capacity:
        storage: {{ .Capacity }}
    {{- if .HostPath }}
    hostPath:
        path: {{ .HostPath }}
    {{- else if .NFS }}
    nfs:
        server: {{ .NFS.Server }}
        path: {{ .NFS.Path }}
    {{- end }}
    storageClassName: {{ .StorageClassName }}
`
	// for minikube, where node count is 1, we don't need ReadWriteMany and NFS
	persistentVolumeClaimYamlTpl = `---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
    name: {{ .Name }}
spec:
    accessModes:
        - ReadWriteOnce
        - ReadOnlyMany
    storageClassName: {{ .StorageClassName }}
    volumeName: {{ .VolumeName }}
    resources:
        requests:
            storage: {{ .Capacity }}
`
	nfsExportPath      = "/ephemeral-integration-tests/plugins"
	nfsExportServer    = "voxel51-dev-nfs-server.us-east5-a.c.computer-vision-team.internal"
	pvCapacity         = "100Mi"
	pvStorageClassName = "\"\""
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

var (
	suffix = generateRandomString(6)
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

func defineKubeCtx() string {
	kubeCtx := "minikube"
	requiredSubstring := "voxel51-ephemeral-test" // enforce it goes to ephemeral env
	if kc := os.Getenv("INTEGRATION_TEST_KUBECONTEXT"); kc != "" {
		if strings.Contains(kc, requiredSubstring) {
			kubeCtx = kc
		} else {
			fmt.Printf("The string '%s' does not contain the required context slug. Defaulting to minikube.\n", kc)
		}
	}
	return kubeCtx
}

func renderTemplate(templateString string, data interface{}) (string, error) {
	tmpl, err := template.New("resource").Parse(templateString)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func pvToYaml(pv PersistentVolume) (string, error) {
	return renderTemplate(persistentVolumeYamlTpl, pv)
}

func pvcToYaml(pvc PersistentVolumeClaim) (string, error) {
	return renderTemplate(persistentVolumeClaimYamlTpl, pvc)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a byte slice to store the random characters
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func enforceReady(subT *testing.T, kubectlOptions *k8s.KubectlOptions, vals []serviceValidations) {
	// Pods might have to connect to each other. So, we should
	// wait for all pods to be ready before doing any log checks.
	waitTime := 5 * time.Second
	retries := 96
	for _, expected := range vals {
		deployment := k8s.GetDeployment(subT, kubectlOptions, expected.name)
		// when pulling images for the first time, it may take longer than 90s
		// 360 seconds of retries. Pods typically ready in ~51 seconds if the image is already pulled.
		k8s.WaitUntilDeploymentAvailable(subT, kubectlOptions, deployment.Name, retries, waitTime)

		// Validate that k8s service is ready (pods are started and in service)

		if expected.name != "teams-do" {
			k8s.WaitUntilServiceAvailable(subT, kubectlOptions, expected.name, 10, 1*time.Second)
		}
	}
}

func checkPodLogsWithRetries(subT *testing.T, kubectlOptions *k8s.KubectlOptions, pods []corev1.Pod, tc string, svc string, expected string) {
	// The pods report they're ready before the final log that we sometimes
	// test. The root issue is that pods report ready before they truly are.
	// Once that is fixed, this test becomes redundant and isn't required.
	maxRetries := 6
	retryDelay := 2 * time.Second

	if svc == "teams-do" {
		// AF: DO has no health/readiness probes. It also depends on the API
		// so I think we should bump the timeout to address test flakes
		maxRetries = 10
	}

	for _, pod := range pods {
		var log string

		for i := 0; i < maxRetries; i++ {
			log = get_logs(subT, kubectlOptions, &pod, "")
			if strings.Contains(log, expected) {
				// Log entry found, proceed to next pod
				break
			}

			// Log entry not found, wait before retrying
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
			}
		}

		// Final assertion
		subT.Run(fmt.Sprintf("%s - %s", tc, svc), func(t *testing.T) {
			if !strings.Contains(log, expected) {
				t.Errorf("[ERROR]: %s - %s - log should contain matching entry:\n\t%s", tc, svc, expected)
			}
		})
	}
}

func ternary(condition bool, a string, b string) string {
	if condition {
		return a
	}
	return b
}
