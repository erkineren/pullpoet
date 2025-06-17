package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pullpoet/config"
	"pullpoet/internal/ai"
	"pullpoet/internal/clickup"
	"pullpoet/internal/git"
	"pullpoet/internal/pr"

	"github.com/spf13/cobra"
)

var (
	repo        string
	source      string
	target      string
	description string
	provider    string
	apiKey      string
	ollamaURL   string
	model       string
	fastMode    bool
	outputFile  string
	// ClickUp integration variables
	clickupPAT    string
	clickupTaskID string
)

var rootCmd = &cobra.Command{
	Use:   "pullpoet",
	Short: "Generate AI-powered pull request descriptions",
	Long:  `PullPoet is a CLI tool that generates pull request titles and descriptions using AI providers like OpenAI and Ollama.`,
	RunE:  run,
}

func init() {
	rootCmd.Flags().StringVar(&repo, "repo", "", "Git repository URL (required)")
	rootCmd.Flags().StringVar(&source, "source", "", "Source branch name (required)")
	rootCmd.Flags().StringVar(&target, "target", "", "Target branch name (required)")
	rootCmd.Flags().StringVar(&description, "description", "", "Optional issue/task description from ClickUp, Jira, etc.")
	rootCmd.Flags().StringVar(&provider, "provider", "", "AI provider: 'openai', 'ollama', or 'gemini' (required)")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for OpenAI or Gemini (required when provider is 'openai' or 'gemini')")
	rootCmd.Flags().StringVar(&ollamaURL, "ollama-url", "", "Ollama endpoint URL with credentials (required when provider is 'ollama')")
	rootCmd.Flags().StringVar(&model, "model", "", "AI model to use (required)")
	rootCmd.Flags().BoolVar(&fastMode, "fast", false, "Use fast native git commands (recommended for large repositories)")
	rootCmd.Flags().StringVar(&outputFile, "output", "", "Save PR content to file (optional)")

	// ClickUp integration flags
	rootCmd.Flags().StringVar(&clickupPAT, "clickup-pat", "", "ClickUp Personal Access Token (optional)")
	rootCmd.Flags().StringVar(&clickupTaskID, "clickup-task-id", "", "ClickUp Task ID to fetch description from (optional)")

	rootCmd.MarkFlagRequired("repo")
	rootCmd.MarkFlagRequired("source")
	rootCmd.MarkFlagRequired("target")
	rootCmd.MarkFlagRequired("provider")
	rootCmd.MarkFlagRequired("model")
}

// savePRToFile saves the PR content to the specified file
func savePRToFile(result *pr.Result, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Format the content with title at the top and description below
	content := fmt.Sprintf("# %s\n\n%s\n", result.Title, result.Body)

	// Write to file (overwrite if exists)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Starting PullPoet...")

	// Validate configuration
	fmt.Println("📋 Validating configuration...")
	cfg := &config.Config{
		Repo:          repo,
		Source:        source,
		Target:        target,
		Description:   description,
		Provider:      provider,
		APIKey:        apiKey,
		OllamaURL:     ollamaURL,
		Model:         model,
		ClickUpPAT:    clickupPAT,
		ClickUpTaskID: clickupTaskID,
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	fmt.Printf("✅ Configuration validated - Provider: %s, Model: %s\n", cfg.Provider, cfg.Model)

	// Fetch ClickUp task description if ClickUp credentials are provided
	var finalDescription string
	if cfg.ClickUpPAT != "" && cfg.ClickUpTaskID != "" {
		fmt.Printf("📋 Fetching task description from ClickUp (Task ID: %s)...\n", cfg.ClickUpTaskID)
		clickupClient := clickup.NewClient(cfg.ClickUpPAT)
		task, err := clickupClient.GetTask(cfg.ClickUpTaskID)
		if err != nil {
			return fmt.Errorf("failed to fetch ClickUp task: %w", err)
		}
		finalDescription = task.FormatTaskDescription()
		fmt.Printf("✅ ClickUp task fetched successfully: %s\n", task.Name)
	} else {
		finalDescription = cfg.Description
		if finalDescription != "" {
			fmt.Println("📝 Using manually provided description")
		} else {
			fmt.Println("📝 No task description provided")
		}
	}

	// Clone repository and get diff with commit information
	fmt.Printf("📦 Cloning repository: %s\n", cfg.Repo)
	fmt.Printf("🔄 Analyzing changes between '%s' and '%s' branches...\n", cfg.Source, cfg.Target)

	var gitResult *git.GitResult
	var err error

	if fastMode {
		fmt.Println("⚡ Using fast mode (native git commands)...")
		fastClient := git.NewFastClient()
		gitResult, err = fastClient.GetDiffWithCommits(cfg.Repo, cfg.Source, cfg.Target)
	} else {
		fmt.Println("🐹 Using go-git library (optimized)...")
		gitClient := git.NewClient()
		gitResult, err = gitClient.GetDiffWithCommits(cfg.Repo, cfg.Source, cfg.Target)
	}

	if err != nil {
		return fmt.Errorf("failed to analyze git changes: %w", err)
	}
	fmt.Printf("✅ Git analysis completed successfully (%d characters diff, %d commits)\n", len(gitResult.Diff), len(gitResult.Commits))

	// Create AI client
	fmt.Printf("🤖 Initializing %s AI client with model '%s'...\n", cfg.Provider, cfg.Model)
	var aiClient ai.Client
	switch cfg.Provider {
	case "openai":
		aiClient = ai.NewOpenAIClient(cfg.APIKey, cfg.Model)
	case "ollama":
		aiClient = ai.NewOllamaClient(cfg.OllamaURL, cfg.Model)
	case "gemini":
		var geminiErr error
		aiClient, geminiErr = ai.NewGeminiClient(cfg.APIKey, cfg.Model)
		if geminiErr != nil {
			return fmt.Errorf("failed to create Gemini client: %w", geminiErr)
		}
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
	fmt.Println("✅ AI client initialized successfully")

	// Generate PR description
	fmt.Println("💭 Building prompt and sending to AI...")
	generator := pr.NewGenerator(aiClient)
	result, err := generator.Generate(gitResult, finalDescription, cfg.Repo)
	if err != nil {
		return fmt.Errorf("failed to generate PR description: %w", err)
	}
	fmt.Println("✅ AI response received and parsed successfully")

	// Output result
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("🎉 Generated PR Description")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("\n📋 **Title:**\n%s\n", result.Title)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("\n📝 **Description:**\n%s\n", result.Body)
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("✅ PR description generated successfully!")

	// Save to file if output path is provided
	if outputFile != "" {
		if err := savePRToFile(result, outputFile); err != nil {
			fmt.Printf("⚠️  Warning: Failed to save PR to file: %v\n", err)
		} else {
			fmt.Printf("💾 PR content saved to: %s\n", outputFile)
		}
	}

	fmt.Println("💡 You can now copy this content to your pull request.")

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
