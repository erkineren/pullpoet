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
	"pullpoet/internal/jira"
	"pullpoet/internal/pr"
	"pullpoet/internal/ui"

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
	language        string
	// ClickUp integration variables
	clickupPAT    string
	clickupTaskID string
	// Jira integration variables
	jiraBaseURL  string
	jiraUsername string
	jiraAPIToken string
	jiraTaskID   string
)

// Environment variable names
const (
	EnvProvider        = "PULLPOET_PROVIDER"
	EnvProviderBaseURL = "PULLPOET_PROVIDER_BASE_URL"
	EnvModel           = "PULLPOET_MODEL"
	EnvAPIKey          = "PULLPOET_API_KEY"
	EnvClickUpPAT      = "PULLPOET_CLICKUP_PAT"
	EnvLanguage        = "PULLPOET_LANGUAGE"
	// EnvClickUpTaskID   = "PULLPOET_CLICKUP_TASK_ID" // Removed - task ID should be provided per PR
	EnvJiraBaseURL  = "PULLPOET_JIRA_BASE_URL"
	EnvJiraUsername = "PULLPOET_JIRA_USERNAME"
	EnvJiraAPIToken = "PULLPOET_JIRA_API_TOKEN"
	// EnvJiraTaskID      = "PULLPOET_JIRA_TASK_ID" // Removed - task ID should be provided per PR
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

// getLanguageFromEnvOrFlag returns language from environment or flag
func getLanguageFromEnvOrFlag() string {
	if language != "" {
		return language
	}
	return getEnvOrDefault(EnvLanguage, "en")
}

// getClickUpTaskIDFromEnvOrFlag returns ClickUp Task ID from environment or flag
// func getClickUpTaskIDFromEnvOrFlag() string {
// 	if clickupTaskID != "" {
// 		return clickupTaskID
// 	}
// 	return getEnvOrDefault(EnvClickUpTaskID, "")
// }

// getJiraBaseURLFromEnvOrFlag returns Jira base URL from environment or flag
func getJiraBaseURLFromEnvOrFlag() string {
	if jiraBaseURL != "" {
		return jiraBaseURL
	}
	return getEnvOrDefault(EnvJiraBaseURL, "")
}

// getJiraUsernameFromEnvOrFlag returns Jira username from environment or flag
func getJiraUsernameFromEnvOrFlag() string {
	if jiraUsername != "" {
		return jiraUsername
	}
	return getEnvOrDefault(EnvJiraUsername, "")
}

// getJiraAPITokenFromEnvOrFlag returns Jira API token from environment or flag
func getJiraAPITokenFromEnvOrFlag() string {
	if jiraAPIToken != "" {
		return jiraAPIToken
	}
	return getEnvOrDefault(EnvJiraAPIToken, "")
}

var rootCmd = &cobra.Command{
	Use:     "pullpoet",
	Short:   "Generate AI-powered pull request descriptions",
	Long:    `PullPoet is a CLI tool that generates pull request titles and descriptions using AI providers like OpenAI and Ollama.`,
	Version: version,
	RunE:    run,
}

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview changes before committing",
	Long:  `Analyze staged changes and generate a preview of what the commit message and description would look like.`,
	RunE:  runPreview,
}

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Generate example .pullpoet.yml configuration file",
	Long:  `Creates an example .pullpoet.yml file in the current directory with all available configuration options.`,
	RunE:  runInitConfig,
}

