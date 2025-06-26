package clickup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles ClickUp API operations
type Client struct {
	baseURL string
	pat     string
	client  *http.Client
}

// TaskResponse represents the response from ClickUp API for a single task
type TaskResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TextContent string `json:"text_content"`
	Status      struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Color  string `json:"color"`
		Type   string `json:"type"`
	} `json:"status"`
	Creator struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"creator"`
	URL string `json:"url"`
}

// CommentResponse represents the response from ClickUp API for task comments
type CommentResponse struct {
	Comments []Comment `json:"comments"`
}

// Comment represents a single comment in ClickUp
type Comment struct {
	ID      string `json:"id"`
	Comment []struct {
		Text string `json:"text"`
	} `json:"comment"`
	User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user"`
	DateCreated string    `json:"date_created"`
	ReplyCount  int       `json:"reply_count"`
	Replies     []Comment `json:"replies,omitempty"`
}

// Task represents a simplified task structure for our use
type Task struct {
	ID          string
	Name        string
	Description string
	Status      string
	Creator     string
	URL         string
	Comments    []Comment
}

// NewClient creates a new ClickUp API client
func NewClient(pat string) *Client {
	return &Client{
		baseURL: "https://api.clickup.com/api/v2",
		pat:     pat,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetTask fetches a task by ID from ClickUp
func (c *Client) GetTask(taskID string) (*Task, error) {
	url := fmt.Sprintf("%s/task/%s", c.baseURL, taskID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.pat)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ClickUp API error (status %d): %s", resp.StatusCode, string(body))
	}

	var taskResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our simplified Task structure
	task := &Task{
		ID:          taskResp.ID,
		Name:        taskResp.Name,
		Description: taskResp.Description,
		Status:      taskResp.Status.Status,
		Creator:     taskResp.Creator.Username,
		URL:         taskResp.URL,
	}

	// If description is empty, use text_content as fallback
	if task.Description == "" && taskResp.TextContent != "" {
		task.Description = taskResp.TextContent
	}

	// Fetch comments for the task
	comments, err := c.GetTaskComments(taskID)
	if err != nil {
		// Log the error but don't fail the entire request
		fmt.Printf("Warning: Failed to fetch comments for task %s: %v\n", taskID, err)
	} else {
		task.Comments = comments
	}

	return task, nil
}

// GetTaskComments fetches comments for a task by ID from ClickUp
func (c *Client) GetTaskComments(taskID string) ([]Comment, error) {
	url := fmt.Sprintf("%s/task/%s/comment", c.baseURL, taskID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.pat)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ClickUp API error (status %d): %s", resp.StatusCode, string(body))
	}

	var commentResp CommentResponse
	if err := json.NewDecoder(resp.Body).Decode(&commentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Fetch replies for each comment that has replies
	for i := range commentResp.Comments {
		if commentResp.Comments[i].ReplyCount > 0 {
			replies, err := c.GetCommentReplies(taskID, commentResp.Comments[i].ID)
			if err != nil {
				// Log the error but don't fail the entire request
				fmt.Printf("Warning: Failed to fetch replies for comment %s: %v\n", commentResp.Comments[i].ID, err)
			} else {
				commentResp.Comments[i].Replies = replies
			}
		}
	}

	return commentResp.Comments, nil
}

// GetCommentReplies fetches replies for a specific comment
func (c *Client) GetCommentReplies(taskID, commentID string) ([]Comment, error) {
	url := fmt.Sprintf("%s/comment/%s/reply", c.baseURL, commentID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.pat)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ClickUp API error (status %d): %s", resp.StatusCode, string(body))
	}

	var commentResp CommentResponse
	if err := json.NewDecoder(resp.Body).Decode(&commentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return commentResp.Comments, nil
}

// FormatTaskDescription formats the task information into a structured description
func (t *Task) FormatTaskDescription() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`**ClickUp Task: %s**

**Task ID:** %s
**Status:** %s
**Creator:** %s
**Task URL:** %s

**Description:**
%s`, t.Name, t.ID, t.Status, t.Creator, t.URL, t.Description))

	// Add comments if available
	if len(t.Comments) > 0 {
		builder.WriteString("\n\n**Comments:**\n")
		for i, comment := range t.Comments {
			builder.WriteString(fmt.Sprintf("\n---\n"))
			builder.WriteString(fmt.Sprintf("**Comment %d** (by %s on %s):\n", i+1, comment.User.Username, comment.DateCreated))

			// Extract text from comment blocks
			for _, commentBlock := range comment.Comment {
				builder.WriteString(commentBlock.Text)
				builder.WriteString("\n")
			}

			// Add replies if available
			if len(comment.Replies) > 0 {
				builder.WriteString(fmt.Sprintf("\n  **Replies:**\n"))
				for j, reply := range comment.Replies {
					builder.WriteString(fmt.Sprintf("  **Reply %d** (by %s on %s):\n", j+1, reply.User.Username, reply.DateCreated))

					// Extract text from reply blocks
					for _, replyBlock := range reply.Comment {
						builder.WriteString("    " + replyBlock.Text)
						builder.WriteString("\n")
					}
					builder.WriteString("\n")
				}
			}
		}
	}

	return builder.String()
}
