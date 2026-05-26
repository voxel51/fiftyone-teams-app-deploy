//go:build all || helm || unit || unitApiDeployment || unitApiService || unitAppDeployment || unitAppHpa || unitAppService || unitCasDeployment || unitCasService || unitIngress || unitNamespace || unitPluginsDeployment || unitHpaPlugins || unitPluginsService || unitSecrets || unitServiceAccount || unitTeamsAppDeployment || unitTeamsAppHpa || unitTeamsAppService || unitTelemetryRedisDeployment || unitTelemetryRedisPVC || unitTelemetryRedisService || unitTelemetryRole || unitTelemetryRoleBinding
// +build all helm unit unitApiDeployment unitApiService unitAppDeployment unitAppHpa unitAppService unitCasDeployment unitCasService unitIngress unitNamespace unitPluginsDeployment unitHpaPlugins unitPluginsService unitSecrets unitServiceAccount unitTeamsAppDeployment unitTeamsAppHpa unitTeamsAppService unitTelemetryRedisDeployment unitTelemetryRedisPVC unitTelemetryRedisService unitTelemetryRole unitTelemetryRoleBinding

package unit

const chartPath = "../../../helm/fiftyone-teams-app/"
