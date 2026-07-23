//go:build docker || compose || unit || unitComposeInternalAuth || unitComposeLegacyAuth
// +build docker compose unit unitComposeInternalAuth unitComposeLegacyAuth

package unit

import "github.com/compose-spec/compose-go/v2/types"

const (
	envFixtureFilePath = "../../fixtures/docker/.env"
)

var telemetrySocketDriverOpts = types.Options{
	"device": "tmpfs",
	"o":      "uid=1000,gid=1000,mode=0755",
	"type":   "tmpfs",
}
