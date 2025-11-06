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
	Language        string
	// ClickUp integration fields
	ClickUpPAT    string
	ClickUpTaskID string
	// Jira integration fields
	JiraBaseURL  string
	JiraUsername string
	JiraAPIToken string
	JiraTaskID   string
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

	if cfg.Provider == "" {
		return fmt.Errorf("provider is required (can be set via --provider flag or PULLPOET_PROVIDER environment variable)")
	}

	if cfg.Model == "" {
		return fmt.Errorf("model is required (can be set via --model flag or PULLPOET_MODEL environment variable)")
	}

	provider := strings.ToLower(cfg.Provider)
	if provider != "openai" && provider != "ollama" && provider != "gemini" && provider != "openwebui" {
		return fmt.Errorf("provider must be 'openai', 'ollama', 'gemini', or 'openwebui'")
	}

	if (provider == "openai" || provider == "gemini") && cfg.APIKey == "" {
		return fmt.Errorf("API key is required when using OpenAI or Gemini provider (can be set via --api-key flag or PULLPOET_API_KEY environment variable)")
	}

	// For Ollama and OpenWebUI, either custom URL or default URL must be available
	if provider == "ollama" || provider == "openwebui" {
		baseURL := cfg.GetProviderBaseURL()
		if baseURL == "" {
			return fmt.Errorf("provider base URL is required when using %s provider (can be set via --provider-base-url flag or PULLPOET_PROVIDER_BASE_URL environment variable)", provider)
		}
	}

	// ClickUp validation: task ID requires PAT, but PAT can exist without task ID
	if cfg.ClickUpTaskID != "" && cfg.ClickUpPAT == "" {
		return fmt.Errorf("ClickUp PAT is required when task ID is provided (PAT can be set via --clickup-pat flag or PULLPOET_CLICKUP_PAT environment variable)")
	}

	// Jira validation: task ID requires all credentials
	if cfg.JiraTaskID != "" {
		if cfg.JiraBaseURL == "" {
			return fmt.Errorf("Jira base URL is required when task ID is provided (can be set via --jira-base-url flag or PULLPOET_JIRA_BASE_URL environment variable)")
		}
		if cfg.JiraUsername == "" {
			return fmt.Errorf("Jira username is required when task ID is provided (can be set via --jira-username flag or PULLPOET_JIRA_USERNAME environment variable)")
		}
		if cfg.JiraAPIToken == "" {
			return fmt.Errorf("Jira API token is required when task ID is provided (can be set via --jira-api-token flag or PULLPOET_JIRA_API_TOKEN environment variable)")
		}
	}

	// Cannot use both ClickUp and Jira task IDs at the same time
	if cfg.ClickUpTaskID != "" && cfg.JiraTaskID != "" {
		return fmt.Errorf("cannot use both ClickUp and Jira task IDs at the same time - please provide only one")
	}

	return nil
}
