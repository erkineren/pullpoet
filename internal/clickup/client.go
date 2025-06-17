package clickup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// Task represents a simplified task structure for our use
type Task struct {
	ID          string
	Name        string
	Description string
	Status      string
	Creator     string
	URL         string
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

	return task, nil
}

// FormatTaskDescription formats the task information into a structured description
func (t *Task) FormatTaskDescription() string {
	return fmt.Sprintf(`**ClickUp Task: %s**

**Task ID:** %s
**Status:** %s
**Creator:** %s
**Task URL:** %s

**Description:**
%s`, t.Name, t.ID, t.Status, t.Creator, t.URL, t.Description)
}
