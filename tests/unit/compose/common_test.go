//go:build docker || compose || unit || unitComposeInternalAuth || unitComposeLegacyAuth
// +build docker compose unit unitComposeInternalAuth unitComposeLegacyAuth

package unit

import (
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
)

const (
	envFixtureFilePath = "../../fixtures/docker/.env"
)

// builtinServicesSource is the absolute path compose-go resolves the
// builtin_services.yaml bind mount to. It resolves relative to the directory of
// the file that declares the volume (common-services.yaml, i.e. docker/).
var builtinServicesSource, _ = filepath.Abs("../../../docker/builtin_services.yaml")

var telemetrySocketDriverOpts = types.Options{
	"device": "tmpfs",
	"o":      "uid=1000,gid=1000,mode=0755",
	"type":   "tmpfs",
}
