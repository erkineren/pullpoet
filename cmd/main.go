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

// version will be set by ldflags during build
var version = "dev"

var (
	repo            string
	source          string
	target          string
	description     string
	provider        string
	apiKey          string
	providerBaseURL string
	model           string
	fastMode        bool
	outputFile      string
	systemPrompt    string
	// ClickUp integration variables
	clickupPAT    string
	clickupTaskID string
)

// Environment variable names
const (
	EnvProvider        = "PULLPOET_PROVIDER"
	EnvProviderBaseURL = "PULLPOET_PROVIDER_BASE_URL"
	EnvModel           = "PULLPOET_MODEL"
	EnvAPIKey          = "PULLPOET_API_KEY"
	EnvClickUpPAT      = "PULLPOET_CLICKUP_PAT"
	// EnvClickUpTaskID   = "PULLPOET_CLICKUP_TASK_ID" // Removed - task ID should be provided per PR
)

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(envName, defaultValue string) string {
	if value := os.Getenv(envName); value != "" {
		return value
	}
	return defaultValue
}

// getProviderFromEnvOrFlag returns provider from environment or flag
func getProviderFromEnvOrFlag() string {
	if provider != "" {
		return provider
	}
	return getEnvOrDefault(EnvProvider, "")
}

// getProviderBaseURLFromEnvOrFlag returns provider base URL from environment or flag
func getProviderBaseURLFromEnvOrFlag() string {
	if providerBaseURL != "" {
		return providerBaseURL
	}
	return getEnvOrDefault(EnvProviderBaseURL, "")
}

// getModelFromEnvOrFlag returns model from environment or flag
func getModelFromEnvOrFlag() string {
	if model != "" {
		return model
	}
	return getEnvOrDefault(EnvModel, "")
}

// getAPIKeyFromEnvOrFlag returns API key from environment or flag
func getAPIKeyFromEnvOrFlag() string {
	if apiKey != "" {
		return apiKey
	}
	return getEnvOrDefault(EnvAPIKey, "")
}

// getClickUpPATFromEnvOrFlag returns ClickUp PAT from environment or flag
func getClickUpPATFromEnvOrFlag() string {
	if clickupPAT != "" {
		return clickupPAT
	}
	return getEnvOrDefault(EnvClickUpPAT, "")
}

// getClickUpTaskIDFromEnvOrFlag returns ClickUp Task ID from environment or flag
// func getClickUpTaskIDFromEnvOrFlag() string {
// 	if clickupTaskID != "" {
// 		return clickupTaskID
// 	}
// 	return getEnvOrDefault(EnvClickUpTaskID, "")
// }

var rootCmd = &cobra.Command{
	Use:     "pullpoet",
	Short:   "Generate AI-powered pull request descriptions",
	Long:    `PullPoet is a CLI tool that generates pull request titles and descriptions using AI providers like OpenAI and Ollama.`,
	Version: version,
	RunE:    run,
}

func init() {
	rootCmd.Flags().StringVar(&repo, "repo", "", "Git repository URL (auto-detected if not provided and running in git repo)")
	rootCmd.Flags().StringVar(&source, "source", "", "Source branch name (auto-detected as current branch if not provided)")
	rootCmd.Flags().StringVar(&target, "target", "", "Target branch name (auto-detected as default branch if not provided)")
	rootCmd.Flags().StringVar(&description, "description", "", "Optional issue/task description from ClickUp, Jira, etc.")
	rootCmd.Flags().StringVar(&provider, "provider", "", "AI provider: 'openai', 'ollama', 'gemini', or 'openwebui' (can also be set via PULLPOET_PROVIDER env var)")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for OpenAI or Gemini (can also be set via PULLPOET_API_KEY env var)")
	rootCmd.Flags().StringVar(&providerBaseURL, "provider-base-url", "", "Base URL for AI provider (can also be set via PULLPOET_PROVIDER_BASE_URL env var)")
	rootCmd.Flags().StringVar(&model, "model", "", "AI model to use (can also be set via PULLPOET_MODEL env var)")
	rootCmd.Flags().BoolVar(&fastMode, "fast", false, "Use fast native git commands (recommended for large repositories)")
	rootCmd.Flags().StringVar(&outputFile, "output", "", "Save PR content to file (optional)")
	rootCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "Custom system prompt file path to override default (optional)")

	// ClickUp integration flags
	rootCmd.Flags().StringVar(&clickupPAT, "clickup-pat", "", "ClickUp Personal Access Token (can also be set via PULLPOET_CLICKUP_PAT env var)")
	rootCmd.Flags().StringVar(&clickupTaskID, "clickup-task-id", "", "ClickUp Task ID to fetch description from (must be provided via flag)")

	// Set version template and enable -v shorthand
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.Flags().BoolP("version", "v", false, "version for pullpoet")

	// Flag validasyonunu kaldırdık, run fonksiyonunda manuel validasyon yapacağız
}

