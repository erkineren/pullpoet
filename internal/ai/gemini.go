package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

// GeminiClient implements the Client interface for Google Gemini
type GeminiClient struct {
	apiKey string
	model  string
	client *genai.Client
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(apiKey, model string) (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		client: client,
	}, nil
}

// GenerateDescription sends a prompt to Gemini and returns the response
func (c *GeminiClient) GenerateDescription(prompt string) (string, error) {
	fmt.Printf("   🌐 Sending request to Gemini API (model: %s)...\n", c.model)

	ctx := context.Background()

	// Configure the model for structured JSON output
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"title": {
					Type:        genai.TypeString,
					Description: "A concise one-line summary of the pull request changes",
				},
				"body": {
					Type:        genai.TypeString,
					Description: "A detailed markdown description explaining the changes made in the pull request",
				},
			},
			PropertyOrdering: []string{"title", "body"},
			Required:         []string{"title", "body"},
		},
	}

	// Create a focused prompt for pull request analysis
	analysisPrompt := fmt.Sprintf(`Analyze the following git diff and generate a pull request title and description:

%s

Please provide:
- title: A clear, concise summary of what this pull request does
- body: A detailed explanation in markdown format covering what was changed, why it was changed, and any relevant implementation details`, prompt)

	// Generate content using the Models API
	result, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(analysisPrompt),
		config,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	fmt.Println("   ✅ Gemini API responded successfully")

	// Get the response text (should be structured JSON)
	content := result.Text()
	if content == "" {
		return "", fmt.Errorf("no response generated")
	}

	// Parse the JSON response (guaranteed by structured output)
	var response Response
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return "", fmt.Errorf("failed to parse structured JSON response: %w", err)
	}

	// Validate required fields
	if response.Title == "" {
		return "", fmt.Errorf("title field is empty in response")
	}
	if response.Body == "" {
		return "", fmt.Errorf("body field is empty in response")
	}

	// Format the response
	result_formatted := fmt.Sprintf("TITLE: %s\n\nBODY:\n%s", response.Title, response.Body)
	return result_formatted, nil
}
