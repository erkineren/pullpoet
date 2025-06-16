package pr

import (
	"testing"
)

func TestExtractRepoInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal HTTPS URL",
			input:    "https://github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "SSH URL",
			input:    "git@github.com:owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "URL with PAT token",
			input:    "https://erkineren:ghp_xxxxxxxxxxxxxxxxxxxx@github.com/Generaxion-Dev/CRM.git",
			expected: "https://github.com/Generaxion-Dev/CRM",
		},
		{
			name:     "URL with token only",
			input:    "https://ghp_xxxxxxxxxxxxxxxxxxxx@github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "URL with username and password",
			input:    "https://username:password@github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "Empty URL",
			input:    "",
			expected: "",
		},
		{
			name:     "URL without .git suffix",
			input:    "https://token@github.com/owner/repo",
			expected: "https://github.com/owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoInfo(tt.input)
			if result != tt.expected {
				t.Errorf("extractRepoInfo(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
