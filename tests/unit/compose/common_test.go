//go:build docker || compose || unit || unitComposeInternalAuth || unitComposeLegacyAuth
// +build docker compose unit unitComposeInternalAuth unitComposeLegacyAuth

package unit

const (
	envFixtureFilePath = "../../fixtures/docker/.env"
)
