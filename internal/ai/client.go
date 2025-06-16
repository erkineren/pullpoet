package ai

// Client defines the interface for AI providers
type Client interface {
	GenerateDescription(prompt string) (string, error)
}

// Response represents the AI response structure
type Response struct {
	Title string
	Body  string
}
