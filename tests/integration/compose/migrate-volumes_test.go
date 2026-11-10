//go:build docker || compose || integration || integrationCompose
// +build docker compose integration integrationCompose

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	migrateDockerDir    = "../../../docker"
	migrateScript       = "./migrate-volumes.sh"
	migrateCopyImage    = "redis:8-alpine"
	migrateVolumeKey    = "plugins-vol"
	migrateProbeCompose = `services:
  probe:
    image: redis:8-alpine
    volumes:
      - plugins-vol:/opt/plugins
`
)

// migrateEnv describes one test's destination project
// and the docker compose arguments the script receives after `--`.
type migrateEnv struct {
	t           *testing.T
	destProject string
	composeArgs []string
	env         []string
}

func newMigrateEnv(t *testing.T, files ...string) *migrateEnv {
	t.Helper()
	requireNoOtherPluginVolumes(t)
	ensureImage(t, migrateCopyImage)

	dest := uniqueName("mvdest")
	args := []string{
		"--env-file", "env.template",
		"--env-file", "../tests/fixtures/docker/.env",
		"-p", dest,
	}
	for _, f := range files {
		args = append(args, "-f", f)
	}
	t.Cleanup(func() {
		removeContainers(t, dest)
		removeVolume(t, dest+"_"+migrateVolumeKey)
	})
	return &migrateEnv{t: t, destProject: dest, composeArgs: args}
}

func (e *migrateEnv) destVolume() string {
	return e.destProject + "_" + migrateVolumeKey
}

// migrate runs the script from docker/
// and returns its combined output and exit code.
func (e *migrateEnv) migrate(flags ...string) (string, int) {
	e.t.Helper()
	args := append([]string{}, flags...)
	args = append(args, "--")
	args = append(args, e.composeArgs...)
	cmd := exec.Command(migrateScript, args...)
	cmd.Dir = migrateDockerDir
	cmd.Env = append(os.Environ(), e.env...)
	out, err := cmd.CombinedOutput()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else {
		require.NoError(e.t, err, string(out))
	}
	e.t.Logf("migrate-volumes.sh %s\n%s", strings.Join(args, " "), out)
	return string(out), code
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
}

func run(t *testing.T, stdin string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s %s: %s", name, strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func ensureImage(t *testing.T, image string) {
	t.Helper()
	if exec.Command("docker", "image", "inspect", image).Run() != nil {
		run(t, "", "docker", "pull", "-q", image)
	}
}

// requireNoOtherPluginVolumes skips the test on hosts
// that already have Compose plugins volumes,
// because the script discovers sources host-wide.
func requireNoOtherPluginVolumes(t *testing.T) {
	t.Helper()
	out := run(t, "", "docker", "volume", "ls", "-q", "--filter", "label=com.docker.compose.volume="+migrateVolumeKey)
	if out != "" {
		t.Skipf("host already has Compose %s volumes: %s", migrateVolumeKey, out)
	}
}

func removeContainers(t *testing.T, project string) {
	t.Helper()
	out := run(t, "", "docker", "ps", "-a", "-q", "--filter", "label=com.docker.compose.project="+project)
	for _, id := range strings.Fields(out) {
		_ = exec.Command("docker", "rm", "-f", "-v", id).Run()
	}
}

func removeVolume(t *testing.T, name string) {
	t.Helper()
	_ = exec.Command("docker", "volume", "rm", "-f", name).Run()
}

// createSourceVolume makes a volume the way an older Compose project would have,
// and writes one file into it.
func createSourceVolume(t *testing.T, project, file, content string) string {
	t.Helper()
	name := project + "_" + migrateVolumeKey
	run(t, "", "docker", "volume", "create",
		"--label", "com.docker.compose.project="+project,
		"--label", "com.docker.compose.volume="+migrateVolumeKey,
		name)
	t.Cleanup(func() { removeVolume(t, name) })
	dir := filepath.Dir(file)
	run(t, content, "docker", "run", "--rm", "-i",
		"-v", name+":/v",
		"--entrypoint", "sh",
		migrateCopyImage,
		"-c", fmt.Sprintf("mkdir -p /v/%s && cat > /v/%s && chown -R 1000:1000 /v/%s", dir, file, dir))
	return name
}

