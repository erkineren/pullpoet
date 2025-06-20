package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// FastClient uses native git commands for maximum speed
type FastClient struct{}

// NewFastClient creates a new fast git client that uses native git commands
func NewFastClient() *FastClient {
	return &FastClient{}
}

// detectDefaultBranchFast uses native git commands to detect the default branch
func (c *FastClient) detectDefaultBranchFast(tempDir string) string {
	// Try to get remote HEAD symbolic reference
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = tempDir
	output, err := cmd.Output()
	if err == nil {
		// Extract branch name from refs/remotes/origin/branch-name
		ref := strings.TrimSpace(string(output))
		parts := strings.Split(ref, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	// Fallback: try common branch names
	defaultBranches := []string{"main", "master", "dev", "develop"}
	for _, branchName := range defaultBranches {
		cmd := exec.Command("git", "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/remotes/origin/%s", branchName))
		cmd.Dir = tempDir
		if cmd.Run() == nil {
			return branchName
		}
	}

	// Final fallback
	return "main"
}

// GetDiffFast uses native git commands for maximum performance
func (c *FastClient) GetDiffFast(repoURL, source, target string) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pullpoet-fast-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("   📁 Creating temporary directory: %s\n", tempDir)

	// Clean branch names (remove origin/ prefix if present)
	sourceBranch := strings.TrimPrefix(source, "origin/")
	targetBranch := strings.TrimPrefix(target, "origin/")

	fmt.Printf("   🔧 Cleaned branch names: source='%s', target='%s'\n", sourceBranch, targetBranch)

	// Initialize git repository
	fmt.Println("   ⚡ Initializing git repository...")
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to init git repository: %w", err)
	}

	// Add remote
	fmt.Println("   🔗 Adding remote origin...")
	cmd = exec.Command("git", "remote", "add", "origin", repoURL)
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to add remote: %w", err)
	}

	// Fetch only the specific branches with minimal depth
	fmt.Printf("   🚀 Fast fetching branches '%s' and '%s' (depth: 50)...\n", sourceBranch, targetBranch)
	cmd = exec.Command("git", "fetch", "origin",
		fmt.Sprintf("%s:%s", sourceBranch, sourceBranch),
		fmt.Sprintf("%s:%s", targetBranch, targetBranch),
		"--depth=50")
	cmd.Dir = tempDir

	// Get detailed error output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("   ❌ Fetch failed. Git output: %s\n", string(output))
		return "", fmt.Errorf("failed to fetch branches: %w", err)
	}
	fmt.Println("   ✅ Branches fetched successfully")

	// Generate diff using native git
	fmt.Println("   📊 Generating diff using native git...")
	cmd = exec.Command("git", "diff", targetBranch, sourceBranch)
	cmd.Dir = tempDir

	diffOutput, err := cmd.Output()
	if err != nil {
		// Try alternative diff approach if direct diff fails
		fmt.Println("   🔄 Trying alternative diff approach...")
		cmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s", targetBranch), fmt.Sprintf("origin/%s", sourceBranch))
		cmd.Dir = tempDir
		diffOutput, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to generate diff: %w", err)
		}
	}

	fmt.Printf("   ✅ Diff generated successfully (%d characters)\n", len(diffOutput))
	return string(diffOutput), nil
}

// GetDiff wraps the fast implementation to maintain interface compatibility
func (c *FastClient) GetDiff(repoURL, source, target string) (string, error) {
	return c.GetDiffFast(repoURL, source, target)
}

