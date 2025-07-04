//go:build kubeall || helm || integration || integrationHelmTopology
// +build kubeall helm integration integrationHelmTopology

package integration

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"path/filepath"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type topologyAuthHelmTest struct {
	suite.Suite
	chartPath   string
	namespace   string
	context     string
	valuesFiles []string
}

func TestHelmTopologyAuth(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	kubeCtx := defineKubeCtx()
	integrationValuesPath, err := filepath.Abs(integrationValues)

	suite.Run(t, &topologyAuthHelmTest{
		Suite:     suite.Suite{},
		chartPath: helmChartPath,
		namespace: "fiftyone-" + strings.ToLower(random.UniqueId()),
		context:   kubeCtx,
		valuesFiles: []string{
			integrationValuesPath, // Copy of values from `skaffold.yaml`'s `helm.releases[0].overrides`
		},
	})
}

func (s *topologyAuthHelmTest) TestHelmInstall() {
	testCases := []struct {
		name     string
		values   map[string]string
		expected []serviceValidations
	}{
		{
			"dedicatedPluginsTopologyConstraints", // plugins run in plugins deployment
			map[string]string{
				"secret.fiftyone.fiftyoneDatabaseName":                            "fiftyone-top-dp-" + suffix,
				"apiSettings.env.FIFTYONE_PLUGINS_DIR":                            "/opt/plugins",
				"apiSettings.topologySpreadConstraints[0].maxSkew":                "1",
				"apiSettings.topologySpreadConstraints[0].topologyKey":            "kubernetes.io/hostname",
				"apiSettings.topologySpreadConstraints[0].whenUnsatisfiable":      "ScheduleAnyway",
				"apiSettings.volumes[0].name":                                     "plugins-vol",
				"apiSettings.volumes[0].persistentVolumeClaim.claimName":          "pvc-top-dp-" + suffix,
				"apiSettings.volumeMounts[0].name":                                "plugins-vol",
				"apiSettings.volumeMounts[0].mountPath":                           "/opt/plugins",
				"appSettings.topologySpreadConstraints[0].maxSkew":                "1",
				"appSettings.topologySpreadConstraints[0].topologyKey":            "kubernetes.io/hostname",
				"appSettings.topologySpreadConstraints[0].whenUnsatisfiable":      "ScheduleAnyway",
				"casSettings.env.FIFTYONE_AUTH_MODE":                              "internal",
				"casSettings.env.CAS_DATABASE_NAME":                               "cas-top-dp-" + suffix,
				"casSettings.topologySpreadConstraints[0].maxSkew":                "1",
				"casSettings.topologySpreadConstraints[0].topologyKey":            "kubernetes.io/hostname",
				"casSettings.topologySpreadConstraints[0].whenUnsatisfiable":      "ScheduleAnyway",
				"pluginsSettings.enabled":                                         "true",
				"pluginsSettings.env.FIFTYONE_PLUGINS_DIR":                        "/opt/plugins",
				"pluginsSettings.topologySpreadConstraints[0].maxSkew":            "1",
				"pluginsSettings.topologySpreadConstraints[0].topologyKey":        "kubernetes.io/hostname",
				"pluginsSettings.topologySpreadConstraints[0].whenUnsatisfiable":  "ScheduleAnyway",
				"pluginsSettings.volumes[0].name":                                 "plugins-vol-ro",
				"pluginsSettings.volumes[0].persistentVolumeClaim.claimName":      "pvc-top-dp-" + suffix,
				"pluginsSettings.volumes[0].persistentVolumeClaim.readOnly":       "true",
				"pluginsSettings.volumeMounts[0].name":                            "plugins-vol-ro",
				"pluginsSettings.volumeMounts[0].mountPath":                       "/opt/plugins",
				"teamsAppSettings.topologySpreadConstraints[0].maxSkew":           "1",
				"teamsAppSettings.topologySpreadConstraints[0].topologyKey":       "kubernetes.io/hostname",
				"teamsAppSettings.topologySpreadConstraints[0].whenUnsatisfiable": "ScheduleAnyway",
			},
			[]serviceValidations{
				{
					name:             "teams-api",
					url:              ternary(s.context == "minikube", "https://local.fiftyone.ai/health", ""),
					responsePayload:  `{"status":{"teams":"available"}}`,
					httpResponseCode: 200,
					log:              "Starting worker",
				},
				{
					name:             "teams-app",
					url:              ternary(s.context == "minikube", "https://local.fiftyone.ai/api/hello", ""),
					responsePayload:  `{"name":"John Doe"}`,
					httpResponseCode: 200,
					log:              " ✓ Ready in",
				},
				{
					name:             "teams-cas",
					url:              ternary(s.context == "minikube", "https://local.fiftyone.ai/cas/api", ""),
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
					log:              "Running on http://0.0.0.0:5151",
				},
				{
					name:             "teams-plugins",
					url:              "",
					responsePayload:  "",
					httpResponseCode: 0,
					log:              "Running on http://0.0.0.0:5151", // same as fiftyone-app since plugins uses or is based on the fiftyone-app image
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		s.Run(testCase.name, func() {
			subT := s.T()
			subT.Parallel()

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
			needsPersistentVolume := []string{"dedicatedPluginsTopologyConstraints"}
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
					Name:             testCase.values["apiSettings.volumes[0].persistentVolumeClaim.claimName"],
					AccessModes:      []string{"ReadWriteOnce", "ReadOnlyMany"},
					Capacity:         pvCapacity,
					StorageClassName: pvStorageClassName,
					HostPath:         hostPath,
					NFS:              nfsConfig,
				}

				pvc := PersistentVolumeClaim{
					Name:             testCase.values["apiSettings.volumes[0].persistentVolumeClaim.claimName"],
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
			base64EncodedLicenseFile := getBase64EncodedStringOfFile(licenseFileInternal)
			defer k8s.KubectlDeleteFromString(subT, kubectlOptions, licenseFileSecretTemplateYaml+base64EncodedLicenseFile)
			k8s.KubectlApplyFromString(subT, kubectlOptions, licenseFileSecretTemplateYaml+base64EncodedLicenseFile)

			helmOptions := &helm.Options{
				KubectlOptions: kubectlOptions,
				SetValues:      testCase.values,
				ValuesFiles:    s.valuesFiles,
			}
			releaseName := fmt.Sprintf(
				"%s-topology-%s",
				strings.ToLower(testCase.name),
				strings.ToLower(random.UniqueId()),
			)
			defer helm.Delete(subT, helmOptions, releaseName, true)
			helm.Install(subT, helmOptions, s.chartPath, releaseName)

			enforceReady(subT, kubectlOptions, testCase.expected)

			// Validate system health
			for _, expected := range testCase.expected {
				logger.Log(subT, fmt.Sprintf("Validating service %s...", expected.name))

				// get deployment
				deployment := k8s.GetDeployment(subT, kubectlOptions, expected.name)

				// get deployment match labels
				selectorLabelsPods := makeLabels(deployment.Spec.Selector.MatchLabels)

				// use deployment match labels to get the associated pods
				listOptions := metav1.ListOptions{LabelSelector: selectorLabelsPods}
				pods := k8s.ListPods(subT, kubectlOptions, listOptions)

				// Validate log output is expected
				checkPodLogsWithRetries(subT, kubectlOptions, pods, testCase.name, expected.name, expected.log)

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