func init() {
	// Root command flags
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
	rootCmd.Flags().StringVar(&language, "language", "", "Language for the generated PR description (default: en, can also be set via PULLPOET_LANGUAGE env var)")

	// ClickUp integration flags
	rootCmd.Flags().StringVar(&clickupPAT, "clickup-pat", "", "ClickUp Personal Access Token (can also be set via PULLPOET_CLICKUP_PAT env var)")
	rootCmd.Flags().StringVar(&clickupTaskID, "clickup-task-id", "", "ClickUp Task ID(s) to fetch description from, comma-separated for multiple tasks (e.g., 'task1,task2,task3')")

	// Jira integration flags
	rootCmd.Flags().StringVar(&jiraBaseURL, "jira-base-url", "", "Jira base URL (e.g., https://yourcompany.atlassian.net, can also be set via PULLPOET_JIRA_BASE_URL env var)")
	rootCmd.Flags().StringVar(&jiraUsername, "jira-username", "", "Jira username/email (can also be set via PULLPOET_JIRA_USERNAME env var)")
	rootCmd.Flags().StringVar(&jiraAPIToken, "jira-api-token", "", "Jira API token (can also be set via PULLPOET_JIRA_API_TOKEN env var)")
	rootCmd.Flags().StringVar(&jiraTaskID, "jira-task-id", "", "Jira issue key(s) to fetch description from, comma-separated for multiple issues (e.g., 'HIP-1234,HIP-1250')")

	// Preview command flags (inherit from root)
	previewCmd.Flags().StringVar(&repo, "repo", "", "Git repository URL (auto-detected if not provided and running in git repo)")
	previewCmd.Flags().StringVar(&source, "source", "", "Source branch name (auto-detected as current branch if not provided)")
	previewCmd.Flags().StringVar(&target, "target", "", "Target branch name (auto-detected as default branch if not provided)")
	previewCmd.Flags().StringVar(&description, "description", "", "Optional issue/task description from ClickUp, Jira, etc.")
	previewCmd.Flags().StringVar(&provider, "provider", "", "AI provider: 'openai', 'ollama', 'gemini', or 'openwebui' (can also be set via PULLPOET_PROVIDER env var)")
	previewCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for OpenAI or Gemini (can also be set via PULLPOET_API_KEY env var)")
	previewCmd.Flags().StringVar(&providerBaseURL, "provider-base-url", "", "Base URL for AI provider (can also be set via PULLPOET_PROVIDER_BASE_URL env var)")
	previewCmd.Flags().StringVar(&model, "model", "", "AI model to use (can also be set via PULLPOET_MODEL env var)")
	previewCmd.Flags().BoolVar(&fastMode, "fast", false, "Use fast native git commands (recommended for large repositories)")
	previewCmd.Flags().StringVar(&outputFile, "output", "", "Save preview content to file (optional)")
	previewCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "Custom system prompt file path to override default (optional)")
	previewCmd.Flags().StringVar(&language, "language", "", "Language for the generated preview (default: en, can also be set via PULLPOET_LANGUAGE env var)")

	// ClickUp integration flags for preview
	previewCmd.Flags().StringVar(&clickupPAT, "clickup-pat", "", "ClickUp Personal Access Token (can also be set via PULLPOET_CLICKUP_PAT env var)")
	previewCmd.Flags().StringVar(&clickupTaskID, "clickup-task-id", "", "ClickUp Task ID(s) to fetch description from, comma-separated for multiple tasks (e.g., 'task1,task2,task3')")

	// Jira integration flags for preview
	previewCmd.Flags().StringVar(&jiraBaseURL, "jira-base-url", "", "Jira base URL (e.g., https://yourcompany.atlassian.net, can also be set via PULLPOET_JIRA_BASE_URL env var)")
	previewCmd.Flags().StringVar(&jiraUsername, "jira-username", "", "Jira username/email (can also be set via PULLPOET_JIRA_USERNAME env var)")
	previewCmd.Flags().StringVar(&jiraAPIToken, "jira-api-token", "", "Jira API token (can also be set via PULLPOET_JIRA_API_TOKEN env var)")
	previewCmd.Flags().StringVar(&jiraTaskID, "jira-task-id", "", "Jira issue key(s) to fetch description from, comma-separated for multiple issues (e.g., 'HIP-1234,HIP-1250')")

	// Set version template and enable -v shorthand
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.Flags().BoolP("version", "v", false, "version for pullpoet")

	// Add subcommands to root
	rootCmd.AddCommand(previewCmd)
	rootCmd.AddCommand(initConfigCmd)

	// Flag validasyonunu kaldırdık, run fonksiyonunda manuel validasyon yapacağız
}

