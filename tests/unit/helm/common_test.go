//go:build all || helm || unit || unitApiDeployment || unitApiService || unitAppDeployment || unitAppHpa || unitAppService || unitCasDeployment || unitCasService || unitIngress || unitNamespace || unitPluginsDeployment || unitHpaPlugins || unitPluginsService || unitSecrets || unitServiceAccount || unitTeamsAppDeployment || unitTeamsAppHpa || unitTeamsAppService || unitTelemetryRedis || unitTelemetryRoleBinding || unitTelemetrySidecar
// +build all helm unit unitApiDeployment unitApiService unitAppDeployment unitAppHpa unitAppService unitCasDeployment unitCasService unitIngress unitNamespace unitPluginsDeployment unitHpaPlugins unitPluginsService unitSecrets unitServiceAccount unitTeamsAppDeployment unitTeamsAppHpa unitTeamsAppService unitTelemetryRedis unitTelemetryRoleBinding unitTelemetrySidecar

package unit

const chartPath = "../../../helm/fiftyone-teams-app/"