// GetDiffWithCommits uses native git commands to get both diff and commit information
func (c *FastClient) GetDiffWithCommits(repoURL, source, target string) (*GitResult, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pullpoet-fast-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("   📁 Creating temporary directory: %s\n", tempDir)

	// Clean branch names (remove origin/ prefix if present)
	sourceBranch := strings.TrimPrefix(source, "origin/")
	targetBranch := strings.TrimPrefix(target, "origin/")

	fmt.Printf("   🔧 Cleaned branch names: source='%s', target='%s'\n", sourceBranch, targetBranch)

	// Initialize git repository
	fmt.Println("   ⚡ Initializing git repository...")
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to init git repository: %w", err)
	}

	// Add remote
	fmt.Println("   🔗 Adding remote origin...")
	cmd = exec.Command("git", "remote", "add", "origin", repoURL)
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to add remote: %w", err)
	}

	// Fetch only the specific branches with more depth for commit history
	fmt.Printf("   🚀 Fast fetching branches '%s' and '%s' (depth: 100)...\n", sourceBranch, targetBranch)
	cmd = exec.Command("git", "fetch", "origin",
		fmt.Sprintf("%s:%s", sourceBranch, sourceBranch),
		fmt.Sprintf("%s:%s", targetBranch, targetBranch),
		"--depth=100")
	cmd.Dir = tempDir

	// Get detailed error output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("   ❌ Fetch failed. Git output: %s\n", string(output))
		return nil, fmt.Errorf("failed to fetch branches: %w", err)
	}
	fmt.Println("   ✅ Branches fetched successfully")

	// Generate diff using native git
	fmt.Println("   📊 Generating diff using native git...")
	cmd = exec.Command("git", "diff", targetBranch, sourceBranch)
	cmd.Dir = tempDir

	diffOutput, err := cmd.Output()
	if err != nil {
		// Try alternative diff approach if direct diff fails
		fmt.Println("   🔄 Trying alternative diff approach...")
		cmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s", targetBranch), fmt.Sprintf("origin/%s", sourceBranch))
		cmd.Dir = tempDir
		diffOutput, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to generate diff: %w", err)
		}
	}

	fmt.Printf("   ✅ Diff generated successfully (%d characters)\n", len(diffOutput))

	// Get commit information
	fmt.Println("   📝 Collecting commit information...")
	commits, err := c.getCommitsBetweenBranchesFast(tempDir, sourceBranch, targetBranch)
	if err != nil {
		fmt.Printf("   ⚠️  Warning: Failed to get commit info: %v\n", err)
		// Continue without commit info
		commits = []CommitInfo{}
	}

	fmt.Printf("   ✅ Found %d commits in source branch ahead of target\n", len(commits))
	fmt.Printf("   ✅ Git analysis completed successfully\n")

	// Detect default branch using git command
	fmt.Println("   🔍 Detecting default branch...")
	defaultBranch := c.detectDefaultBranchFast(tempDir)
	fmt.Printf("   ✅ Default branch detected: %s\n", defaultBranch)

	return &GitResult{
		Diff:          string(diffOutput),
		Commits:       commits,
		DefaultBranch: defaultBranch,
	}, nil
}

// getCommitsBetweenBranchesFast uses native git to get commits between branches
func (c *FastClient) getCommitsBetweenBranchesFast(tempDir, sourceBranch, targetBranch string) ([]CommitInfo, error) {
	// Get commits that are in source but not in target
	// Format: hash|subject|author name|author email|author date
	cmd := exec.Command("git", "log", fmt.Sprintf("%s..%s", targetBranch, sourceBranch),
		"--pretty=format:%H|%s|%an|%ae|%ai", "--max-count=20")
	cmd.Dir = tempDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	if len(output) == 0 {
		return []CommitInfo{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var commits []CommitInfo

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 5 {
			continue // Skip malformed lines
		}

		// Parse date
		date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[4])
		if err != nil {
			// Try alternative format
			date, err = time.Parse("2006-01-02T15:04:05-07:00", parts[4])
			if err != nil {
				date = time.Now() // Fallback
			}
		}

		commitInfo := CommitInfo{
			Hash:      parts[0],
			ShortHash: parts[0][:8],
			Message:   parts[1],
			Author:    parts[2],
			Email:     parts[3],
			Date:      date,
		}
		commits = append(commits, commitInfo)
	}

	return commits, nil
}

// GetStagedDiff gets the diff of staged changes (git add'd files)
func (c *FastClient) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	return string(output), nil
}
