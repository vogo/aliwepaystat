package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// printJSON marshals the value to JSON and writes it to the writer.
func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printTable prints a simple column-aligned text table.
func printTable(w io.Writer, headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(w, "  ")
		}
		_, _ = fmt.Fprintf(w, "%-*s", widths[i], h)
	}
	_, _ = fmt.Fprintln(w)

	// Print separator
	for i, width := range widths {
		if i > 0 {
			_, _ = fmt.Fprint(w, "  ")
		}
		_, _ = fmt.Fprint(w, strings.Repeat("-", width))
	}
	_, _ = fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i := range headers {
			if i > 0 {
				_, _ = fmt.Fprint(w, "  ")
			}
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			_, _ = fmt.Fprintf(w, "%-*s", widths[i], cell)
		}
		_, _ = fmt.Fprintln(w)
	}
}

// printKeyValue prints key-value pairs in "key: value" format.
func printKeyValue(w io.Writer, pairs [][2]string) {
	maxKeyLen := 0
	for _, p := range pairs {
		if len(p[0]) > maxKeyLen {
			maxKeyLen = len(p[0])
		}
	}
	for _, p := range pairs {
		_, _ = fmt.Fprintf(w, "%-*s : %s\n", maxKeyLen, p[0], p[1])
	}
}
