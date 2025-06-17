package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// OllamaClient implements the Client interface for Ollama
type OllamaClient struct {
	BaseURL string
	Auth    string
	Model   string
	Client  *http.Client
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(ollamaURL, model string) *OllamaClient {
	client := &OllamaClient{
		Model:  model,
		Client: &http.Client{},
	}

	// Parse URL to extract credentials
	parsedURL, err := url.Parse(ollamaURL)
	if err != nil {
		client.BaseURL = ollamaURL
		return client
	}

	// Extract credentials if present
	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		password, _ := parsedURL.User.Password()
		credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		client.Auth = "Basic " + credentials

		// Remove credentials from URL
		parsedURL.User = nil
		client.BaseURL = parsedURL.String()
	} else {
		client.BaseURL = ollamaURL
	}

	return client
}

// Ollama API request structure
type ollamaRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Format   map[string]interface{} `json:"format,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Ollama API response structure
type ollamaResponse struct {
	Message ollamaResponseMessage `json:"message"`
	Done    bool                  `json:"done"`
}

type ollamaResponseMessage struct {
	Content string `json:"content"`
}

// GenerateDescription sends a prompt to Ollama and returns the response
func (c *OllamaClient) GenerateDescription(prompt string) (string, error) {
	fmt.Printf("   🌐 Sending request to Ollama API (model: %s) with structured outputs...\n", c.Model)
	apiURL := strings.TrimSuffix(c.BaseURL, "/") + "/api/chat"

	// Define JSON schema for structured output
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "A concise PR title with appropriate emoji (max 80 characters)",
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "A detailed markdown PR description following professional format",
			},
		},
		"required": []string{"title", "body"},
	}

	reqBody := ollamaRequest{
		Model: c.Model,
		Messages: []ollamaMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
		Format: schema,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Auth != "" {
		req.Header.Set("Authorization", c.Auth)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error: %s - %s", resp.Status, string(body))
	}
	fmt.Println("   ✅ Ollama API responded successfully")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Println("   ✅ Received structured JSON response from Ollama")
	return ollamaResp.Message.Content, nil
}

// GetProviderInfo returns the provider name and model
func (c *OllamaClient) GetProviderInfo() (provider, model string) {
	return "Ollama", c.Model
}