// fetchClickUpTasks fetches multiple ClickUp tasks and combines their descriptions
func fetchClickUpTasks(pat, taskIDs string) (string, error) {
	// Parse task IDs (comma-separated)
	ids := strings.Split(taskIDs, ",")
	var cleanIDs []string
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			cleanIDs = append(cleanIDs, trimmed)
		}
	}

	if len(cleanIDs) == 0 {
		return "", fmt.Errorf("no valid task IDs provided")
	}

	fmt.Printf("📋 Fetching %d task(s) from ClickUp...\n", len(cleanIDs))
	clickupClient := clickup.NewClient(pat)

	var descriptions []string
	for i, taskID := range cleanIDs {
		fmt.Printf("   [%d/%d] Fetching task: %s\n", i+1, len(cleanIDs), taskID)
		task, err := clickupClient.GetTask(taskID)
		if err != nil {
			return "", fmt.Errorf("failed to fetch ClickUp task %s: %w", taskID, err)
		}

		descriptions = append(descriptions, task.FormatTaskDescription())
		fmt.Printf("   ✅ Task fetched: %s\n", task.Name)

		if len(task.Comments) > 0 {
			totalReplies := 0
			for _, comment := range task.Comments {
				totalReplies += len(comment.Replies)
			}
			fmt.Printf("   💬 %d comments", len(task.Comments))
			if totalReplies > 0 {
				fmt.Printf(" (%d replies)", totalReplies)
			}
			fmt.Println()
		}
	}

	// Combine all task descriptions
	if len(descriptions) == 1 {
		return descriptions[0], nil
	}

	var combined strings.Builder
	combined.WriteString(fmt.Sprintf("**Multiple ClickUp Tasks (%d tasks)**\n\n", len(descriptions)))
	combined.WriteString(strings.Repeat("=", 80) + "\n\n")

	for i, desc := range descriptions {
		combined.WriteString(fmt.Sprintf("### Task %d of %d\n\n", i+1, len(descriptions)))
		combined.WriteString(desc)
		if i < len(descriptions)-1 {
			combined.WriteString("\n\n" + strings.Repeat("-", 80) + "\n\n")
		}
	}

	return combined.String(), nil
}

