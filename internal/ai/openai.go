package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

// OpenAI API request structure
type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI API response structure
type openAIResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
}

// GenerateDescription sends a prompt to OpenAI and returns the response
func (c *OpenAIClient) GenerateDescription(prompt string) (string, error) {
	fmt.Printf("   🌐 Sending request to OpenAI API (model: %s)...\n", c.model)
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := openAIRequest{
		Model: c.model,
		Messages: []message{
			{
				Role:    "system",
				Content: "You are a helpful assistant that generates pull request titles and descriptions. Respond with a JSON object containing 'title' and 'body' fields. The title should be a concise one-line summary, and the body should be a detailed markdown description explaining the changes.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}
	fmt.Println("   ✅ OpenAI API responded successfully")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	content := openAIResp.Choices[0].Message.Content

	// Try to parse as JSON first
	var response Response
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		// If not JSON, treat as plain text and extract title/body
		lines := strings.Split(strings.TrimSpace(content), "\n")
		if len(lines) > 0 {
			response.Title = strings.TrimSpace(lines[0])
			if len(lines) > 1 {
				response.Body = strings.TrimSpace(strings.Join(lines[1:], "\n"))
			}
		}
	}

	// Format the response
	result := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", response.Title, response.Body)
	return result, nil
}

// GetProviderInfo returns the provider name and model
func (c *OpenAIClient) GetProviderInfo() (provider, model string) {
	return "OpenAI", c.model
}