func readFile(t *testing.T, volume, file string) string {
	t.Helper()
	return run(t, "", "docker", "run", "--rm",
		"-v", volume+":/v:ro",
		"--entrypoint", "sh",
		migrateCopyImage,
		"-c", fmt.Sprintf("stat -c '%%u:%%g' /v/%s && cat /v/%s", file, file))
}

func volumeExists(volume string) bool {
	return exec.Command("docker", "volume", "inspect", volume).Run() == nil
}

func TestMigrateVolumesCopiesPlugins(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	src := createSourceVolume(t, "legacy-auth", "plugin/fiftyone.yml", "name: my-plugin")

	out, code := e.migrate()

	require.Equal(t, 0, code, out)
	require.Equal(t, "1000:1000\nname: my-plugin", readFile(t, e.destVolume(), "plugin/fiftyone.yml"))
	require.True(t, volumeExists(src), "source volume should still exist")
	require.Contains(t, out, "docker volume rm "+src)
}

func TestMigrateVolumesSkipsMissing(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")

	out, code := e.migrate()

	require.Equal(t, 0, code, out)
	require.Contains(t, out, "skipped")
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")
}

func TestMigrateVolumesSkipsWithoutPluginsOverlay(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml")
	createSourceVolume(t, "legacy-auth", "plugin/fiftyone.yml", "name: my-plugin")

	out, code := e.migrate()

	require.Equal(t, 0, code, out)
	require.Contains(t, out, "skipped")
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")
}

func TestMigrateVolumesStopsOnTwoSources(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	legacy := "legacy-auth"
	internal := "internal-auth"
	createSourceVolume(t, legacy, "plugin/fiftyone.yml", "from: legacy")
	createSourceVolume(t, internal, "plugin/fiftyone.yml", "from: internal")

	out, code := e.migrate()

	require.Equal(t, 1, code, out)
	require.Contains(t, out, legacy)
	require.Contains(t, out, internal)
	require.Contains(t, out, "--from-project")
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")

	out, code = e.migrate("--from-project", internal)

	require.Equal(t, 0, code, out)
	require.Equal(t, "1000:1000\nfrom: internal", readFile(t, e.destVolume(), "plugin/fiftyone.yml"))
}

func TestMigrateVolumesStopsWhileOldProjectRuns(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	old := "legacy-auth"
	createSourceVolume(t, old, "plugin/fiftyone.yml", "name: my-plugin")
	container := run(t, "", "docker", "run", "-d",
		"--label", "com.docker.compose.project="+old,
		"--entrypoint", "sleep",
		migrateCopyImage, "300")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", container).Run() })

	out, code := e.migrate()

	require.Equal(t, 1, code, out)
	require.Contains(t, out, "docker compose -p "+old+" down")
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")
}

func TestMigrateVolumesStopsWhenImageMissing(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	createSourceVolume(t, "legacy-auth", "plugin/fiftyone.yml", "name: my-plugin")

	out, code := e.migrate("--image", "voxel51/does-not-exist:"+uniqueName("tag"))

	require.Equal(t, 1, code, out)
	require.Contains(t, out, "pull")
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")
}

func TestMigrateVolumesSkipsNonEmptyDestination(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	src := createSourceVolume(t, "legacy-auth", "plugin/fiftyone.yml", "version: 1")

	out, code := e.migrate()
	require.Equal(t, 0, code, out)

	run(t, "version: 2", "docker", "run", "--rm", "-i",
		"-v", src+":/v",
		"--entrypoint", "sh",
		migrateCopyImage,
		"-c", "cat > /v/plugin/fiftyone.yml")

	out, code = e.migrate()

	require.Equal(t, 0, code, out)
	require.Contains(t, out, "skipped")
	require.Equal(t, "1000:1000\nversion: 1", readFile(t, e.destVolume(), "plugin/fiftyone.yml"))

	out, code = e.migrate("--force")

	require.Equal(t, 0, code, out)
	require.Contains(t, readFile(t, e.destVolume(), "plugin/fiftyone.yml"), "version: 2")
}

