package unit

import "strings"

// splitYAMLDocs splits a multi-doc YAML string on `---` separators.
// Empty/whitespace-only docs are dropped.
func splitYAMLDocs(s string) []string {
	out := []string{}
	for _, raw := range strings.Split(s, "\n---") {
		doc := strings.TrimSpace(raw)
		if doc == "" {
			continue
		}
		out = append(out, doc)
	}
	return out
}
