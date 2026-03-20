package aliwepaystat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestImportResultJSON(t *testing.T) {
	result := ImportResult{
		FilesProcessed: 3,
		Imported:       100,
		Skipped:        5,
		Errors:         1,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded ImportResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.FilesProcessed != 3 {
		t.Errorf("FilesProcessed: got %d, want 3", decoded.FilesProcessed)
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
		FilesProcessed: 1,
		Imported:       2,
		Skipped:        3,
		Errors:         4,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]int
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedKeys := []string{"files_processed", "imported", "skipped", "errors"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestImportCsvToDBWithResultEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	existing := make(map[string]struct{})
	result, err := ImportCsvToDBWithResult(tmpDir, nil, existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FilesProcessed != 0 {
		t.Errorf("FilesProcessed: got %d, want 0", result.FilesProcessed)
	}
	if result.Imported != 0 {
		t.Errorf("Imported: got %d, want 0", result.Imported)
	}
}

func TestImportCsvToDBWithResultBadDir(t *testing.T) {
	existing := make(map[string]struct{})
	_, err := ImportCsvToDBWithResult("/nonexistent/dir/path", nil, existing)
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestImportFileToDBWithResultUnknownFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "unknown.csv")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	existing := make(map[string]struct{})
	_, err := ImportFileToDBWithResult(filePath, nil, existing)
	if err == nil {
		t.Fatal("expected error for unknown file pattern")
	}
}

func TestImportFileToDBWithResultNonexistent(t *testing.T) {
	existing := make(map[string]struct{})
	_, err := ImportFileToDBWithResult("/nonexistent/alipay.csv", nil, existing)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestImportCsvToDBWithResultSkipsNonCsv(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a non-CSV file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	existing := make(map[string]struct{})
	result, err := ImportCsvToDBWithResult(tmpDir, nil, existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FilesProcessed != 0 {
		t.Errorf("FilesProcessed: got %d, want 0", result.FilesProcessed)
	}
}
