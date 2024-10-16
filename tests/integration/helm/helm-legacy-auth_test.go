//go:build kubeall || helm || integration || integrationHelmLegacyAuth
// +build kubeall helm integration integrationHelmLegacyAuth

package integration

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"path/filepath"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type legacyAuthHelmTest struct {
	suite.Suite
	chartPath   string
	namespace   string
	context     string
	valuesFiles []string
}

func TestHelmLegacyAuth(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)
	kubeCtx := defineKubeCtx()
	integrationValuesPath, err := filepath.Abs(integrationValues)

	suite.Run(t, &legacyAuthHelmTest{
		Suite:     suite.Suite{},
		chartPath: helmChartPath,
		namespace: "fiftyone-" + strings.ToLower(random.UniqueId()),
		context:   kubeCtx,
		valuesFiles: []string{
			integrationValuesPath, // Copy of values from `skaffold.yaml`'s `helm.releases[0].overrides`
		},
	})
}

func (s *legacyAuthHelmTest) TestHelmInstall() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []serviceValidations
	}{
		{
			"builtinPlugins",
			map[string]string{
				"casSettings.env.FIFTYONE_AUTH_MODE": "legacy",
			},
			[]serviceValidations{
				{
					name:             "teams-api",
					url:              "", // "https://local.fiftyone.ai/health",
					responsePayload:  `{"status":{"teams":"available"}}`,
					httpResponseCode: 200,
					log:              "[INFO] Starting worker",
				},
				{
					name:             "teams-app",
					url:              "", // "https://local.fiftyone.ai/api/hello",
					responsePayload:  `{"name":"John Doe"}`,
					httpResponseCode: 200,
					log:              "Listening on port 3000",
				},
				{
					name:             "teams-cas",
					url:              "", // "https://local.fiftyone.ai/cas/api",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              " ✓ Ready in",
				},
				// ordering this last to avoid test flakes where testing for log before the container is running
				{
					name:             "fiftyone-app",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 200,
					log:              "[INFO] Running on http://0.0.0.0:5151",
				},
			},
		},
		{
			"sharedPlugins", // plugins run in fiftyone-app deployment
			map[string]string{
				"casSettings.env.FIFTYONE_AUTH_MODE":                     "legacy",
				"apiSettings.env.FIFTYONE_PLUGINS_DIR":                   "/opt/plugins",
				"apiSettings.volumes[0].name":                            "plugins-vol",
				"apiSettings.volumes[0].persistentVolumeClaim.claimName": "pvc-" + pvSuffix,
				"apiSettings.volumeMounts[0].name":                       "plugins-vol",
				"apiSettings.volumeMounts[0].mountPath":                  "/opt/plugins",
				"appSettings.env.FIFTYONE_PLUGINS_DIR":                   "/opt/plugins",
				"appSettings.volumes[0].name":                            "plugins-vol-ro",
				"appSettings.volumes[0].persistentVolumeClaim.claimName": "pvc-" + pvSuffix,
				"appSettings.volumes[0].persistentVolumeClaim.readOnly":  "true",
				"appSettings.volumeMounts[0].name":                       "plugins-vol-ro",
				"appSettings.volumeMounts[0].mountPath":                  "/opt/plugins",
			},
			[]serviceValidations{
				{
					name:             "teams-api",
					url:              "", // "https://local.fiftyone.ai/health",
					responsePayload:  `{"status":{"teams":"available"}}`,
					httpResponseCode: 200,
					log:              "[INFO] Starting worker",
				},
				{
					name:             "teams-app",
					url:              "", // "https://local.fiftyone.ai/api/hello",
					responsePayload:  `{"name":"John Doe"}`,
					httpResponseCode: 200,
					log:              "Listening on port 3000",
				},
				{
					name:             "teams-cas",
					url:              "", // "https://local.fiftyone.ai/cas/api",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              " ✓ Ready in",
				},
				// ordering this last to avoid test flakes where testing for log before the container is running
				{
					name:             "fiftyone-app",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 200,
					log:              "[INFO] Running on http://0.0.0.0:5151",
				},
			},
		},
		{
			"dedicatedPlugins", // plugins run in plugins deployment
			map[string]string{
				"apiSettings.env.FIFTYONE_PLUGINS_DIR":                       "/opt/plugins",
				"apiSettings.volumes[0].name":                                "plugins-vol",
				"apiSettings.volumes[0].persistentVolumeClaim.claimName":     "pvc-" + pvSuffix,
				"apiSettings.volumeMounts[0].name":                           "plugins-vol",
				"apiSettings.volumeMounts[0].mountPath":                      "/opt/plugins",
				"casSettings.env.FIFTYONE_AUTH_MODE":                         "legacy",
				"pluginsSettings.enabled":                                    "true",
				"pluginsSettings.env.FIFTYONE_PLUGINS_DIR":                   "/opt/plugins",
				"pluginsSettings.volumes[0].name":                            "plugins-vol-ro",
				"pluginsSettings.volumes[0].persistentVolumeClaim.claimName": "pvc-" + pvSuffix,
				"pluginsSettings.volumes[0].persistentVolumeClaim.readOnly":  "true",
				"pluginsSettings.volumeMounts[0].name":                       "plugins-vol-ro",
				"pluginsSettings.volumeMounts[0].mountPath":                  "/opt/plugins",
			},
			[]serviceValidations{
				{
					name:             "teams-api",
					url:              "", // "https://local.fiftyone.ai/health",
					responsePayload:  `{"status":{"teams":"available"}}`,
					httpResponseCode: 200,
					log:              "[INFO] Starting worker",
				},
				{
					name:             "teams-app",
					url:              "", //"https://local.fiftyone.ai/api/hello",
					responsePayload:  `{"name":"John Doe"}`,
					httpResponseCode: 200,
					log:              "Listening on port 3000",
				},
				{
					name:             "teams-cas",
					url:              "", // "https://local.fiftyone.ai/cas/api",
					responsePayload:  `{"status":"available"}`,
					httpResponseCode: 200,
					log:              " ✓ Ready in",
				},
				// ordering this last to avoid test flakes where testing for log before the container is running
				{
					name:             "fiftyone-app",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 200,
					log:              "[INFO] Running on http://0.0.0.0:5151",
				},
				{
					name:             "teams-plugins",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 0,
					log:              "[INFO] Running on http://0.0.0.0:5151", // same as fiftyone-app since plugins uses or is based on the fiftyone-app image
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			// Disabling parallelization until we configure discrete databases
			// subT.Parallel()

			// Create namespace name for the test case
			namespace := fmt.Sprintf(
				"%s-%s",
				s.namespace,
				strings.ToLower(testCase.name),
			)
			//  Add namespace to helm values map
			testCase.values["namespace.name"] = namespace

			// Add namespace to kubectl options
			kubectlOptions := k8s.NewKubectlOptions(s.context, "", namespace)
			// Create namespace
			defer k8s.DeleteNamespace(subT, kubectlOptions, namespace)
			k8s.CreateNamespace(subT, kubectlOptions, namespace)

			// create persistent volume, when necessary
			needsPersistentVolume := []string{"sharedPlugins", "dedicatedPlugins"}
			if slices.Contains(needsPersistentVolume, testCase.name) {

				var nfsConfig *NFSConfig
				hostPath := "/data/pv0001/"

				if s.context != "minikube" {
					nfsConfig = &NFSConfig{
						Server: nfsExportServer,
						Path:   nfsExportPath,
					}
					hostPath = ""
				}

				pv := PersistentVolume{
					Name:             "pv-" + pvSuffix,
					AccessModes:      []string{"ReadWriteOnce", "ReadOnlyMany"},
					Capacity:         pvCapacity,
					StorageClassName: pvStorageClassName,
					HostPath:         hostPath,
					NFS:              nfsConfig,
				}

				pvc := PersistentVolumeClaim{
					Name:             "pvc-" + pvSuffix,
					AccessModes:      []string{"ReadWriteOnce", "ReadOnlyMany"},
					Capacity:         pv.Capacity,
					VolumeName:       pv.Name,
					StorageClassName: pv.StorageClassName,
				}

				persistentVolumeYaml, err := pvToYaml(pv)

				if err != nil {
					panic(err)
				}
				persistentVolumeClaimYaml, err := pvcToYaml(pvc)

				if err != nil {
					panic(err)
				}

				defer k8s.KubectlDeleteFromString(subT, kubectlOptions, persistentVolumeYaml)
				k8s.KubectlApplyFromString(subT, kubectlOptions, persistentVolumeYaml)
				defer k8s.KubectlDeleteFromString(subT, kubectlOptions, persistentVolumeClaimYaml)
				k8s.KubectlApplyFromString(subT, kubectlOptions, persistentVolumeClaimYaml)
			}

			// create license-file secret
			base64EncodedLicenseFile := getBase64EncodedStringOfFile(licenseFileLegacy)
			defer k8s.KubectlDeleteFromString(subT, kubectlOptions, licenseFileSecretTemplateYaml+base64EncodedLicenseFile)
			k8s.KubectlApplyFromString(subT, kubectlOptions, licenseFileSecretTemplateYaml+base64EncodedLicenseFile)

			helmOptions := &helm.Options{
				KubectlOptions: kubectlOptions,
				SetValues:      testCase.values,
				ValuesFiles:    s.valuesFiles,
			}
			releaseName := fmt.Sprintf(
				"%s-legacy-%s",
				strings.ToLower(testCase.name),
				strings.ToLower(random.UniqueId()),
			)
			defer helm.Delete(subT, helmOptions, releaseName, true)
			helm.Install(subT, helmOptions, s.chartPath, releaseName)

			// Validate system health
			for _, expected := range testCase.expected {
				logger.Log(subT, fmt.Sprintf("Validating service %s...", expected.name))

				// get deployment
				deployment := k8s.GetDeployment(subT, kubectlOptions, expected.name)

				waitTime := 5 * time.Second
				retries := 72

				// when pulling images for the first time, it may take longer than 90s
				// 360 seconds of retries. Pods typically ready in ~51 seconds if the image is already pulled.
				k8s.WaitUntilDeploymentAvailable(subT, kubectlOptions, deployment.Name, retries, waitTime)

				// get deployment match labels
				selectorLabelsPods := makeLabels(deployment.Spec.Selector.MatchLabels)

				// use deployment match labels to get the associated pods
				listOptions := metav1.ListOptions{LabelSelector: selectorLabelsPods}
				pods := k8s.ListPods(subT, kubectlOptions, listOptions)

				// Validate log output is expected
				for _, pod := range pods {
					s.Contains(
						get_logs(subT, kubectlOptions, &pod, ""),
						expected.log,
						fmt.Sprintf("%s - %s - log should contain matching entry", testCase.name, expected.name),
					)
				}

				// Validate that k8s service is ready (pods are started and in service)
				k8s.WaitUntilServiceAvailable(subT, kubectlOptions, expected.name, 10, 1*time.Second)

				// Validate endpoint response
				// Skip fiftyone-app and teams-plugins because they do not have callable endpoints that return a response payload.
				if expected.url != "" {
					// Validate url endpoint response is expected
					validate_endpoint(subT, expected.url, expected.responsePayload, expected.httpResponseCode)
				}
			}
		})
	}
}
