package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppConfig holds global application settings from ~/.aliwepaystat.conf
type AppConfig struct {
	DBPath string // "db" key - path to SQLite database
	Dir    string // "dir" key - default transaction file directory
}

// appContext holds shared state for all subcommands
type appContext struct {
	configPath string
	appConfig  *AppConfig
	db         *sql.DB
	jsonOutput bool
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".aliwepaystat.conf"
	}
	return filepath.Join(home, ".aliwepaystat.conf")
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".aliwepaystat"
	}
	return filepath.Join(home, ".aliwepaystat")
}

func loadAppConfig(path string) (*AppConfig, error) {
	cfg := &AppConfig{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, value, ok := strings.Cut(line, "="); ok {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			switch key {
			case "db":
				cfg.DBPath = value
			case "dir":
				cfg.Dir = value
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func saveAppConfig(path string, cfg *AppConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	content := fmt.Sprintf("# aliwepaystat global configuration\ndb = %s\ndir = %s\n", cfg.DBPath, cfg.Dir)
	return os.WriteFile(path, []byte(content), 0600)
}

func ensureAppConfig(path string) (*AppConfig, error) {
	cfg, err := loadAppConfig(path)
	if err == nil {
		// Fill in defaults for any missing values
		if cfg.DBPath == "" {
			cfg.DBPath = filepath.Join(defaultDataDir(), "data.db")
		}
		if cfg.Dir == "" {
			cfg.Dir = filepath.Join(defaultDataDir(), "files")
		}
		return cfg, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	// File does not exist, create with defaults
	dataDir := defaultDataDir()
	cfg = &AppConfig{
		DBPath: filepath.Join(dataDir, "data.db"),
		Dir:    filepath.Join(dataDir, "files"),
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := saveAppConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return cfg, nil
}