// fetchJiraIssues fetches multiple Jira issues and combines their descriptions
func fetchJiraIssues(baseURL, username, apiToken, issueKeys string) (string, error) {
	// Parse issue keys (comma-separated)
	keys := strings.Split(issueKeys, ",")
	var cleanKeys []string
	for _, key := range keys {
		trimmed := strings.TrimSpace(key)
		if trimmed != "" {
			cleanKeys = append(cleanKeys, trimmed)
		}
	}

	if len(cleanKeys) == 0 {
		return "", fmt.Errorf("no valid issue keys provided")
	}

	fmt.Printf("📋 Fetching %d issue(s) from Jira...\n", len(cleanKeys))
	jiraClient := jira.NewClient(baseURL, username, apiToken)

	var descriptions []string
	for i, issueKey := range cleanKeys {
		fmt.Printf("   [%d/%d] Fetching issue: %s\n", i+1, len(cleanKeys), issueKey)
		issue, err := jiraClient.GetIssue(issueKey)
		if err != nil {
			return "", fmt.Errorf("failed to fetch Jira issue %s: %w", issueKey, err)
		}

		descriptions = append(descriptions, issue.FormatIssueDescription())
		fmt.Printf("   ✅ Issue fetched: %s\n", issue.Summary)

		if len(issue.Comments) > 0 {
			fmt.Printf("   💬 %d comments\n", len(issue.Comments))
		}
	}

	// Combine all issue descriptions
	if len(descriptions) == 1 {
		return descriptions[0], nil
	}

	var combined strings.Builder
	combined.WriteString(fmt.Sprintf("**Multiple Jira Issues (%d issues)**\n\n", len(descriptions)))
	combined.WriteString(strings.Repeat("=", 80) + "\n\n")

	for i, desc := range descriptions {
		combined.WriteString(fmt.Sprintf("### Issue %d of %d\n\n", i+1, len(descriptions)))
		combined.WriteString(desc)
		if i < len(descriptions)-1 {
			combined.WriteString("\n\n" + strings.Repeat("-", 80) + "\n\n")
		}
	}

	return combined.String(), nil
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

	// Load configuration file
	fileConfig, err := config.LoadConfigFile()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to load config file: %v\n", err)
		fileConfig = &config.FileConfig{UI: config.DefaultUIConfig()}
	}

	// Initialize UI
	uiConfig := ui.Config{
		Colors:       fileConfig.UI.Colors,
		ProgressBars: fileConfig.UI.ProgressBars,
		Emoji:        fileConfig.UI.Emoji,
		Verbose:      fileConfig.UI.Verbose,
		Theme:        fileConfig.UI.Theme,
	}
	termUI := ui.New(uiConfig)

	termUI.Section("Starting PullPoet")

	// Merge file config values with CLI flags (file config has lower priority)
	// Git configuration from file
	if repo == "" && fileConfig.Repo != "" {
		repo = fileConfig.Repo
		termUI.Verbose(fmt.Sprintf("Using repo from config file: %s", repo))
	}
	if source == "" && fileConfig.Source != "" {
		source = fileConfig.Source
		termUI.Verbose(fmt.Sprintf("Using source branch from config file: %s", source))
	}
	if target == "" && fileConfig.Target != "" {
		target = fileConfig.Target
		termUI.Verbose(fmt.Sprintf("Using target branch from config file: %s", target))
	}

	// AI Provider configuration from file
	if provider == "" && fileConfig.Provider != "" {
		provider = fileConfig.Provider
		termUI.Verbose(fmt.Sprintf("Using provider from config file: %s", provider))
	}
	if model == "" && fileConfig.Model != "" {
		model = fileConfig.Model
		termUI.Verbose(fmt.Sprintf("Using model from config file: %s", model))
	}
	if apiKey == "" && fileConfig.APIKey != "" {
		apiKey = fileConfig.APIKey
		termUI.Verbose("Using API key from config file")
	}
	if systemPrompt == "" && fileConfig.SystemPrompt != "" {
		systemPrompt = fileConfig.SystemPrompt
		termUI.Verbose(fmt.Sprintf("Using system prompt from config file: %s", systemPrompt))
	}
	if language == "" && fileConfig.Language != "" {
		language = fileConfig.Language
		termUI.Verbose(fmt.Sprintf("Using language from config file: %s", language))
	}

	// Fast mode from config file (only if not set via CLI flag)
	// Note: For bool flags, cobra sets them to false by default, so we need to check if flag was actually provided
	if !cmd.Flags().Changed("fast") && fileConfig.FastMode {
		fastMode = fileConfig.FastMode
		termUI.Verbose(fmt.Sprintf("Using fast mode from config file: %v", fastMode))
	}

	// Output file from config
	if outputFile == "" && fileConfig.Output != "" {
		outputFile = fileConfig.Output
		termUI.Verbose(fmt.Sprintf("Using output file from config file: %s", outputFile))
	}

	// Manual validation for required fields (including environment variables)
	finalProvider := getProviderFromEnvOrFlag()
	finalModel := getModelFromEnvOrFlag()

	if finalProvider == "" {
		return fmt.Errorf("provider is required (can be set via --provider flag, .pullpoet.yml, or PULLPOET_PROVIDER environment variable)")
	}

	if finalModel == "" {
		return fmt.Errorf("model is required (can be set via --model flag, .pullpoet.yml, or PULLPOET_MODEL environment variable)")
	}

	// Auto-detect git information if not provided (after config file merge)
	if repo == "" || source == "" || target == "" {
		termUI.Info("Auto-detecting git repository information...")
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
		JiraBaseURL:     getJiraBaseURLFromEnvOrFlag(),
		JiraUsername:    getJiraUsernameFromEnvOrFlag(),
		JiraAPIToken:    getJiraAPITokenFromEnvOrFlag(),
		JiraTaskID:      jiraTaskID,
		Language:        getLanguageFromEnvOrFlag(),
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	fmt.Printf("✅ Configuration validated - Provider: %s, Model: %s\n", cfg.Provider, cfg.Model)

	// Fetch task description from ClickUp or Jira
	var finalDescription string
	if cfg.ClickUpPAT != "" && cfg.ClickUpTaskID != "" {
		var err error
		finalDescription, err = fetchClickUpTasks(cfg.ClickUpPAT, cfg.ClickUpTaskID)
		if err != nil {
			return err
		}
		fmt.Println("✅ All ClickUp tasks fetched successfully")
	} else if cfg.JiraBaseURL != "" && cfg.JiraUsername != "" && cfg.JiraAPIToken != "" && cfg.JiraTaskID != "" {
		var err error
		finalDescription, err = fetchJiraIssues(cfg.JiraBaseURL, cfg.JiraUsername, cfg.JiraAPIToken, cfg.JiraTaskID)
		if err != nil {
			return err
		}
		fmt.Println("✅ All Jira issues fetched successfully")
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
	result, err := generator.Generate(gitResult, finalDescription, cfg.Repo, cfg.Language, true)
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

func runPreview(cmd *cobra.Command, args []string) error {
	// Load configuration file
	fileConfig, err := config.LoadConfigFile()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to load config file: %v\n", err)
		fileConfig = &config.FileConfig{UI: config.DefaultUIConfig()}
	}

	// Initialize UI
	uiConfig := ui.Config{
		Colors:       fileConfig.UI.Colors,
		ProgressBars: fileConfig.UI.ProgressBars,
		Emoji:        fileConfig.UI.Emoji,
		Verbose:      fileConfig.UI.Verbose,
		Theme:        fileConfig.UI.Theme,
	}
	termUI := ui.New(uiConfig)

	termUI.Section("Starting PullPoet Preview Mode")

	// Merge file config values with CLI flags (file config has lower priority)
	// Git configuration from file
	if repo == "" && fileConfig.Repo != "" {
		repo = fileConfig.Repo
		termUI.Verbose(fmt.Sprintf("Using repo from config file: %s", repo))
	}
	if source == "" && fileConfig.Source != "" {
		source = fileConfig.Source
		termUI.Verbose(fmt.Sprintf("Using source branch from config file: %s", source))
	}
	if target == "" && fileConfig.Target != "" {
		target = fileConfig.Target
		termUI.Verbose(fmt.Sprintf("Using target branch from config file: %s", target))
	}

	// AI Provider configuration from file
	if provider == "" && fileConfig.Provider != "" {
		provider = fileConfig.Provider
		termUI.Verbose(fmt.Sprintf("Using provider from config file: %s", provider))
	}
	if model == "" && fileConfig.Model != "" {
		model = fileConfig.Model
		termUI.Verbose(fmt.Sprintf("Using model from config file: %s", model))
	}
	if apiKey == "" && fileConfig.APIKey != "" {
		apiKey = fileConfig.APIKey
		termUI.Verbose("Using API key from config file")
	}
	if systemPrompt == "" && fileConfig.SystemPrompt != "" {
		systemPrompt = fileConfig.SystemPrompt
		termUI.Verbose(fmt.Sprintf("Using system prompt from config file: %s", systemPrompt))
	}
	if language == "" && fileConfig.Language != "" {
		language = fileConfig.Language
		termUI.Verbose(fmt.Sprintf("Using language from config file: %s", language))
	}

	// Fast mode from config file (only if not set via CLI flag)
	if !cmd.Flags().Changed("fast") && fileConfig.FastMode {
		fastMode = fileConfig.FastMode
		termUI.Verbose(fmt.Sprintf("Using fast mode from config file: %v", fastMode))
	}

	// Output file from config
	if outputFile == "" && fileConfig.Output != "" {
		outputFile = fileConfig.Output
		termUI.Verbose(fmt.Sprintf("Using output file from config file: %s", outputFile))
	}

	// Manual validation for required fields (including environment variables)
	finalProvider := getProviderFromEnvOrFlag()
	finalModel := getModelFromEnvOrFlag()

	if finalProvider == "" {
		return fmt.Errorf("provider is required (can be set via --provider flag, .pullpoet.yml, or PULLPOET_PROVIDER environment variable)")
	}

	if finalModel == "" {
		return fmt.Errorf("model is required (can be set via --model flag, .pullpoet.yml, or PULLPOET_MODEL environment variable)")
	}

	// Auto-detect git information if not provided (after config file merge)
	if repo == "" || source == "" || target == "" {
		termUI.Info("Auto-detecting git repository information...")
		gitClient := git.NewClient()
		gitInfo, err := gitClient.GetGitInfoFromCurrentDir()
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w", err)
		}
		if !gitInfo.IsGitRepo {
			return fmt.Errorf("not in a git repository - please provide --repo, --source and --target flags or set them in .pullpoet.yml")
		}
		if repo == "" {
			repo = gitInfo.RepoURL
			termUI.Step(fmt.Sprintf("Repository: %s", repo))
		}
		if source == "" {
			source = gitInfo.CurrentBranch
			termUI.Step(fmt.Sprintf("Source branch: %s", source))
		}
		if target == "" {
			target = gitInfo.DefaultBranch
			termUI.Step(fmt.Sprintf("Target branch: %s", target))
		}
		termUI.Success("Git repository information detected")
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
		JiraBaseURL:     getJiraBaseURLFromEnvOrFlag(),
		JiraUsername:    getJiraUsernameFromEnvOrFlag(),
		JiraAPIToken:    getJiraAPITokenFromEnvOrFlag(),
		JiraTaskID:      jiraTaskID,
		Language:        getLanguageFromEnvOrFlag(),
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	fmt.Printf("✅ Configuration validated - Provider: %s, Model: %s\n", cfg.Provider, cfg.Model)

	// Fetch task description from ClickUp or Jira
	var finalDescription string
	if cfg.ClickUpPAT != "" && cfg.ClickUpTaskID != "" {
		var err error
		finalDescription, err = fetchClickUpTasks(cfg.ClickUpPAT, cfg.ClickUpTaskID)
		if err != nil {
			return err
		}
		fmt.Println("✅ All ClickUp tasks fetched successfully")
	} else if cfg.JiraBaseURL != "" && cfg.JiraUsername != "" && cfg.JiraAPIToken != "" && cfg.JiraTaskID != "" {
		var err error
		finalDescription, err = fetchJiraIssues(cfg.JiraBaseURL, cfg.JiraUsername, cfg.JiraAPIToken, cfg.JiraTaskID)
		if err != nil {
			return err
		}
		fmt.Println("✅ All Jira issues fetched successfully")
	} else {
		finalDescription = cfg.Description
		if finalDescription != "" {
			fmt.Println("📝 Using manually provided description")
		} else {
			fmt.Println("📝 No task description provided")
		}
	}

	// Get staged changes
	fmt.Println("📊 Analyzing staged changes...")
	gitClient := git.NewClient()
	stagedDiff, err := gitClient.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	if stagedDiff == "" {
		fmt.Println("⚠️  No staged changes found. Please run 'git add' to stage your changes first.")
		return nil
	}

	fmt.Printf("✅ Found staged changes (%d characters)\n", len(stagedDiff))

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

	// Generate preview
	fmt.Println("💭 Analyzing changes and generating preview...")
	if cfg.SystemPrompt != "" {
		fmt.Printf("📝 Using custom system prompt from: %s\n", cfg.SystemPrompt)
	} else {
		fmt.Println("📝 Using default embedded system prompt")
	}
	generator := pr.NewGenerator(aiClient, cfg.SystemPrompt)

	// Create a GitResult with staged diff
	gitResult := &git.GitResult{
		Diff:          stagedDiff,
		Commits:       []git.CommitInfo{}, // No commits for staged changes
		DefaultBranch: target,
	}

	result, err := generator.Generate(gitResult, finalDescription, cfg.Repo, cfg.Language, false)
	if err != nil {
		return fmt.Errorf("failed to generate preview: %w", err)
	}
	fmt.Println("✅ Analysis completed successfully")

	// Output result
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("🔍 Preview of Changes (Staged)")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("\n📋 **Analysis Summary:**\n%s\n", result.Title)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("\n📝 **Detailed Analysis:**\n%s\n", result.Body)
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("✅ Preview generated successfully!")

	// Save to file if output path is provided
	if outputFile != "" {
		if err := savePRToFile(result, outputFile); err != nil {
			fmt.Printf("⚠️  Warning: Failed to save preview to file: %v\n", err)
		} else {
			fmt.Printf("💾 Preview content saved to: %s\n", outputFile)
		}
	}

	fmt.Println("💡 You can review these changes before committing.")

	return nil
}

func runInitConfig(cmd *cobra.Command, args []string) error {
	configPath := ".pullpoet.yml"

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Config file %s already exists. Overwrite? (y/N): ", configPath)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("❌ Aborted")
			return nil
		}
	}

	// Generate example config
	exampleConfig := config.GenerateExampleConfig()

	// Write to file
	if err := os.WriteFile(configPath, []byte(exampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("✅ Created example configuration file: .pullpoet.yml")
	fmt.Println("📝 Edit this file to set your default values")
	fmt.Println("💡 Tip: Use environment variables (e.g., ${PULLPOET_API_KEY}) for sensitive data")

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