func TestMigrateVolumesDestinationHasComposeLabels(t *testing.T) {
	probe := filepath.Join(t.TempDir(), "compose.probe.yaml")
	require.NoError(t, os.WriteFile(probe, []byte(migrateProbeCompose), 0o644))
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml", probe)
	createSourceVolume(t, "legacy-auth", "plugin/fiftyone.yml", "name: my-plugin")

	out, code := e.migrate()
	require.Equal(t, 0, code, out)

	args := append([]string{"compose"}, e.composeArgs...)
	args = append(args, "up", "--no-start", "--no-deps", "--pull", "never", "probe")
	cmd := exec.Command("docker", args...)
	cmd.Dir = migrateDockerDir
	upOut, err := cmd.CombinedOutput()
	require.NoError(t, err, string(upOut))
	t.Cleanup(func() {
		down := exec.Command("docker", append(append([]string{"compose"}, e.composeArgs...), "down", "-v")...)
		down.Dir = migrateDockerDir
		_ = down.Run()
	})

	require.NotContains(t, string(upOut), "already exists")
	require.NotContains(t, string(upOut), "external: true")
}

func TestMigrateVolumesFailsWhenVolumeListFails(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	createSourceVolume(t, "legacy-auth", "plugin/fiftyone.yml", "name: my-plugin")

	e.env = failingDockerPath(t, `[ "$arg" = "--volumes" ]`)

	out, code := e.migrate()

	require.Equal(t, 1, code, out)
	require.NotContains(t, out, "skipped")
	require.Contains(t, out, "config --volumes")
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")
}

// On a host that runs several deployments,
// a project that kept its own name already has its plugins,
// so the script should skip without looking at other projects.
func TestMigrateVolumesSkipsWhenDestinationHasFilesOnSharedHost(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	createSourceVolume(t, e.destProject, "plugin/fiftyone.yml", "from: mine")
	createSourceVolume(t, uniqueName("alice"), "plugin/fiftyone.yml", "from: alice")
	createSourceVolume(t, uniqueName("bob"), "plugin/fiftyone.yml", "from: bob")

	out, code := e.migrate()

	require.Equal(t, 0, code, out)
	require.Contains(t, out, "already has files")
	require.Equal(t, "1000:1000\nfrom: mine", readFile(t, e.destVolume(), "plugin/fiftyone.yml"))
}

// A plugins volume from a project with any other name
// may belong to another deployment on the same host,
// so the script only copies it when asked by name.
func TestMigrateVolumesIgnoresOtherProjectNamesWithoutFromProject(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	teammate := uniqueName("teammate")
	createSourceVolume(t, teammate, "plugin/fiftyone.yml", "from: teammate")

	out, code := e.migrate()

	require.Equal(t, 0, code, out)
	require.Contains(t, out, "skipped")
	require.Contains(t, out, "--from-project")
	require.Contains(t, out, teammate)
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")

	out, code = e.migrate("--from-project", teammate)

	require.Equal(t, 0, code, out)
	require.Equal(t, "1000:1000\nfrom: teammate", readFile(t, e.destVolume(), "plugin/fiftyone.yml"))
}

// failingDockerPath returns a PATH setting whose docker wrapper fails
// when the shell condition matches the current argument ($arg)
// or the whole command line ($*),
// and passes every other command to the real docker.
func failingDockerPath(t *testing.T, condition string) []string {
	t.Helper()
	realDocker, err := exec.LookPath("docker")
	require.NoError(t, err)
	shimDir := t.TempDir()
	shim := fmt.Sprintf(`#!/bin/sh
for arg in "$@"; do
  if %s; then
    echo "simulated failure" >&2
    exit 1
  fi
done
exec %q "$@"
`, condition, realDocker)
	require.NoError(t, os.WriteFile(filepath.Join(shimDir, "docker"), []byte(shim), 0o755))
	return []string{"PATH=" + shimDir + string(os.PathListSeparator) + os.Getenv("PATH")}
}

// If docker can't list volumes, for example because of socket permissions,
// the script must fail instead of reporting that there was nothing to copy.
func TestMigrateVolumesFailsWhenVolumeLsFails(t *testing.T) {
	e := newMigrateEnv(t, "compose.yaml", "compose.plugins.yaml")
	createSourceVolume(t, "legacy-auth", "plugin/fiftyone.yml", "name: my-plugin")
	e.env = failingDockerPath(t, `[ "$1 $2" = "volume ls" ]`)

	out, code := e.migrate()

	require.Equal(t, 1, code, out)
	require.NotContains(t, out, "skipped")
	require.Contains(t, out, "docker volume ls")
	require.False(t, volumeExists(e.destVolume()), "no destination volume should be created")
}
