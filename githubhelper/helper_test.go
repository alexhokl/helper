package githubhelper

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty token returns error",
			token:       "",
			wantErr:     true,
			errContains: "GitHub token is not set",
		},
		{
			name:    "valid token creates client",
			token:   "ghp_test_token_12345",
			wantErr: false,
		},
		{
			name:    "token with spaces is accepted",
			token:   "ghp_token_with_content",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := NewClient(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("NewClient() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				if client != nil {
					t.Errorf("NewClient() expected nil client when error, got %v", client)
				}
			} else {
				if err != nil {
					t.Errorf("NewClient() unexpected error: %v", err)
				}
				if client == nil {
					t.Error("NewClient() expected non-nil client, got nil")
				}
			}
		})
	}
}

func TestNewClientWithContext(t *testing.T) {
	// Test with cancelled context - client should still be created
	// (context is used for HTTP client, not validation)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := NewClient(ctx, "test_token")
	if err != nil {
		t.Errorf("NewClient() with cancelled context returned error: %v", err)
	}
	if client == nil {
		t.Error("NewClient() with cancelled context returned nil client")
	}
}

func TestGithubIssueStruct(t *testing.T) {
	// Test creating and using the GithubIssue struct
	now := time.Now()
	issue := GithubIssue{
		ID:     "I_kwDOTest123",
		Number: 42,
		Title:  "Test Issue Title",
		Body:   "Test issue body content",
		URL:    "https://github.com/owner/repo/issues/42",
		DateFields: map[string]time.Time{
			"Due Date":   now,
			"Start Date": now.Add(-24 * time.Hour),
		},
	}

	if issue.ID != "I_kwDOTest123" {
		t.Errorf("GithubIssue.ID = %q, want %q", issue.ID, "I_kwDOTest123")
	}
	if issue.Number != 42 {
		t.Errorf("GithubIssue.Number = %d, want %d", issue.Number, 42)
	}
	if issue.Title != "Test Issue Title" {
		t.Errorf("GithubIssue.Title = %q, want %q", issue.Title, "Test Issue Title")
	}
	if issue.Body != "Test issue body content" {
		t.Errorf("GithubIssue.Body = %q, want %q", issue.Body, "Test issue body content")
	}
	if issue.URL != "https://github.com/owner/repo/issues/42" {
		t.Errorf("GithubIssue.URL = %q, want %q", issue.URL, "https://github.com/owner/repo/issues/42")
	}
	if len(issue.DateFields) != 2 {
		t.Errorf("GithubIssue.DateFields length = %d, want %d", len(issue.DateFields), 2)
	}
	if !issue.DateFields["Due Date"].Equal(now) {
		t.Errorf("GithubIssue.DateFields[\"Due Date\"] = %v, want %v", issue.DateFields["Due Date"], now)
	}
}

func TestGithubIssueStructWithEmptyDateFields(t *testing.T) {
	issue := GithubIssue{
		ID:         "I_kwDOTest456",
		Number:     1,
		Title:      "Minimal Issue",
		DateFields: make(map[string]time.Time),
	}

	if issue.DateFields == nil {
		t.Error("GithubIssue.DateFields should not be nil when initialized with make()")
	}
	if len(issue.DateFields) != 0 {
		t.Errorf("GithubIssue.DateFields length = %d, want %d", len(issue.DateFields), 0)
	}
}

func TestGithubIssueStructWithNilDateFields(t *testing.T) {
	issue := GithubIssue{
		ID:     "I_kwDOTest789",
		Number: 2,
		Title:  "Issue without DateFields",
	}

	if issue.DateFields != nil {
		t.Errorf("GithubIssue.DateFields = %v, want nil", issue.DateFields)
	}
}

