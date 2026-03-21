package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value", "foo": "bar"}

	if err := printJSON(&buf, data); err != nil {
		t.Fatalf("printJSON failed: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if decoded["key"] != "value" {
		t.Errorf("key: got %q, want %q", decoded["key"], "value")
	}
	if decoded["foo"] != "bar" {
		t.Errorf("foo: got %q, want %q", decoded["foo"], "bar")
	}
}

func TestPrintJSONArray(t *testing.T) {
	var buf bytes.Buffer
	data := []string{"a", "b", "c"}

	if err := printJSON(&buf, data); err != nil {
		t.Fatalf("printJSON failed: %v", err)
	}

	var decoded []string
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(decoded) != 3 {
		t.Errorf("expected 3 elements, got %d", len(decoded))
	}
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"db", "/path/to/db"},
		{"dir", "/path/to/dir"},
	}

	printTable(&buf, headers, rows)

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	if len(lines) != 4 { // header + separator + 2 rows
		t.Fatalf("expected 4 lines, got %d: %q", len(lines), output)
	}

	// Header should contain column names
	if !strings.Contains(lines[0], "Name") || !strings.Contains(lines[0], "Value") {
		t.Errorf("header missing column names: %q", lines[0])
	}

	// Separator line should contain dashes
	if !strings.Contains(lines[1], "---") {
		t.Errorf("separator should contain dashes: %q", lines[1])
	}

	// Data rows
	if !strings.Contains(lines[2], "db") || !strings.Contains(lines[2], "/path/to/db") {
		t.Errorf("row 1 missing data: %q", lines[2])
	}
}

func TestPrintTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	printTable(&buf, []string{}, nil)
	if buf.Len() != 0 {
		t.Error("expected no output for empty headers")
	}
}

func TestPrintKeyValue(t *testing.T) {
	var buf bytes.Buffer
	pairs := [][2]string{
		{"db", "/path/to/db"},
		{"dir", "/path/to/dir"},
	}

	printKeyValue(&buf, pairs)

	output := buf.String()
	if !strings.Contains(output, "db") || !strings.Contains(output, "/path/to/db") {
		t.Errorf("missing db entry in output: %q", output)
	}
	if !strings.Contains(output, "dir") || !strings.Contains(output, "/path/to/dir") {
		t.Errorf("missing dir entry in output: %q", output)
	}
	// Should have colon separator
	if !strings.Contains(output, " : ") {
		t.Errorf("missing colon separator in output: %q", output)
	}
}

func TestPrintKeyValueAlignment(t *testing.T) {
	var buf bytes.Buffer
	pairs := [][2]string{
		{"a", "1"},
		{"longer_key", "2"},
	}

	printKeyValue(&buf, pairs)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Both colons should be at the same position
	pos1 := strings.Index(lines[0], " : ")
	pos2 := strings.Index(lines[1], " : ")
	if pos1 != pos2 {
		t.Errorf("colons not aligned: pos1=%d, pos2=%d", pos1, pos2)
	}
}
