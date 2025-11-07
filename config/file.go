package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the structure of .pullpoet.yml
type FileConfig struct {
	// Git Configuration
	Repo   string `yaml:"repo,omitempty"`
	Source string `yaml:"source,omitempty"`
	Target string `yaml:"target,omitempty"`

	// AI Provider Configuration
	Provider        string `yaml:"provider,omitempty"`
	Model           string `yaml:"model,omitempty"`
	APIKey          string `yaml:"api_key,omitempty"`
	ProviderBaseURL string `yaml:"provider_base_url,omitempty"`

	// General Settings
	SystemPrompt string `yaml:"system_prompt,omitempty"`
	Language     string `yaml:"language,omitempty"`
	FastMode     bool   `yaml:"fast_mode,omitempty"`
	Output       string `yaml:"output,omitempty"`

	// Integrations
	ClickUp *ClickUpConfig `yaml:"clickup,omitempty"`
	Jira    *JiraConfig    `yaml:"jira,omitempty"`

	// UI Settings
	UI *UIConfig `yaml:"ui,omitempty"`

	// Custom fields
	CustomFields map[string]interface{} `yaml:",inline"`
}

// ClickUpConfig holds ClickUp-specific configuration
type ClickUpConfig struct {
	PAT string `yaml:"pat,omitempty"`
}

// JiraConfig holds Jira-specific configuration
type JiraConfig struct {
	BaseURL  string `yaml:"base_url,omitempty"`
	Username string `yaml:"username,omitempty"`
	APIToken string `yaml:"api_token,omitempty"`
}

// UIConfig holds UI-specific configuration
type UIConfig struct {
	Colors       bool   `yaml:"colors,omitempty"`
	ProgressBars bool   `yaml:"progress_bars,omitempty"`
	Emoji        bool   `yaml:"emoji,omitempty"`
	Verbose      bool   `yaml:"verbose,omitempty"`
	Theme        string `yaml:"theme,omitempty"` // light, dark, auto
}

// DefaultUIConfig returns default UI configuration
func DefaultUIConfig() *UIConfig {
	return &UIConfig{
		Colors:       true,
		ProgressBars: true,
		Emoji:        true,
		Verbose:      false,
		Theme:        "auto",
	}
}