// autoDetectGitInfo attempts to auto-detect git repository information
func autoDetectGitInfo() (string, string, error) {
	gitClient := git.NewClient()
	gitInfo, err := gitClient.GetGitInfoFromCurrentDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to detect git information: %w", err)
	}

	if !gitInfo.IsGitRepo {
		return "", "", fmt.Errorf("not in a git repository - please provide --repo and --source flags")
	}

	fmt.Printf("🔍 Auto-detected git information:\n")
	fmt.Printf("   📦 Repository: %s\n", gitInfo.RepoURL)
	fmt.Printf("   🌿 Current branch: %s\n", gitInfo.CurrentBranch)
	fmt.Printf("   🎯 Default branch: %s\n", gitInfo.DefaultBranch)

	return gitInfo.RepoURL, gitInfo.CurrentBranch, nil
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
	// Check if version flag was used
	if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
		fmt.Println(version)
		return nil
	}

	fmt.Println("🚀 Starting PullPoet...")

	// Manual validation for required fields (including environment variables)
	finalProvider := getProviderFromEnvOrFlag()
	finalModel := getModelFromEnvOrFlag()

	if finalProvider == "" {
		return fmt.Errorf("provider is required (can be set via --provider flag or PULLPOET_PROVIDER environment variable)")
	}

	if finalModel == "" {
		return fmt.Errorf("model is required (can be set via --model flag or PULLPOET_MODEL environment variable)")
	}

	// Auto-detect git information if not provided
	if repo == "" || source == "" || target == "" {
		fmt.Println("🔍 Auto-detecting git repository information...")
		gitClient := git.NewClient()
		gitInfo, err := gitClient.GetGitInfoFromCurrentDir()
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w", err)
		}
		if !gitInfo.IsGitRepo {
			return fmt.Errorf("not in a git repository - please provide --repo, --source and --target flags")
		}
		if repo == "" {
			repo = gitInfo.RepoURL
			fmt.Printf("✅ Auto-detected repository: %s\n", repo)
		}
		if source == "" {
			source = gitInfo.CurrentBranch
			fmt.Printf("✅ Auto-detected source branch: %s\n", source)
		}
		if target == "" {
			target = gitInfo.DefaultBranch
			fmt.Printf("✅ Auto-detected target branch (default branch): %s\n", target)
		}
	}

	// Validate configuration
	fmt.Println("📋 Validating configuration...")
	cfg := &config.Config{
		Repo:            repo,
		Source:          source,
		Target:          target,
		Description:     description,
		Provider:        finalProvider,
		APIKey:          getAPIKeyFromEnvOrFlag(),
		ProviderBaseURL: getProviderBaseURLFromEnvOrFlag(),
		Model:           finalModel,
		SystemPrompt:    systemPrompt,
		ClickUpPAT:      getClickUpPATFromEnvOrFlag(),
		ClickUpTaskID:   clickupTaskID,
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
		aiClient = ai.NewOllamaClient(cfg.GetProviderBaseURL(), cfg.Model)
	case "gemini":
		var geminiErr error
		aiClient, geminiErr = ai.NewGeminiClient(cfg.APIKey, cfg.Model)
		if geminiErr != nil {
			return fmt.Errorf("failed to create Gemini client: %w", geminiErr)
		}
	case "openwebui":
		aiClient = ai.NewOpenWebUIClient(cfg.GetProviderBaseURL(), cfg.APIKey, cfg.Model)
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
	fmt.Println("✅ AI client initialized successfully")

	// Generate PR description
	fmt.Println("💭 Building prompt and sending to AI...")
	if cfg.SystemPrompt != "" {
		fmt.Printf("📝 Using custom system prompt from: %s\n", cfg.SystemPrompt)
	} else {
		fmt.Println("📝 Using default embedded system prompt")
	}
	generator := pr.NewGenerator(aiClient, cfg.SystemPrompt)
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
