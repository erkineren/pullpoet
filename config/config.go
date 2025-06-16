package config

import (
	"fmt"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Repo        string
	Source      string
	Target      string
	Description string
	Provider    string
	APIKey      string
	OllamaURL   string
	Model       string
}

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	if cfg.Repo == "" {
		return fmt.Errorf("repository URL is required")
	}

	if cfg.Source == "" {
		return fmt.Errorf("source branch is required")
	}

	if cfg.Target == "" {
		return fmt.Errorf("target branch is required")
	}

	if cfg.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if cfg.Model == "" {
		return fmt.Errorf("model is required")
	}

	provider := strings.ToLower(cfg.Provider)
	if provider != "openai" && provider != "ollama" {
		return fmt.Errorf("provider must be either 'openai' or 'ollama'")
	}

	if provider == "openai" && cfg.APIKey == "" {
		return fmt.Errorf("API key is required when using OpenAI provider")
	}

	if provider == "ollama" && cfg.OllamaURL == "" {
		return fmt.Errorf("Ollama URL is required when using Ollama provider")
	}

	return nil
}
