package config

import (
	"encoding/json"
	"fmt"
	"os"

	"eng_club/database"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
	Telegram TelegramConfig `json:"telegram"`
	Server   ServerConfig   `json:"server"`
	Club     ClubConfig     `json:"club"`
}

// DatabaseConfig holds MySQL database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Charset  string `json:"charset"`
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken     string `json:"bot_token"`
	NotifyChatID string `json:"notify_chat_id"`
	Enabled      bool   `json:"enabled"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// ClubConfig holds club-specific settings
type ClubConfig struct {
	DefaultSessionDuration int    `json:"default_session_duration_minutes"`
	MaxMembersPerTable     int    `json:"max_members_per_table"`
	NotificationTimings    []int  `json:"notification_timings_minutes_before_end"`
	VenueName              string `json:"venue_name"`
}

// LoadAppConfig reads configuration only from a JSON file. The file must exist.
// No built-in defaults, .env, or process environment variables are applied.
// Pass empty path to use "config.json" in the working directory.
func LoadAppConfig(path string) (*Config, error) {
	if path == "" {
		path = "config.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &cfg, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(cfg *Config, configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if c.Database.Charset == "" {
		return fmt.Errorf("database charset is required")
	}

	if c.Telegram.Enabled && c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is required when telegram is enabled")
	}
	if c.Telegram.Enabled && c.Telegram.BotToken == "YOUR_TELEGRAM_BOT_TOKEN_HERE" {
		return fmt.Errorf("please set a real telegram bot token")
	}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}

	if c.Club.DefaultSessionDuration <= 0 {
		return fmt.Errorf("default session duration must be positive")
	}
	if c.Club.MaxMembersPerTable <= 1 {
		return fmt.Errorf("max members per table must be at least 2")
	}

	return nil
}

// MySQL returns database connection parameters for the MySQL driver.
func (c *Config) MySQL() database.MySQLConfig {
	return database.MySQLConfig{
		Host:     c.Database.Host,
		Port:     c.Database.Port,
		Database: c.Database.Database,
		Username: c.Database.Username,
		Password: c.Database.Password,
		Charset:  c.Database.Charset,
	}
}
