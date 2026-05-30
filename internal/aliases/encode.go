package aliases

import (
	"fmt"
	"io"
	"strings"
)

func encodeFile(f File, w io.Writer) error {
	var b strings.Builder
	fmt.Fprintf(&b, "schema_version: %d\n", SchemaVersion)
	fmt.Fprintln(&b, "aliases:")
	for _, e := range f.Entries {
		fmt.Fprintf(&b, "  - name: %s\n", quoteYAML(e.Name))
		fmt.Fprintf(&b, "    value: %s\n", quoteYAML(e.Value))
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func quoteYAML(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, ":#\n\"'") || strings.Contains(s, " ") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}
