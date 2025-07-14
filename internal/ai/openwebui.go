package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenWebUIClient implements the Client interface for OpenWebUI
type OpenWebUIClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// NewOpenWebUIClient creates a new OpenWebUI client
func NewOpenWebUIClient(baseURL, apiKey, model string) *OpenWebUIClient {
	// Ensure baseURL doesn't end with slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OpenWebUIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{},
	}
}

// OpenWebUI API request structure (similar to OpenAI but with OpenWebUI endpoint)
type openWebUIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

// OpenWebUI API response structure (same as OpenAI)
type openWebUIResponse struct {
	Choices []choice `json:"choices"`
}

// GenerateDescription sends a prompt to OpenWebUI and returns the response
func (c *OpenWebUIClient) GenerateDescription(prompt string) (string, error) {
	fmt.Printf("   🌐 Sending request to OpenWebUI API (model: %s)...\n", c.model)
	url := c.baseURL + "/api/chat/completions"

	reqBody := openWebUIRequest{
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
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenWebUI API error: %s - %s", resp.Status, string(body))
	}
	fmt.Println("   ✅ OpenWebUI API responded successfully")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openWebUIResp openWebUIResponse
	if err := json.Unmarshal(body, &openWebUIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openWebUIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenWebUI response")
	}

	content := openWebUIResp.Choices[0].Message.Content

	// Try to parse as JSON first
	var response Response

	// First attempt: try to parse the entire content as JSON
	if err := json.Unmarshal([]byte(content), &response); err == nil {
		// Format the response
		result := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", response.Title, response.Body)
		return result, nil
	}

	// Second attempt: Look for JSON object within the content
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := content[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &response); err == nil {
			// Format the response
			result := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", response.Title, response.Body)
			return result, nil
		}
	}

	// Third attempt: Look for JSON block with ```json markers
	if jsonStart := strings.Index(content, "```json"); jsonStart >= 0 {
		jsonStart += len("```json")
		if jsonEnd := strings.Index(content[jsonStart:], "```"); jsonEnd >= 0 {
			jsonStr := strings.TrimSpace(content[jsonStart : jsonStart+jsonEnd])
			if err := json.Unmarshal([]byte(jsonStr), &response); err == nil {
				// Format the response
				result := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", response.Title, response.Body)
				return result, nil
			}
		}
	}

	// Fallback: treat as plain text and extract title/body
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) > 0 {
		response.Title = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			response.Body = strings.TrimSpace(strings.Join(lines[1:], "\n"))
		}
	}

	// Format the response
	result := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", response.Title, response.Body)
	return result, nil
}

// GetProviderInfo returns the provider name and model
func (c *OpenWebUIClient) GetProviderInfo() (provider, model string) {
	return "OpenWebUI", c.model
}