// LoadConfigFile loads configuration from .pullpoet.yml file
// Searches in current directory and parent directories up to home
func LoadConfigFile() (*FileConfig, error) {
	// Try to find .pullpoet.yml in current directory and parents
	configPath, err := findConfigFile()
	if err != nil {
		// No config file found, return empty config (not an error)
		return &FileConfig{UI: DefaultUIConfig()}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config FileConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Set defaults for UI if not specified
	if config.UI == nil {
		config.UI = DefaultUIConfig()
	} else {
		// Fill in missing UI fields with defaults
		defaults := DefaultUIConfig()
		if config.UI.Theme == "" {
			config.UI.Theme = defaults.Theme
		}
	}

	// Expand environment variables in config
	expandEnvVars(&config)

	return &config, nil
}

// findConfigFile searches for .pullpoet.yml in current directory and parents
func findConfigFile() (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Get home directory as upper limit
	home, err := os.UserHomeDir()
	if err != nil {
		home = "" // If we can't get home, just search up to root
	}

	// Search upwards from current directory
	dir := cwd
	for {
		configPath := filepath.Join(dir, ".pullpoet.yml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Also check for .pullpoet.yaml variant
		configPath = filepath.Join(dir, ".pullpoet.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Check if we've reached home or root
		parent := filepath.Dir(dir)
		if parent == dir || (home != "" && dir == home) {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no .pullpoet.yml found")
}

// expandEnvVars expands environment variables in config values
func expandEnvVars(config *FileConfig) {
	config.APIKey = os.ExpandEnv(config.APIKey)
	config.ProviderBaseURL = os.ExpandEnv(config.ProviderBaseURL)
	config.SystemPrompt = os.ExpandEnv(config.SystemPrompt)

	if config.ClickUp != nil {
		config.ClickUp.PAT = os.ExpandEnv(config.ClickUp.PAT)
	}

	if config.Jira != nil {
		config.Jira.BaseURL = os.ExpandEnv(config.Jira.BaseURL)
		config.Jira.Username = os.ExpandEnv(config.Jira.Username)
		config.Jira.APIToken = os.ExpandEnv(config.Jira.APIToken)
	}
}

// MergeWithConfig merges FileConfig into runtime Config
// Priority: CLI flags (already set) > FileConfig > Environment variables (handled elsewhere)
func (fc *FileConfig) MergeWithConfig(cfg *Config) {
	// Git configuration
	if cfg.Repo == "" && fc.Repo != "" {
		cfg.Repo = fc.Repo
	}
	if cfg.Source == "" && fc.Source != "" {
		cfg.Source = fc.Source
	}
	if cfg.Target == "" && fc.Target != "" {
		cfg.Target = fc.Target
	}

	// AI Provider configuration
	if cfg.Provider == "" && fc.Provider != "" {
		cfg.Provider = fc.Provider
	}
	if cfg.Model == "" && fc.Model != "" {
		cfg.Model = fc.Model
	}
	if cfg.APIKey == "" && fc.APIKey != "" {
		cfg.APIKey = fc.APIKey
	}
	if cfg.ProviderBaseURL == "" && fc.ProviderBaseURL != "" {
		cfg.ProviderBaseURL = fc.ProviderBaseURL
	}

	// General settings
	if cfg.SystemPrompt == "" && fc.SystemPrompt != "" {
		cfg.SystemPrompt = fc.SystemPrompt
	}
	if cfg.Language == "" && fc.Language != "" {
		cfg.Language = fc.Language
	}

	// ClickUp config
	if cfg.ClickUpPAT == "" && fc.ClickUp != nil && fc.ClickUp.PAT != "" {
		cfg.ClickUpPAT = fc.ClickUp.PAT
	}

	// Jira config
	if cfg.JiraBaseURL == "" && fc.Jira != nil && fc.Jira.BaseURL != "" {
		cfg.JiraBaseURL = fc.Jira.BaseURL
	}
	if cfg.JiraUsername == "" && fc.Jira != nil && fc.Jira.Username != "" {
		cfg.JiraUsername = fc.Jira.Username
	}
	if cfg.JiraAPIToken == "" && fc.Jira != nil && fc.Jira.APIToken != "" {
		cfg.JiraAPIToken = fc.Jira.APIToken
	}
}

// GenerateExampleConfig generates an example .pullpoet.yml file
func GenerateExampleConfig() string {
	return `# PullPoet Configuration File
# This file configures default values for PullPoet
# Priority: CLI flags > .pullpoet.yml > Environment variables > Defaults

# Git Configuration (optional - auto-detected if not specified)
# repo: https://github.com/your-username/your-repo.git  # Git repository URL
# source: feature/branch-name  # Source branch (default: current branch)
# target: main  # Target branch (default: repository's default branch)

# AI Provider Configuration
provider: openai  # openai, ollama, gemini, openwebui
model: gpt-4     # AI model to use
api_key: ${PULLPOET_API_KEY}  # Use environment variable
# provider_base_url: http://localhost:11434  # For Ollama/OpenWebUI

# General Settings
language: en  # Language for generated content (en, tr, es, fr, de, etc.)
# fast_mode: true  # Use fast native git commands for large repos
# output: pr-description.md  # Save output to file
# system_prompt: /path/to/custom-prompt.md  # Custom system prompt

# ClickUp Integration
clickup:
  pat: ${PULLPOET_CLICKUP_PAT}  # ClickUp Personal Access Token

# Jira Integration
jira:
  base_url: ${PULLPOET_JIRA_BASE_URL}  # e.g., https://company.atlassian.net
  username: ${PULLPOET_JIRA_USERNAME}  # Your Jira email
  api_token: ${PULLPOET_JIRA_API_TOKEN}  # Jira API token

# UI Configuration
ui:
  colors: true  # Enable colored output
  progress_bars: true  # Show progress bars
  emoji: true  # Use emoji in output
  verbose: false  # Show detailed logs
  theme: auto  # auto, light, dark

# Advanced: Custom fields (preserved but not used by core)
# custom_field: value
`
}
