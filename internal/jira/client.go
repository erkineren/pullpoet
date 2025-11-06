package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles Jira API operations
type Client struct {
	baseURL  string
	username string
	apiToken string
	client   *http.Client
}

// IssueResponse represents the response from Jira API for a single issue
type IssueResponse struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary     string      `json:"summary"`
		Description interface{} `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Creator struct {
			DisplayName  string `json:"displayName"`
			EmailAddress string `json:"emailAddress"`
		} `json:"creator"`
		Reporter struct {
			DisplayName  string `json:"displayName"`
			EmailAddress string `json:"emailAddress"`
		} `json:"reporter"`
	} `json:"fields"`
}

// CommentsResponse represents the response from Jira API for issue comments
type CommentsResponse struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
}

// Comment represents a single comment in Jira
type Comment struct {
	ID     string      `json:"id"`
	Body   interface{} `json:"body"`
	Author struct {
		DisplayName  string `json:"displayName"`
		EmailAddress string `json:"emailAddress"`
	} `json:"author"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// Issue represents a simplified issue structure for our use
type Issue struct {
	ID          string
	Key         string
	Summary     string
	Description string
	Status      string
	IssueType   string
	Creator     string
	Reporter    string
	URL         string
	Comments    []Comment
}

// NewClient creates a new Jira API client
func NewClient(baseURL, username, apiToken string) *Client {
	// Remove trailing slash from baseURL if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		baseURL:  baseURL,
		username: username,
		apiToken: apiToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetIssue fetches an issue by key from Jira
func (c *Client) GetIssue(issueKey string) (*Issue, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", c.baseURL, issueKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Basic Auth header
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.username, c.apiToken)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var issueResp IssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&issueResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our simplified Issue structure
	issue := &Issue{
		ID:          issueResp.ID,
		Key:         issueResp.Key,
		Summary:     issueResp.Fields.Summary,
		Description: extractText(issueResp.Fields.Description),
		Status:      issueResp.Fields.Status.Name,
		IssueType:   issueResp.Fields.IssueType.Name,
		Creator:     issueResp.Fields.Creator.DisplayName,
		Reporter:    issueResp.Fields.Reporter.DisplayName,
		URL:         fmt.Sprintf("%s/browse/%s", c.baseURL, issueResp.Key),
	}

	// Fetch comments for the issue
	comments, err := c.GetIssueComments(issueKey)
	if err != nil {
		// Log the error but don't fail the entire request
		fmt.Printf("Warning: Failed to fetch comments for issue %s: %v\n", issueKey, err)
	} else {
		issue.Comments = comments
	}

	return issue, nil
}

// GetIssueComments fetches comments for an issue by key from Jira
func (c *Client) GetIssueComments(issueKey string) ([]Comment, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", c.baseURL, issueKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Basic Auth header
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.username, c.apiToken)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var commentResp CommentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&commentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return commentResp.Comments, nil
}

// extractText extracts plain text from Jira's Atlassian Document Format (ADF)
// This is a simplified version that handles the most common cases
func extractText(data interface{}) string {
	if data == nil {
		return ""
	}

	// If it's a string, return it directly
	if str, ok := data.(string); ok {
		return str
	}

	// If it's a map (ADF format), extract text recursively
	if adf, ok := data.(map[string]interface{}); ok {
		var builder strings.Builder
		extractTextFromADF(adf, &builder)
		return builder.String()
	}

	return ""
}

// extractTextFromADF recursively extracts text from Atlassian Document Format
func extractTextFromADF(node map[string]interface{}, builder *strings.Builder) {
	// Check for text content
	if text, ok := node["text"].(string); ok {
		builder.WriteString(text)
	}

	// Process content array (paragraphs, lists, etc.)
	if content, ok := node["content"].([]interface{}); ok {
		for _, item := range content {
			if contentNode, ok := item.(map[string]interface{}); ok {
				// Add newlines between block elements
				nodeType, _ := contentNode["type"].(string)
				if nodeType == "paragraph" || nodeType == "heading" {
					extractTextFromADF(contentNode, builder)
					builder.WriteString("\n")
				} else {
					extractTextFromADF(contentNode, builder)
				}
			}
		}
	}
}

// FormatIssueDescription formats the issue information into a structured description
func (i *Issue) FormatIssueDescription() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`**Jira Issue: %s - %s**

**Issue Key:** %s
**Issue Type:** %s
**Status:** %s
**Creator:** %s
**Reporter:** %s
**Issue URL:** %s

**Description:**
%s`, i.Summary, i.Key, i.Key, i.IssueType, i.Status, i.Creator, i.Reporter, i.URL, i.Description))

	// Add comments if available
	if len(i.Comments) > 0 {
		builder.WriteString("\n\n**Comments:**\n")
		for idx, comment := range i.Comments {
			builder.WriteString(fmt.Sprintf("\n---\n"))
			builder.WriteString(fmt.Sprintf("**Comment %d** (by %s on %s):\n", idx+1, comment.Author.DisplayName, comment.Created))

			// Extract text from comment body (ADF format)
			commentText := extractText(comment.Body)
			builder.WriteString(commentText)
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
