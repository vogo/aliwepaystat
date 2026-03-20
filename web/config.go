package web

import (
	"database/sql"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	db *sql.DB
}

// NewConfigManager 创建配置管理器
func NewConfigManager(db *sql.DB) *ConfigManager {
	return &ConfigManager{db: db}
}

// GetConfig 获取配置项
func (c *ConfigManager) GetConfig(key string) (string, error) {
	var value string
	err := c.db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	return value, err
}

// SetConfig 设置配置项
func (c *ConfigManager) SetConfig(key, value string) error {
	_, err := c.db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", key, value)
	return err
}

// GetAllConfig 获取所有配置项
func (c *ConfigManager) GetAllConfig() (map[string]string, error) {
	rows, err := c.db.Query("SELECT key, value FROM config")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	configs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		configs[key] = value
	}

	return configs, nil
}
