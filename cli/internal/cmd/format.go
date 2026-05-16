package cmd

import (
	"fmt"
	"strings"
)

// formatRow builds a tab-separated, newline-terminated table row. The first
// argument is rendered as an ID; the rest as plain string columns.
func formatRow(id uint, cols ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d", id)
	for _, c := range cols {
		b.WriteByte('\t')
		b.WriteString(c)
	}
	b.WriteByte('\n')
	return b.String()
}
