//go:build docker || compose || unit || unitCompose
// +build docker compose unit unitCompose

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
// the file that declares the volume (docker/compose.yaml).
var builtinServicesSource, _ = filepath.Abs("../../../docker/builtin_services.yaml")

var telemetrySocketDriverOpts = types.Options{
	"device": "tmpfs",
	"o":      "uid=1000,gid=1000,mode=0755",
	"type":   "tmpfs",
}