// skipIfNoGitHubToken skips tests that require a real GitHub token
func skipIfNoGitHubToken(t *testing.T) string {
	t.Helper()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("skipping test: GITHUB_TOKEN not set")
	}
	// Skip if running in CI without valid token
	if os.Getenv("HELPER_SKIP_GITHUB_INTEGRATION") == "1" {
		t.Skip("skipping test: HELPER_SKIP_GITHUB_INTEGRATION=1")
	}
	return token
}

// isAuthError checks if the error is an authentication error
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "Bad credentials")
}

// Integration tests - these require GITHUB_TOKEN environment variable
// and make real API calls to GitHub

func TestGetIssueIntegration(t *testing.T) {
	token := skipIfNoGitHubToken(t)

	ctx := context.Background()
	client, err := NewClient(ctx, token)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	// Test with a well-known public issue (Go repository issue #1)
	issue, err := GetIssue(ctx, client, "golang", "go", 1)
	if err != nil {
		if isAuthError(err) {
			t.Skipf("skipping test: GitHub token invalid or expired: %v", err)
		}
		t.Fatalf("GetIssue() error: %v", err)
	}

	if issue == nil {
		t.Fatal("GetIssue() returned nil issue")
	}
	if issue.Number != 1 {
		t.Errorf("GetIssue() issue.Number = %d, want %d", issue.Number, 1)
	}
	if issue.Title == "" {
		t.Error("GetIssue() issue.Title is empty")
	}
	if issue.URL == "" {
		t.Error("GetIssue() issue.URL is empty")
	}
	if issue.DateFields == nil {
		t.Error("GetIssue() issue.DateFields is nil")
	}
}

func TestGetIssueNotFoundIntegration(t *testing.T) {
	token := skipIfNoGitHubToken(t)

	ctx := context.Background()
	client, err := NewClient(ctx, token)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	// Test with a non-existent issue number
	_, err = GetIssue(ctx, client, "golang", "go", 999999999)
	if err == nil {
		t.Error("GetIssue() with non-existent issue should return error")
	} else if isAuthError(err) {
		t.Skipf("skipping test: GitHub token invalid or expired: %v", err)
	}
}

func TestGetLabelIntegration(t *testing.T) {
	token := skipIfNoGitHubToken(t)

	ctx := context.Background()
	client, err := NewClient(ctx, token)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	// Test with a well-known label in Go repository
	labelID, err := GetLabel(ctx, client, "golang", "go", "Documentation")
	if err != nil {
		if isAuthError(err) {
			t.Skipf("skipping test: GitHub token invalid or expired: %v", err)
		}
		t.Fatalf("GetLabel() error: %v", err)
	}

	if labelID == "" {
		t.Error("GetLabel() returned empty label ID")
	}
}

func TestGetLabelNotFoundIntegration(t *testing.T) {
	token := skipIfNoGitHubToken(t)

	ctx := context.Background()
	client, err := NewClient(ctx, token)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	// Test with a non-existent label
	labelID, err := GetLabel(ctx, client, "golang", "go", "this-label-definitely-does-not-exist-12345")
	if err != nil {
		if isAuthError(err) {
			t.Skipf("skipping test: GitHub token invalid or expired: %v", err)
		}
		// Some APIs return error for missing labels
		t.Logf("GetLabel() returned error for non-existent label: %v", err)
	} else if labelID != "" {
		t.Errorf("GetLabel() for non-existent label returned ID: %q", labelID)
	}
}

// Benchmark tests
func BenchmarkNewClient(b *testing.B) {
	ctx := context.Background()
	token := "test_token_for_benchmark"

	for i := 0; i < b.N; i++ {
		_, _ = NewClient(ctx, token)
	}
}

func BenchmarkGithubIssueCreation(b *testing.B) {
	now := time.Now()

	for i := 0; i < b.N; i++ {
		_ = GithubIssue{
			ID:     "I_kwDOTest123",
			Number: 42,
			Title:  "Test Issue Title",
			Body:   "Test issue body content",
			URL:    "https://github.com/owner/repo/issues/42",
			DateFields: map[string]time.Time{
				"Due Date": now,
			},
		}
	}
}
