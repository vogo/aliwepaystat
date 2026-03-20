package aliwepaystat

import (
	"encoding/json"
	"testing"
)

func TestImportResultJSON(t *testing.T) {
	result := ImportResult{
		Imported: 100,
		Skipped:  5,
		Errors:   1,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded ImportResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Imported != 100 {
		t.Errorf("Imported: got %d, want 100", decoded.Imported)
	}
	if decoded.Skipped != 5 {
		t.Errorf("Skipped: got %d, want 5", decoded.Skipped)
	}
	if decoded.Errors != 1 {
		t.Errorf("Errors: got %d, want 1", decoded.Errors)
	}
}

func TestImportResultJSONKeys(t *testing.T) {
	result := ImportResult{
		Imported: 2,
		Skipped:  3,
		Errors:   4,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]int
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedKeys := []string{"imported", "skipped", "errors"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}

	// 确保不再有 files_processed 字段
	if _, ok := m["files_processed"]; ok {
		t.Error("unexpected JSON key \"files_processed\"")
	}
}

func TestParserForPlatform(t *testing.T) {
	tests := []struct {
		platform string
		wantErr  bool
	}{
		{"alipay", false},
		{"wechat", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		parser, err := ParserForPlatform(tt.platform)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParserForPlatform(%q): expected error, got nil", tt.platform)
			}
			if parser != nil {
				t.Errorf("ParserForPlatform(%q): expected nil parser on error", tt.platform)
			}
		} else {
			if err != nil {
				t.Errorf("ParserForPlatform(%q): unexpected error: %v", tt.platform, err)
			}
			if parser == nil {
				t.Errorf("ParserForPlatform(%q): expected non-nil parser", tt.platform)
			}
		}
	}
}

func TestImportFileToDBWithResultInvalidPlatform(t *testing.T) {
	existing := make(map[string]struct{})
	_, err := ImportFileToDBWithResult("/some/file.csv", "unknown", nil, existing)
	if err == nil {
		t.Fatal("expected error for invalid platform")
	}
}

func TestImportFileToDBWithResultNonexistent(t *testing.T) {
	existing := make(map[string]struct{})
	_, err := ImportFileToDBWithResult("/nonexistent/file.csv", "alipay", nil, existing)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
