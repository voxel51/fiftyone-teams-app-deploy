//go:build kubeall || helm || unit || unitNotes
// +build kubeall helm unit unitNotes

package unit

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const authSecretWarning = "secret.fiftyone.fiftyoneAuthSecret is empty"

// renderNotes returns the NOTES section that `helm install` prints.
// `helm template` can't render NOTES.txt,
// so this uses a client-side dry run, which needs no cluster.
func renderNotes(t *testing.T, setValues ...string) string {
	t.Helper()
	chart, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	args := []string{"install", "notes-test", chart, "--dry-run=client"}
	for _, v := range setValues {
		args = append(args, "--set", v)
	}
	out, err := exec.Command("helm", args...).CombinedOutput()
	require.NoError(t, err, string(out))

	_, notes, found := strings.Cut(string(out), "NOTES:")
	require.True(t, found, "helm install output should have a NOTES section:\n%s", out)
	return notes
}

func TestNotesWarnsWhenAuthSecretIsEmpty(t *testing.T) {
	t.Parallel()

	require.Contains(t, renderNotes(t), authSecretWarning)
}

func TestNotesQuietWhenAuthSecretIsSet(t *testing.T) {
	t.Parallel()

	notes := renderNotes(t, "secret.fiftyone.fiftyoneAuthSecret=test-auth-secret")

	require.NotContains(t, notes, authSecretWarning)
}

func TestNotesQuietWhenChartDoesNotCreateSecret(t *testing.T) {
	t.Parallel()

	notes := renderNotes(t, "secret.create=false")

	require.NotContains(t, notes, authSecretWarning)
}
