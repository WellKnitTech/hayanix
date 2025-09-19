package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
)

type RuleManager struct {
	rulesDir string
}

type RuleSource struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Branch      string `yaml:"branch"`
	Description string `yaml:"description"`
	Enabled     bool   `yaml:"enabled"`
}

type RuleConfig struct {
	Sources []RuleSource `yaml:"sources"`
}

func NewRuleManager(rulesDir string) *RuleManager {
	return &RuleManager{
		rulesDir: rulesDir,
	}
}

func (rm *RuleManager) Initialize() error {
	// Create rules directory structure
	dirs := []string{
		rm.rulesDir,
		filepath.Join(rm.rulesDir, "linux"),
		filepath.Join(rm.rulesDir, "linux", "syslog"),
		filepath.Join(rm.rulesDir, "linux", "journald"),
		filepath.Join(rm.rulesDir, "linux", "auditd"),
		filepath.Join(rm.rulesDir, "external"),
		filepath.Join(rm.rulesDir, "external", "chopchopgo"),
		filepath.Join(rm.rulesDir, "external", "sigmahq"),
		filepath.Join(rm.rulesDir, "external", "custom"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create default rule sources configuration
	return rm.createDefaultConfig()
}

func (rm *RuleManager) createDefaultConfig() error {
	configPath := filepath.Join(rm.rulesDir, "sources.yml")
	
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config already exists
	}

	config := RuleConfig{
		Sources: []RuleSource{
			{
				Name:        "ChopChopGo",
				URL:         "https://github.com/M00NLIG7/ChopChopGo",
				Branch:      "master",
				Description: "ChopChopGo Linux forensics rules",
				Enabled:     true,
			},
			{
				Name:        "SigmaHQ",
				URL:         "https://github.com/SigmaHQ/sigma",
				Branch:      "master",
				Description: "Official Sigma rules repository",
				Enabled:     true,
			},
		},
	}

	return rm.saveConfig(config)
}

func (rm *RuleManager) saveConfig(config RuleConfig) error {
	configPath := filepath.Join(rm.rulesDir, "sources.yml")
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func (rm *RuleManager) loadConfig() (RuleConfig, error) {
	configPath := filepath.Join(rm.rulesDir, "sources.yml")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return RuleConfig{}, fmt.Errorf("failed to read config: %w", err)
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return RuleConfig{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func (rm *RuleManager) DownloadRules(sourceName string) error {
	config, err := rm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var source *RuleSource
	for _, s := range config.Sources {
		if s.Name == sourceName {
			source = &s
			break
		}
	}

	if source == nil {
		return fmt.Errorf("source %s not found", sourceName)
	}

	if !source.Enabled {
		return fmt.Errorf("source %s is disabled", sourceName)
	}

	return rm.downloadSource(source)
}

func (rm *RuleManager) downloadSource(source *RuleSource) error {
	// Create target directory
	targetDir := filepath.Join(rm.rulesDir, "external", strings.ToLower(source.Name))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Download rules based on source type
	switch strings.ToLower(source.Name) {
	case "chopchopgo":
		return rm.downloadChopChopGoRules(targetDir, source)
	case "sigmahq":
		return rm.downloadSigmaHQRules(targetDir, source)
	default:
		return fmt.Errorf("unsupported source: %s", source.Name)
	}
}

func (rm *RuleManager) downloadChopChopGoRules(targetDir string, source *RuleSource) error {
	// ChopChopGo specific rule paths
	rulePaths := []string{
		"rules/linux/builtin/syslog",
		"rules/linux/builtin/journald", 
		"rules/linux/builtin/auditd",
	}

	downloader := NewGitHubDownloader()
	return downloader.DownloadRepositoryRules(source.URL, source.Branch, targetDir, rulePaths)
}

func (rm *RuleManager) downloadSigmaHQRules(targetDir string, source *RuleSource) error {
	// SigmaHQ specific rule paths for Linux
	rulePaths := []string{
		"rules/linux",
		"rules/linux/auditd",
		"rules/linux/systemd",
		"rules/linux/rsyslog",
	}

	downloader := NewGitHubDownloader()
	return downloader.DownloadRepositoryRules(source.URL, source.Branch, targetDir, rulePaths)
}


func (rm *RuleManager) ListSources() ([]RuleSource, error) {
	config, err := rm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return config.Sources, nil
}

func (rm *RuleManager) AddSource(source RuleSource) error {
	config, err := rm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if source already exists
	for _, s := range config.Sources {
		if s.Name == source.Name {
			return fmt.Errorf("source %s already exists", source.Name)
		}
	}

	config.Sources = append(config.Sources, source)
	return rm.saveConfig(config)
}

func (rm *RuleManager) RemoveSource(sourceName string) error {
	config, err := rm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var newSources []RuleSource
	for _, s := range config.Sources {
		if s.Name != sourceName {
			newSources = append(newSources, s)
		}
	}

	config.Sources = newSources
	return rm.saveConfig(config)
}

func (rm *RuleManager) EnableSource(sourceName string) error {
	return rm.toggleSource(sourceName, true)
}

func (rm *RuleManager) DisableSource(sourceName string) error {
	return rm.toggleSource(sourceName, false)
}

func (rm *RuleManager) toggleSource(sourceName string, enabled bool) error {
	config, err := rm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for i, s := range config.Sources {
		if s.Name == sourceName {
			config.Sources[i].Enabled = enabled
			return rm.saveConfig(config)
		}
	}

	return fmt.Errorf("source %s not found", sourceName)
}
