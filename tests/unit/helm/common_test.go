//go:build all || helm || unit || unitApiDeployment || unitApiService || unitAppDeployment || unitAppHpa || unitAppService || unitIngress || unitNamespace || unitPluginsDeployment || unitHpaPlugins || unitPluginsService || unitSecrets || unitServiceAccount || unitTeamsAppDeployment || unitTeamsAppHpa || unitTeamsAppService
// +build all helm unit unitApiDeployment unitApiService unitAppDeployment unitAppHpa unitAppService unitIngress unitNamespace unitPluginsDeployment unitHpaPlugins unitPluginsService unitSecrets unitServiceAccount unitTeamsAppDeployment unitTeamsAppHpa unitTeamsAppService

package unit

const chartPath = "../../../helm/fiftyone-teams-app/"
