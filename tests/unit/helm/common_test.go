//go:build all || helm || unit || unitApiDeployment || unitApiService || unitAppDeployment || unitAppHpa || unitAppService || unitCasDeployment || unitCasService || unitFiftyoneMqRedisDeployment || unitFiftyoneMqRedisService || unitIngress || unitNamespace || unitPluginsDeployment || unitHpaPlugins || unitPluginsService || unitSecrets || unitServiceAccount || unitTeamsAppDeployment || unitTeamsAppHpa || unitTeamsAppService || unitTelemetryRedis || unitTelemetryRoleBinding || unitTelemetrySidecar || unitWorkflowDatasetEvents || unitWorkflowSchedulesEnv
// +build all helm unit unitApiDeployment unitApiService unitAppDeployment unitAppHpa unitAppService unitCasDeployment unitCasService unitFiftyoneMqRedisDeployment unitFiftyoneMqRedisService unitIngress unitNamespace unitPluginsDeployment unitHpaPlugins unitPluginsService unitSecrets unitServiceAccount unitTeamsAppDeployment unitTeamsAppHpa unitTeamsAppService unitTelemetryRedis unitTelemetryRoleBinding unitTelemetrySidecar unitWorkflowDatasetEvents unitWorkflowSchedulesEnv

package unit

const chartPath = "../../../helm/fiftyone-teams-app/"
