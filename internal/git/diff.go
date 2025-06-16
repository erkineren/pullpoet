package git

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Client handles git operations
type Client struct{}

// NewClient creates a new git client
func NewClient() *Client {
	return &Client{}
}

// detectDefaultBranch tries to detect the default branch (main, master, dev, etc.)
func (c *Client) detectDefaultBranch(repo *git.Repository) string {
	// Common default branch names to try in order of preference
	defaultBranches := []string{"main", "master", "dev", "develop"}

	for _, branchName := range defaultBranches {
		_, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", branchName), true)
		if err == nil {
			return branchName
		}
	}

	// If none of the common branches exist, try to get the default from remote HEAD
	headRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", "HEAD"), true)
	if err == nil {
		// Extract branch name from symbolic ref
		if headRef.Type() == plumbing.SymbolicReference {
			target := headRef.Target()
			if target.IsRemote() {
				parts := strings.Split(target.String(), "/")
				if len(parts) >= 3 {
					return parts[len(parts)-1]
				}
			}
		}
	}

	// Fallback to "main"
	return "main"
}

// CommitInfo represents commit information
type CommitInfo struct {
	Hash      string
	Message   string
	Author    string
	Email     string
	Date      time.Time
	ShortHash string
}

// GitResult represents the result of git operations
type GitResult struct {
	Diff          string
	Commits       []CommitInfo
	DefaultBranch string
}

// GetDiff clones a repository and returns the diff between source and target branches
func (c *Client) GetDiff(repoURL, source, target string) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pullpoet-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clean branch names (remove origin/ prefix if present)
	sourceBranch := strings.TrimPrefix(source, "origin/")
	targetBranch := strings.TrimPrefix(target, "origin/")
	fmt.Printf("   🔧 Cleaned branch names: source='%s', target='%s'\n", sourceBranch, targetBranch)

	// Clone repository with optimization
	fmt.Printf("   📁 Creating temporary directory: %s\n", tempDir)
	fmt.Println("   🔄 Cloning repository (shallow clone for faster performance)...")
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:          repoURL,
		Depth:        1,     // Shallow clone - only get latest commit
		SingleBranch: false, // We need multiple branches
		NoCheckout:   true,  // Don't checkout files, we only need git data
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}
	fmt.Println("   ✅ Repository cloned successfully (optimized)")

	// Since shallow clone may not have all branches, fetch them specifically
	fmt.Printf("   🔄 Fetching required branches ('%s' and '%s')...\n", sourceBranch, targetBranch)

	// Get remote
	remote, err := repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %w", err)
	}

	// Fetch specific branches with shallow depth
	err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", targetBranch, targetBranch)),
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", sourceBranch, sourceBranch)),
		},
		Depth: 50, // Get enough history to find common ancestor
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("failed to fetch branches: %w", err)
	}
	fmt.Println("   ✅ Required branches fetched successfully")

	// Get target branch commit
	fmt.Printf("   🎯 Finding target branch '%s'...\n", targetBranch)
	targetRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", targetBranch), true)
	if err != nil {
		return "", fmt.Errorf("failed to find target branch '%s': %w", targetBranch, err)
	}

	targetCommit, err := repo.CommitObject(targetRef.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get target commit: %w", err)
	}
	fmt.Printf("   ✅ Target branch found: %s\n", targetRef.Hash().String()[:8])

	// Get source branch commit
	fmt.Printf("   🎯 Finding source branch '%s'...\n", sourceBranch)
	sourceRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", sourceBranch), true)
	if err != nil {
		return "", fmt.Errorf("failed to find source branch '%s': %w", sourceBranch, err)
	}

	sourceCommit, err := repo.CommitObject(sourceRef.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get source commit: %w", err)
	}
	fmt.Printf("   ✅ Source branch found: %s\n", sourceRef.Hash().String()[:8])

	// Get diff between commits
	fmt.Println("   📊 Generating patch diff...")
	patch, err := targetCommit.Patch(sourceCommit)
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	fmt.Printf("   ✅ Patch generated successfully\n")
	return patch.String(), nil
}

