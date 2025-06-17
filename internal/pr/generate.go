package pr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// NewGenerator creates a new PR generator
func NewGenerator(aiClient ai.Client) *Generator {
	return &Generator{
		aiClient: aiClient,
	}
}

// Generate creates a PR description based on the git diff and optional description
func (g *Generator) Generate(gitResult *git.GitResult, issueContext, repoURL string) (*Result, error) {
	fmt.Println("   📝 Building unified AI prompt...")

	prompt, err := g.buildUnifiedPrompt(gitResult, issueContext, repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	fmt.Printf("   ✅ Unified prompt built (%d characters)\n", len(prompt))

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

	// Add pullpoet signature to the end of the PR body
	result.Body = g.addPullpoetSignature(result.Body)

	return result, nil
}

// loadPromptTemplate loads the unified prompt template from the .prompt file
func (g *Generator) loadPromptTemplate() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for .prompt file in the current directory or walk up to find it
	promptPath := filepath.Join(cwd, ".prompt")

	// If not found in current directory, try to find it in parent directories
	for {
		if _, err := os.Stat(promptPath); err == nil {
			break
		}

		parentDir := filepath.Dir(cwd)
		if parentDir == cwd {
			// Reached root directory
			return "", fmt.Errorf(".prompt file not found")
		}

		cwd = parentDir
		promptPath = filepath.Join(cwd, ".prompt")
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read .prompt file: %w", err)
	}

	return string(content), nil
}

// buildUnifiedPrompt constructs the prompt using the unified template
func (g *Generator) buildUnifiedPrompt(gitResult *git.GitResult, issueContext, repoURL string) (string, error) {
	// Load the base prompt template
	baseTemplate, err := g.loadPromptTemplate()
	if err != nil {
		return "", err
	}

	var promptBuilder strings.Builder

	// Add the base template
	promptBuilder.WriteString(baseTemplate)
	promptBuilder.WriteString("\n\n")

	// Add context section
	contextSection := g.buildContextSection(gitResult, issueContext)
	promptBuilder.WriteString(contextSection)

	// Add git diff section
	diffSection := g.buildDiffSection(gitResult)
	promptBuilder.WriteString(diffSection)

	// Add repository information if available
	if repoURL != "" {
		repoInfo := extractRepoInfo(repoURL)
		if repoInfo != "" {
			promptBuilder.WriteString(fmt.Sprintf("**Repository**: %s\n", repoInfo))
			promptBuilder.WriteString(fmt.Sprintf("**Default Branch**: %s\n\n", gitResult.DefaultBranch))
		}
	}

	// Add final instruction
	promptBuilder.WriteString("**Analyze the above information and create a professional PR description following the JSON format specified above.**")

	return promptBuilder.String(), nil
}

// buildContextSection creates the context section with issue and commit info
func (g *Generator) buildContextSection(gitResult *git.GitResult, issueContext string) string {
	var contextBuilder strings.Builder

	// Add issue context if provided
	if issueContext != "" {
		contextBuilder.WriteString("## 📋 Issue/Task Context\n\n")
		contextBuilder.WriteString("```\n")
		contextBuilder.WriteString(issueContext)
		contextBuilder.WriteString("\n```\n\n")
	}

	// Add commit information if available
	if len(gitResult.Commits) > 0 {
		contextBuilder.WriteString("## 📝 Commit History\n\n")

		for _, commit := range gitResult.Commits {
			contextBuilder.WriteString(fmt.Sprintf("- **%s**: %s\n", commit.ShortHash, commit.Message))
			contextBuilder.WriteString(fmt.Sprintf("  *By %s on %s*\n", commit.Author, commit.Date.Format("2006-01-02 15:04")))
		}
		contextBuilder.WriteString("\n")
	}

	return contextBuilder.String()
}

// buildDiffSection creates the git diff section
func (g *Generator) buildDiffSection(gitResult *git.GitResult) string {
	var diffBuilder strings.Builder

	diffBuilder.WriteString("## 🔍 Git Diff to Analyze\n\n")
	diffBuilder.WriteString("```diff\n")
	diffBuilder.WriteString(gitResult.Diff)
	diffBuilder.WriteString("\n```\n\n")

	return diffBuilder.String()
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

// addPullpoetSignature adds a footer indicating the PR was generated by pullpoet
func (g *Generator) addPullpoetSignature(body string) string {
	provider, model := g.aiClient.GetProviderInfo()
	signature := fmt.Sprintf("\n\n---\n\n*🤖 This PR description was generated by [pullpoet](https://github.com/erkineren/pullpoet) using %s (%s) - an AI-powered tool for creating professional pull request descriptions.*", provider, model)
	return body + signature
}

// parseResponse extracts the title and body from the AI response using multiple parsing strategies
func (g *Generator) parseResponse(response string) (*Result, error) {
	fmt.Printf("   📊 AI response length: %d characters\n", len(response))

	response = strings.TrimSpace(response)

	// Try to parse as JSON first (multiple methods)
	var jsonResult struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

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
