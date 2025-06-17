package ai

// Client defines the interface for AI providers
type Client interface {
	GenerateDescription(prompt string) (string, error)
	GetProviderInfo() (provider, model string)
}

// Response represents the AI response structure
type Response struct {
	Title string
	Body  string
}
