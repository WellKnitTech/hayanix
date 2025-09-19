package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	LogFile      string   `json:"log_file"`
	LogType      string   `json:"log_type"`
	RulesDir     string   `json:"rules_dir"`
	OutputFormat string   `json:"output_format"`
	RuleSources  []string `json:"rule_sources"`
	LastUpdated  string   `json:"last_updated"`
}

const (
	ConfigFileName = "hayanix.json"
	ConfigDir      = ".hayanix"
)

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, ConfigFileName), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{
			LogType:      "syslog",
			LogFile:      "/var/log/messages",
			RulesDir:     "./rules",
			OutputFormat: "table",
			RuleSources:  []string{},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) GetAnalysisCommand() string {
	return fmt.Sprintf("./hayanix analyze --target %s --file %s --rules %s --output %s",
		c.LogType, c.LogFile, c.RulesDir, c.OutputFormat)
}
