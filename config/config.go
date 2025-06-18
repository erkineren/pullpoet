package config

import (
	"fmt"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Repo            string
	Source          string
	Target          string
	Description     string
	Provider        string
	APIKey          string
	ProviderBaseURL string
	Model           string
	SystemPrompt    string
	// ClickUp integration fields
	ClickUpPAT    string
	ClickUpTaskID string
}

// GetProviderBaseURL returns the appropriate base URL for the provider
func (cfg *Config) GetProviderBaseURL() string {
	provider := strings.ToLower(cfg.Provider)

	// If custom URL is provided, use it
	if cfg.ProviderBaseURL != "" {
		return cfg.ProviderBaseURL
	}

	// Return default URLs for each provider
	switch provider {
	case "openai":
		return "https://api.openai.com"
	case "gemini":
		return "https://generativelanguage.googleapis.com"
	case "ollama":
		return "http://localhost:11434"
	case "openwebui":
		return "http://localhost:3000"
	default:
		return ""
	}
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
	if provider != "openai" && provider != "ollama" && provider != "gemini" && provider != "openwebui" {
		return fmt.Errorf("provider must be 'openai', 'ollama', 'gemini', or 'openwebui'")
	}

	if (provider == "openai" || provider == "gemini") && cfg.APIKey == "" {
		return fmt.Errorf("API key is required when using OpenAI or Gemini provider")
	}

	// For Ollama and OpenWebUI, either custom URL or default URL must be available
	if provider == "ollama" || provider == "openwebui" {
		baseURL := cfg.GetProviderBaseURL()
		if baseURL == "" {
			return fmt.Errorf("provider base URL is required when using %s provider", provider)
		}
	}

	// ClickUp validation: both PAT and task ID must be provided together
	if (cfg.ClickUpPAT != "" && cfg.ClickUpTaskID == "") || (cfg.ClickUpPAT == "" && cfg.ClickUpTaskID != "") {
		return fmt.Errorf("both ClickUp PAT and task ID must be provided together")
	}

	return nil
}