// GetCommitMessages returns commit messages between source and target branches
func (c *Client) GetCommitMessages(repoURL, source, target string) ([]string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pullpoet-commits-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get target branch commit
	targetRef, err := repo.Reference(plumbing.NewBranchReferenceName(target), true)
	if err != nil {
		return nil, fmt.Errorf("failed to find target branch '%s': %w", target, err)
	}

	// Get source branch commit
	sourceRef, err := repo.Reference(plumbing.NewBranchReferenceName(source), true)
	if err != nil {
		return nil, fmt.Errorf("failed to find source branch '%s': %w", source, err)
	}

	// Get commit iterator from source to target
	commits, err := repo.Log(&git.LogOptions{
		From: sourceRef.Hash(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var messages []string
	targetHash := targetRef.Hash()

	err = commits.ForEach(func(commit *object.Commit) error {
		if commit.Hash == targetHash {
			return fmt.Errorf("reached target commit") // Stop iteration
		}

		// Clean up commit message
		message := strings.TrimSpace(commit.Message)
		if message != "" {
			messages = append(messages, message)
		}
		return nil
	})

	// If we hit the target commit, that's expected
	if err != nil && !strings.Contains(err.Error(), "reached target commit") {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return messages, nil
}

// GetDiffWithCommits clones a repository and returns both diff and commit info between source and target branches
func (c *Client) GetDiffWithCommits(repoURL, source, target string) (*GitResult, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pullpoet-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clean branch names (remove origin/ prefix if present)
	sourceBranch := strings.TrimPrefix(source, "origin/")
	targetBranch := strings.TrimPrefix(target, "origin/")
	fmt.Printf("   🔧 Cleaned branch names: source='%s', target='%s'\n", sourceBranch, targetBranch)

	// Clone repository with optimization
	fmt.Printf("   📁 Creating temporary directory: %s\n", tempDir)
	fmt.Println("   🔄 Cloning repository (shallow clone for faster performance)...")
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:          repoURL,
		Depth:        50,    // Get more commits for commit history
		SingleBranch: false, // We need multiple branches
		NoCheckout:   true,  // Don't checkout files, we only need git data
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}
	fmt.Println("   ✅ Repository cloned successfully (optimized)")

	// Since shallow clone may not have all branches, fetch them specifically
	fmt.Printf("   🔄 Fetching required branches ('%s' and '%s')...\n", sourceBranch, targetBranch)

	// Get remote
	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, fmt.Errorf("failed to get remote: %w", err)
	}

	// Fetch specific branches with shallow depth
	err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", targetBranch, targetBranch)),
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", sourceBranch, sourceBranch)),
		},
		Depth: 100, // Get more history to find commits between branches
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, fmt.Errorf("failed to fetch branches: %w", err)
	}
	fmt.Println("   ✅ Required branches fetched successfully")

	// Detect default branch
	fmt.Println("   🔍 Detecting default branch...")
	defaultBranch := c.detectDefaultBranch(repo)
	fmt.Printf("   ✅ Default branch detected: %s\n", defaultBranch)

	// Get target branch commit
	fmt.Printf("   🎯 Finding target branch '%s'...\n", targetBranch)
	targetRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", targetBranch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to find target branch '%s': %w", targetBranch, err)
	}

	targetCommit, err := repo.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get target commit: %w", err)
	}
	fmt.Printf("   ✅ Target branch found: %s\n", targetRef.Hash().String()[:8])

	// Get source branch commit
	fmt.Printf("   🎯 Finding source branch '%s'...\n", sourceBranch)
	sourceRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", sourceBranch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to find source branch '%s': %w", sourceBranch, err)
	}

	sourceCommit, err := repo.CommitObject(sourceRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get source commit: %w", err)
	}
	fmt.Printf("   ✅ Source branch found: %s\n", sourceRef.Hash().String()[:8])

	// Get diff between commits
	fmt.Println("   📊 Generating patch diff...")
	patch, err := targetCommit.Patch(sourceCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate patch: %w", err)
	}

	// Get commit information from source branch that are ahead of target
	fmt.Println("   📝 Collecting commit information...")
	commits, err := c.getCommitsBetweenBranches(repo, sourceRef.Hash(), targetRef.Hash())
	if err != nil {
		fmt.Printf("   ⚠️  Warning: Failed to get commit info: %v\n", err)
		// Continue without commit info
		commits = []CommitInfo{}
	}

	fmt.Printf("   ✅ Found %d commits in source branch ahead of target\n", len(commits))
	fmt.Printf("   ✅ Git analysis completed successfully\n")

	return &GitResult{
		Diff:          patch.String(),
		Commits:       commits,
		DefaultBranch: defaultBranch,
	}, nil
}

// getCommitsBetweenBranches returns commits that are in source but not in target
func (c *Client) getCommitsBetweenBranches(repo *git.Repository, sourceHash, targetHash plumbing.Hash) ([]CommitInfo, error) {
	// Get commit iterator from source branch
	sourceCommit, err := repo.CommitObject(sourceHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get source commit: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{
		From: sourceHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var commits []CommitInfo
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Stop when we reach the target commit
		if commit.Hash == targetHash {
			return fmt.Errorf("reached target") // Signal to stop
		}

		// Check if this commit is an ancestor of target (already merged)
		isAncestor, err := commit.IsAncestor(sourceCommit)
		if err == nil && !isAncestor {
			// This commit is new in source branch
			commitInfo := CommitInfo{
				Hash:      commit.Hash.String(),
				ShortHash: commit.Hash.String()[:8],
				Message:   strings.TrimSpace(commit.Message),
				Author:    commit.Author.Name,
				Email:     commit.Author.Email,
				Date:      commit.Author.When,
			}
			commits = append(commits, commitInfo)
		}

		// Limit to prevent infinite iteration
		if len(commits) >= 20 {
			return fmt.Errorf("commit limit reached")
		}

		return nil
	})

	// Clean up expected errors
	if err != nil && !strings.Contains(err.Error(), "reached target") && !strings.Contains(err.Error(), "commit limit reached") {
		return nil, err
	}

	return commits, nil
}
