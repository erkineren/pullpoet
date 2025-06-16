package pr

import (
	"encoding/json"
	"fmt"
	"pullpoet/internal/ai"
	"pullpoet/internal/git"
	"strings"
)

// Generator handles PR description generation
type Generator struct {
	aiClient ai.Client
}

// Result represents the generated PR description
type Result struct {
	Title string
	Body  string
}

// PromptTemplate contains reusable prompt components
type PromptTemplate struct {
	SystemRole       string
	BaseInstructions string
	FormatTemplate   string
	Guidelines       []string
	Restrictions     []string
}

// EmojiCategories defines emoji sets for different sections
type EmojiCategories struct {
	Technical   []string
	WebDev      []string
	Database    []string
	Testing     []string
	Performance []string
	BugFix      []string
}

// NewGenerator creates a new PR generator
func NewGenerator(aiClient ai.Client) *Generator {
	return &Generator{
		aiClient: aiClient,
	}
}

// Generate creates a PR description based on the git diff and optional description
func (g *Generator) Generate(gitResult *git.GitResult, issueContext, repoURL string) (*Result, error) {
	fmt.Println("   📝 Building AI prompt...")

	// Check if we're using Ollama client (for structured outputs)
	if isOllamaClient(g.aiClient) {
		prompt := g.buildOllamaPrompt(gitResult, issueContext, repoURL)
		fmt.Printf("   ✅ Ollama structured prompt built (%d characters)\n", len(prompt))

		response, err := g.aiClient.GenerateDescription(prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to get AI response: %w", err)
		}

		fmt.Println("   🔍 Parsing structured Ollama response...")
		result, err := g.parseStructuredResponse(response)
		if err != nil {
			return nil, err
		}
		fmt.Println("   ✅ Structured response parsed successfully")
		return result, nil
	}

	// Default behavior for other clients (OpenAI)
	prompt := g.buildPrompt(gitResult, issueContext, repoURL)
	fmt.Printf("   ✅ Prompt built (%d characters)\n", len(prompt))

	response, err := g.aiClient.GenerateDescription(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	fmt.Println("   🔍 Parsing AI response...")
	result, err := g.parseResponse(response)
	if err != nil {
		return nil, err
	}
	fmt.Println("   ✅ Response parsed successfully")

	return result, nil
}

// isOllamaClient checks if the AI client is an Ollama client
func isOllamaClient(client ai.Client) bool {
	// Use type assertion to check if it's an Ollama client
	// We can check by looking at the type name or implement a marker interface
	switch client.(type) {
	case *ai.OllamaClient:
		return true
	default:
		return false
	}
}

// getPromptTemplate returns the appropriate template based on client type
func (g *Generator) getPromptTemplate(isOllama bool) *PromptTemplate {
	template := &PromptTemplate{
		SystemRole:       "You are an expert software engineer who creates professional, visually appealing pull request descriptions.",
		BaseInstructions: "Analyze the git changes and create a well-structured PR title and description.",
		Guidelines: []string{
			"Use emojis for better readability",
			"Focus on what changed and why it was necessary",
			"Keep the tone professional but engaging",
			"Include specific file paths where possible",
			"Use **bold** for important components/files",
			"Use bullet points and checkboxes for better readability",
		},
		Restrictions: []string{
			"NEVER mention fake testing coverage or suggest workflow processes",
			"DO NOT include: Merge checklists, next steps, documentation updates, CI/CD processes",
			"FOCUS ONLY on actual code changes",
			"Do NOT use placeholder URLs like 'your-repo' - use actual repository information",
		},
	}

	if isOllama {
		template.FormatTemplate = g.getOllamaFormatTemplate()
	} else {
		template.FormatTemplate = g.getOpenAIFormatTemplate()
	}

	return template
}

// getOllamaFormatTemplate returns the format template for Ollama
func (g *Generator) getOllamaFormatTemplate() string {
	return `Create a professional PR with:
- Title: Concise with emoji (max 80 chars)
- Body: Structured markdown with problem statement, solution, technical changes, and manual testing notes`
}

// getOpenAIFormatTemplate returns the detailed format template for OpenAI
func (g *Generator) getOpenAIFormatTemplate() string {
	return `**Create a JSON response with the following structure:**

` + "```json" + `
{
  "title": "🚀 Concise PR title with emoji (max 80 chars)",
  "body": "Professional markdown description (see format below)"
}
` + "```" + `

**PR Description Format Requirements:**
1. **Start with an appropriate emoji** for the title
2. **Use this markdown structure:**

` + "```markdown" + `
# 🚀 [Title with Emoji]

## 📋 Problem Statement / Overview
[Brief description of what was addressed]

## 🎯 Solution Overview
[High-level description of the approach]

## 🔧 Technical Changes

### 🔨 **[Category 1 with Emoji]**
- **[Component/File]**: [Description of change]
- **[Component/File]**: [Description of change]

### 🛠️ **[Category 2 with Emoji]**
- **[Component/File]**: [Description of change]

### 📱 **[Category 3 with Emoji]** (if applicable)
- **[Component/File]**: [Description of change]

## ✅ Key Features / Acceptance Criteria

- [x] **[Feature/Requirement]**: [Description]
- [x] **[Feature/Requirement]**: [Description]

## 🧪 Testing Notes

- **🔍 Code Review**: [Areas that need attention during review]
- **🌐 Deployment Notes**: [Important considerations for deployment]
- **👤 Usage Notes**: [How to use/test the new functionality]

## 📋 Files Changed
- ` + "`path/to/file1.ext`" + `
- ` + "`path/to/file2.ext`" + `
` + "```"
}

// getEmojiGuidelines returns emoji usage guidelines
func (g *Generator) getEmojiGuidelines() string {
	return `**HEAVILY use emojis throughout the description**, especially in subsections:
  - 🔧 🔨 🛠️ for technical changes and tools
  - 📱 💻 🌐 for frontend, backend, web-related
  - 🗃️ 📊 💾 for database and data-related
  - 🔍 🧪 👤 for testing categories
  - ⚡ 🚀 ✨ for performance and new features
  - 🐛 🔒 📋 for bugs, security, documentation

Title emoji examples: 🐛 for bug fix, 🚀 for feature, ♻️ for refactoring, 🔒 for security, ⚡ for performance`
}

// buildContextSection creates the context section with issue and commit info
func (g *Generator) buildContextSection(gitResult *git.GitResult, issueContext string, isOllama bool) string {
	var contextBuilder strings.Builder

	// Add issue context if provided
	if issueContext != "" {
		if isOllama {
			contextBuilder.WriteString("Original Issue/Task Description (from ClickUp/Jira/etc.):\n")
		} else {
			contextBuilder.WriteString("**Original Issue/Task (from ClickUp/Jira/etc.):**\n")
		}
		contextBuilder.WriteString("```\n")
		contextBuilder.WriteString(issueContext)
		contextBuilder.WriteString("\n```\n\n")
	}

	// Add commit information if available
	if len(gitResult.Commits) > 0 {
		if isOllama {
			contextBuilder.WriteString("Commit History:\n")
		} else {
			contextBuilder.WriteString("**Commit History:**\n")
		}

		for _, commit := range gitResult.Commits {
			if isOllama {
				contextBuilder.WriteString(fmt.Sprintf("- %s: %s (by %s on %s)\n",
					commit.ShortHash, commit.Message, commit.Author, commit.Date.Format("2006-01-02")))
			} else {
				contextBuilder.WriteString(fmt.Sprintf("- **%s**: %s\n", commit.ShortHash, commit.Message))
				contextBuilder.WriteString(fmt.Sprintf("  *By %s on %s*\n", commit.Author, commit.Date.Format("2006-01-02 15:04")))
			}
		}
		contextBuilder.WriteString("\n")
	}

	return contextBuilder.String()
}

// buildDiffSection creates the git diff section
func (g *Generator) buildDiffSection(gitResult *git.GitResult, isOllama bool) string {
	var diffBuilder strings.Builder

	if isOllama {
		diffBuilder.WriteString("Git diff:\n")
	} else {
		diffBuilder.WriteString("**Git diff to analyze:**\n")
	}

	diffBuilder.WriteString("```diff\n")
	diffBuilder.WriteString(gitResult.Diff)
	diffBuilder.WriteString("\n```\n\n")

	return diffBuilder.String()
}

// buildFileLinksGuideline creates file linking guidelines
func (g *Generator) buildFileLinksGuideline(repoURL, defaultBranch string) string {
	repoInfo := extractRepoInfo(repoURL)
	if repoInfo == "" {
		return ""
	}

	return fmt.Sprintf("- When referencing files, use actual repository URLs like: %s/blob/%s/path/to/file.ext\n",
		repoInfo, defaultBranch)
}

// buildOllamaPrompt creates a simplified prompt for Ollama structured outputs
func (g *Generator) buildOllamaPrompt(gitResult *git.GitResult, issueContext, repoURL string) string {
	var promptBuilder strings.Builder
	template := g.getPromptTemplate(true)

	// System role and basic instructions
	promptBuilder.WriteString(template.SystemRole)
	promptBuilder.WriteString(" ")
	promptBuilder.WriteString(template.BaseInstructions)
	promptBuilder.WriteString("\n\n")

	// Format instructions
	promptBuilder.WriteString("Generate a concise title with appropriate emoji and a detailed markdown description.\n\n")

	// Context section
	contextSection := g.buildContextSection(gitResult, issueContext, true)
	promptBuilder.WriteString(contextSection)

	// Diff section
	diffSection := g.buildDiffSection(gitResult, true)
	promptBuilder.WriteString(diffSection)

	// Format template
	promptBuilder.WriteString(template.FormatTemplate)
	promptBuilder.WriteString("\n")

	// Guidelines
	for _, guideline := range template.Guidelines {
		promptBuilder.WriteString("- ")
		promptBuilder.WriteString(guideline)
		promptBuilder.WriteString("\n")
	}

	// Restrictions
	for _, restriction := range template.Restrictions {
		promptBuilder.WriteString("- **")
		promptBuilder.WriteString(restriction)
		promptBuilder.WriteString("**\n")
	}

	// File links guideline
	fileLinksGuideline := g.buildFileLinksGuideline(repoURL, gitResult.DefaultBranch)
	if fileLinksGuideline != "" {
		promptBuilder.WriteString(fileLinksGuideline)
	}

	return promptBuilder.String()
}

// buildPrompt constructs the detailed prompt for OpenAI and other clients
func (g *Generator) buildPrompt(gitResult *git.GitResult, issueContext, repoURL string) string {
	var promptBuilder strings.Builder
	template := g.getPromptTemplate(false)

	// System role and instructions
	promptBuilder.WriteString(template.SystemRole)
	promptBuilder.WriteString(" ")
	promptBuilder.WriteString(template.BaseInstructions)
	promptBuilder.WriteString("\n\n")

	// Context section
	contextSection := g.buildContextSection(gitResult, issueContext, false)
	promptBuilder.WriteString(contextSection)

	// Diff section
	diffSection := g.buildDiffSection(gitResult, false)
	promptBuilder.WriteString(diffSection)

	// Format template
	promptBuilder.WriteString(template.FormatTemplate)
	promptBuilder.WriteString("\n\n")

	// Additional guidelines section
	promptBuilder.WriteString("**Additional Guidelines:**\n")

	// Emoji guidelines
	promptBuilder.WriteString("- ")
	promptBuilder.WriteString(g.getEmojiGuidelines())
	promptBuilder.WriteString("\n")

	// Regular guidelines
	for _, guideline := range template.Guidelines {
		promptBuilder.WriteString("- ")
		promptBuilder.WriteString(guideline)
		promptBuilder.WriteString("\n")
	}

	// Restrictions
	for _, restriction := range template.Restrictions {
		promptBuilder.WriteString("- **")
		promptBuilder.WriteString(restriction)
		promptBuilder.WriteString("**\n")
	}

	// File links guidelines
	fileLinksGuideline := g.buildFileLinksGuideline(repoURL, gitResult.DefaultBranch)
	if fileLinksGuideline != "" {
		promptBuilder.WriteString(fileLinksGuideline)
	}
	promptBuilder.WriteString("- Do NOT use placeholder URLs like 'your-repo', 'example-repo', or similar placeholders\n")
	promptBuilder.WriteString("- Use the actual repository information provided above when creating file links\n\n")

	// Final instruction
	promptBuilder.WriteString("**Analyze the git diff and create a professional PR description following this exact format.**")

	return promptBuilder.String()
}

// extractRepoInfo extracts repository URL information from a git repository URL
// and removes sensitive information like PAT tokens
func extractRepoInfo(repoURL string) string {
	if repoURL == "" {
		return ""
	}

	// Convert SSH to HTTPS format
	if strings.HasPrefix(repoURL, "git@github.com:") {
		repoURL = strings.Replace(repoURL, "git@github.com:", "https://github.com/", 1)
	}

	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Remove PAT (Personal Access Token) credentials from URL
	// Format: https://username:token@github.com/owner/repo
	// or: https://token@github.com/owner/repo
	if strings.Contains(repoURL, "@") {
		// Find the @ symbol and extract everything after it
		atIndex := strings.LastIndex(repoURL, "@")
		if atIndex != -1 {
			// Extract the protocol part (https://)
			protocolEnd := strings.Index(repoURL, "://")
			if protocolEnd != -1 {
				protocol := repoURL[:protocolEnd+3] // includes "://"
				hostAndPath := repoURL[atIndex+1:]  // everything after @
				repoURL = protocol + hostAndPath
			}
		}
	}

	return repoURL
}

// parseStructuredResponse parses the JSON response from Ollama structured output
func (g *Generator) parseStructuredResponse(response string) (*Result, error) {
	fmt.Printf("   📊 Structured response length: %d characters\n", len(response))

	var structuredResult struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	response = strings.TrimSpace(response)

	if err := json.Unmarshal([]byte(response), &structuredResult); err != nil {
		fmt.Printf("   ⚠️  Failed to parse structured response, trying fallback: %v\n", err)
		// Fallback to regular parsing if structured parsing fails
		return g.parseResponse(response)
	}

	fmt.Println("   ✅ Successfully parsed structured JSON response")

	return &Result{
		Title: cleanTitle(structuredResult.Title),
		Body:  structuredResult.Body,
	}, nil
}

// parseResponse extracts the title and body from the AI response
func (g *Generator) parseResponse(response string) (*Result, error) {
	fmt.Printf("   📊 Raw AI response length: %d characters\n", len(response))

	// Try to parse as JSON first (multiple attempts)
	var jsonResult struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	response = strings.TrimSpace(response)

	// Method 1: Look for JSON block with ```json markers
	if jsonStart := strings.Index(response, "```json"); jsonStart >= 0 {
		jsonStart += len("```json")
		if jsonEnd := strings.Index(response[jsonStart:], "```"); jsonEnd >= 0 {
			jsonStr := strings.TrimSpace(response[jsonStart : jsonStart+jsonEnd])
			if err := json.Unmarshal([]byte(jsonStr), &jsonResult); err == nil {
				fmt.Println("   ✅ Successfully parsed JSON from ```json block")
				return &Result{
					Title: cleanTitle(jsonResult.Title),
					Body:  jsonResult.Body,
				}, nil
			}
		}
	}

	// Method 2: Look for any JSON object
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &jsonResult); err == nil {
			fmt.Println("   ✅ Successfully parsed JSON object")
			return &Result{
				Title: cleanTitle(jsonResult.Title),
				Body:  jsonResult.Body,
			}, nil
		}
	}

	// Method 3: Try to extract from markdown structure
	if strings.Contains(response, "# ") {
		lines := strings.Split(response, "\n")
		var title, body string
		var bodyLines []string
		titleFound := false

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") && !titleFound {
				title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
				titleFound = true
				// Collect everything after the title as body
				if i+1 < len(lines) {
					bodyLines = lines[i+1:]
				}
				break
			}
		}

		if titleFound {
			body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
			fmt.Println("   ✅ Successfully parsed markdown structure")
			return &Result{
				Title: cleanTitle(title),
				Body:  body,
			}, nil
		}
	}

	// Method 4: Legacy TITLE:/BODY: format
	if strings.HasPrefix(response, "TITLE:") {
		lines := strings.Split(response, "\n")
		var title, body string

		bodyStart := false
		var bodyLines []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "TITLE:") {
				title = strings.TrimSpace(strings.TrimPrefix(line, "TITLE:"))
			} else if strings.HasPrefix(line, "BODY:") {
				bodyStart = true
			} else if bodyStart {
				bodyLines = append(bodyLines, line)
			}
		}

		body = strings.TrimSpace(strings.Join(bodyLines, "\n"))

		if title != "" {
			fmt.Println("   ✅ Successfully parsed TITLE:/BODY: format")
			return &Result{
				Title: cleanTitle(title),
				Body:  body,
			}, nil
		}
	}

	// Fallback: extract from unstructured response
	lines := strings.Split(strings.TrimSpace(response), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	// Find the first meaningful line as title
	title := ""
	bodyStartIndex := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && title == "" {
			title = line
			bodyStartIndex = i + 1
			break
		}
	}

	// Clean up title
	title = cleanTitle(title)

	// Get body from remaining lines
	var body string
	if bodyStartIndex < len(lines) {
		bodyLines := lines[bodyStartIndex:]
		body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}

	fmt.Println("   ⚠️  Used fallback parsing method")
	return &Result{
		Title: title,
		Body:  body,
	}, nil
}

// cleanTitle removes common prefixes and limits length
func cleanTitle(title string) string {
	// Remove common prefixes
	prefixes := []string{"Title:", "PR Title:", "Pull Request Title:", "**Title:**", "📋 **Title:**"}
	for _, prefix := range prefixes {
		title = strings.TrimPrefix(title, prefix)
	}

	title = strings.TrimSpace(title)

	// Remove markdown formatting
	title = strings.TrimPrefix(title, "**")
	title = strings.TrimSuffix(title, "**")
	title = strings.TrimSpace(title)

	// Limit title length
	if len(title) > 80 {
		title = title[:77] + "..."
	}

	return title
}
