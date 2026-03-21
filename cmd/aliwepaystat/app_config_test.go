package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	path := defaultConfigPath()
	if path == "" {
		t.Fatal("defaultConfigPath returned empty string")
	}
	if filepath.Base(path) != ".aliwepaystat.conf" {
		t.Errorf("expected filename .aliwepaystat.conf, got %s", filepath.Base(path))
	}
}

func TestSaveAndLoadAppConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")

	cfg := &AppConfig{
		DBPath: "/tmp/test.db",
		Dir:    "/tmp/test-files",
	}

	// Save
	if err := saveAppConfig(configPath, cfg); err != nil {
		t.Fatalf("saveAppConfig failed: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}

	// Load
	loaded, err := loadAppConfig(configPath)
	if err != nil {
		t.Fatalf("loadAppConfig failed: %v", err)
	}

	if loaded.DBPath != cfg.DBPath {
		t.Errorf("DBPath: got %q, want %q", loaded.DBPath, cfg.DBPath)
	}
	if loaded.Dir != cfg.Dir {
		t.Errorf("Dir: got %q, want %q", loaded.Dir, cfg.Dir)
	}
}

func TestLoadAppConfigNotFound(t *testing.T) {
	_, err := loadAppConfig("/nonexistent/path/config.conf")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadAppConfigComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")

	content := "# comment line\ndb = /path/to/db\n# another comment\ndir = /path/to/dir\n"
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	cfg, err := loadAppConfig(configPath)
	if err != nil {
		t.Fatalf("loadAppConfig failed: %v", err)
	}

	if cfg.DBPath != "/path/to/db" {
		t.Errorf("DBPath: got %q, want %q", cfg.DBPath, "/path/to/db")
	}
	if cfg.Dir != "/path/to/dir" {
		t.Errorf("Dir: got %q, want %q", cfg.Dir, "/path/to/dir")
	}
}

func TestLoadAppConfigEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")

	if err := os.WriteFile(configPath, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadAppConfig(configPath)
	if err != nil {
		t.Fatalf("loadAppConfig failed: %v", err)
	}

	if cfg.DBPath != "" {
		t.Errorf("DBPath should be empty, got %q", cfg.DBPath)
	}
	if cfg.Dir != "" {
		t.Errorf("Dir should be empty, got %q", cfg.Dir)
	}
}

func TestEnsureAppConfigCreatesNew(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "test.conf")

	cfg, err := ensureAppConfig(configPath)
	if err != nil {
		t.Fatalf("ensureAppConfig failed: %v", err)
	}

	if cfg.DBPath == "" {
		t.Error("DBPath should not be empty")
	}
	if cfg.Dir == "" {
		t.Error("Dir should not be empty")
	}

	// File should have been created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load again should return same values
	cfg2, err := ensureAppConfig(configPath)
	if err != nil {
		t.Fatalf("ensureAppConfig (second call) failed: %v", err)
	}

	if cfg2.DBPath != cfg.DBPath {
		t.Errorf("DBPath mismatch: %q vs %q", cfg2.DBPath, cfg.DBPath)
	}
	if cfg2.Dir != cfg.Dir {
		t.Errorf("Dir mismatch: %q vs %q", cfg2.Dir, cfg.Dir)
	}
}

func TestEnsureAppConfigFillsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")

	// Write config with only db set
	if err := os.WriteFile(configPath, []byte("db = /custom/path.db\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := ensureAppConfig(configPath)
	if err != nil {
		t.Fatalf("ensureAppConfig failed: %v", err)
	}

	if cfg.DBPath != "/custom/path.db" {
		t.Errorf("DBPath: got %q, want %q", cfg.DBPath, "/custom/path.db")
	}
	// Dir should be filled with default
	if cfg.Dir == "" {
		t.Error("Dir should be filled with default value")
	}
}
