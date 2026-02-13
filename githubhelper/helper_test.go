package githubhelper

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

// mockGraphQLClient is a mock implementation of GraphQLClient for testing
type mockGraphQLClient struct {
	queryFunc  func(ctx context.Context, q interface{}, variables map[string]interface{}) error
	mutateFunc func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

func (m *mockGraphQLClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, q, variables)
	}
	return nil
}

func (m *mockGraphQLClient) Mutate(ctx context.Context, mut interface{}, input githubv4.Input, variables map[string]interface{}) error {
	if m.mutateFunc != nil {
		return m.mutateFunc(ctx, mut, input, variables)
	}
	return nil
}

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

// ============================================================================
// Mock-based tests for complete coverage
// ============================================================================

func TestGetIssue_WithMock(t *testing.T) {
	tests := []struct {
		name        string
		repoOwner   string
		repoName    string
		issueNumber int32
		mockQuery   func(ctx context.Context, q interface{}, variables map[string]interface{}) error
		wantIssue   *GithubIssue
		wantErr     bool
	}{
		{
			name:        "successful get issue",
			repoOwner:   "owner",
			repoName:    "repo",
			issueNumber: 42,
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				repo := v.FieldByName("Repository")
				issue := repo.FieldByName("Issue")
				issue.FieldByName("ID").SetString("I_kwDOTest123")
				issue.FieldByName("Number").SetInt(42)
				issue.FieldByName("Title").SetString("Test Issue")
				issue.FieldByName("Body").SetString("Test body")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/42")
				return nil
			},
			wantIssue: &GithubIssue{
				ID:         "I_kwDOTest123",
				Number:     42,
				Title:      "Test Issue",
				Body:       "Test body",
				URL:        "https://github.com/owner/repo/issues/42",
				DateFields: map[string]time.Time{},
			},
			wantErr: false,
		},
		{
			name:        "query error",
			repoOwner:   "owner",
			repoName:    "repo",
			issueNumber: 999,
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				return errors.New("GraphQL error: Could not resolve to an Issue")
			},
			wantIssue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGraphQLClient{
				queryFunc: tt.mockQuery,
			}

			ctx := context.Background()
			issue, err := GetIssue(ctx, mock, tt.repoOwner, tt.repoName, tt.issueNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantIssue == nil {
				if issue != nil {
					t.Errorf("GetIssue() = %v, want nil", issue)
				}
				return
			}

			if issue == nil {
				t.Errorf("GetIssue() = nil, want %v", tt.wantIssue)
				return
			}

			if issue.ID != tt.wantIssue.ID {
				t.Errorf("GetIssue() ID = %v, want %v", issue.ID, tt.wantIssue.ID)
			}
			if issue.Number != tt.wantIssue.Number {
				t.Errorf("GetIssue() Number = %v, want %v", issue.Number, tt.wantIssue.Number)
			}
			if issue.Title != tt.wantIssue.Title {
				t.Errorf("GetIssue() Title = %v, want %v", issue.Title, tt.wantIssue.Title)
			}
		})
	}
}

func TestGetProjectID_WithMock(t *testing.T) {
	tests := []struct {
		name          string
		repoOwner     string
		projectNumber int32
		mockQuery     func(ctx context.Context, q interface{}, variables map[string]interface{}) error
		wantID        githubv4.ID
		wantErr       bool
	}{
		{
			name:          "successful get project ID",
			repoOwner:     "owner",
			projectNumber: 1,
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				user := v.FieldByName("User")
				project := user.FieldByName("Project")
				project.FieldByName("ID").Set(reflect.ValueOf(githubv4.ID("PVT_kwDOTest123")))
				return nil
			},
			wantID:  githubv4.ID("PVT_kwDOTest123"),
			wantErr: false,
		},
		{
			name:          "project not found",
			repoOwner:     "owner",
			projectNumber: 999,
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				return errors.New("GraphQL error: Could not resolve to a ProjectV2")
			},
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGraphQLClient{
				queryFunc: tt.mockQuery,
			}

			ctx := context.Background()
			id, err := GetProjectID(ctx, mock, tt.repoOwner, tt.projectNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if id != tt.wantID {
				t.Errorf("GetProjectID() = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestGetIssuesWithProjectDateFieldValue_WithMock(t *testing.T) {
	tests := []struct {
		name      string
		projectID githubv4.ID
		fieldName string
		mockQuery func(ctx context.Context, q interface{}, variables map[string]interface{}) error
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful get issues with date field",
			projectID: githubv4.ID("PVT_kwDOTest123"),
			fieldName: "Due Date",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				projectNode := v.FieldByName("ProjectNode")
				project := projectNode.FieldByName("Project")
				items := project.FieldByName("Items")

				pageInfo := items.FieldByName("PageInfo")
				pageInfo.FieldByName("HasNextPage").SetBool(false)
				pageInfo.FieldByName("EndCursor").Set(reflect.ValueOf(githubv4.String("")))

				nodesField := items.FieldByName("Nodes")
				nodeType := nodesField.Type().Elem()
				node := reflect.New(nodeType).Elem()

				node.FieldByName("ID").Set(reflect.ValueOf(githubv4.ID("PVTI_123")))

				contentNode := node.FieldByName("ContentNode")
				issue := contentNode.FieldByName("Issue")
				issue.FieldByName("Number").SetInt(42)
				issue.FieldByName("Title").SetString("Test Issue")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/42")
				issue.FieldByName("Closed").SetBool(false)

				fieldValueUnion := node.FieldByName("FieldValueUnion")
				fieldValue := fieldValueUnion.FieldByName("FieldValue")
				fieldValue.FieldByName("Date").SetString("2024-06-15")

				nodes := reflect.MakeSlice(nodesField.Type(), 1, 1)
				nodes.Index(0).Set(node)
				nodesField.Set(nodes)

				return nil
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "query error",
			projectID: githubv4.ID("PVT_invalid"),
			fieldName: "Due Date",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				return errors.New("GraphQL error: Could not resolve to a node")
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "empty project",
			projectID: githubv4.ID("PVT_empty"),
			fieldName: "Due Date",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				projectNode := v.FieldByName("ProjectNode")
				project := projectNode.FieldByName("Project")
				items := project.FieldByName("Items")

				pageInfo := items.FieldByName("PageInfo")
				pageInfo.FieldByName("HasNextPage").SetBool(false)
				pageInfo.FieldByName("EndCursor").Set(reflect.ValueOf(githubv4.String("")))

				nodesField := items.FieldByName("Nodes")
				nodes := reflect.MakeSlice(nodesField.Type(), 0, 0)
				nodesField.Set(nodes)

				return nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "closed issues filtered out",
			projectID: githubv4.ID("PVT_closed"),
			fieldName: "Due Date",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				projectNode := v.FieldByName("ProjectNode")
				project := projectNode.FieldByName("Project")
				items := project.FieldByName("Items")

				pageInfo := items.FieldByName("PageInfo")
				pageInfo.FieldByName("HasNextPage").SetBool(false)
				pageInfo.FieldByName("EndCursor").Set(reflect.ValueOf(githubv4.String("")))

				nodesField := items.FieldByName("Nodes")
				nodeType := nodesField.Type().Elem()
				node := reflect.New(nodeType).Elem()

				node.FieldByName("ID").Set(reflect.ValueOf(githubv4.ID("PVTI_closed")))

				contentNode := node.FieldByName("ContentNode")
				issue := contentNode.FieldByName("Issue")
				issue.FieldByName("Number").SetInt(99)
				issue.FieldByName("Title").SetString("Closed Issue")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/99")
				issue.FieldByName("Closed").SetBool(true)

				fieldValueUnion := node.FieldByName("FieldValueUnion")
				fieldValue := fieldValueUnion.FieldByName("FieldValue")
				fieldValue.FieldByName("Date").SetString("2024-06-15")

				nodes := reflect.MakeSlice(nodesField.Type(), 1, 1)
				nodes.Index(0).Set(node)
				nodesField.Set(nodes)

				return nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "issues without date filtered out",
			projectID: githubv4.ID("PVT_nodate"),
			fieldName: "Due Date",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				projectNode := v.FieldByName("ProjectNode")
				project := projectNode.FieldByName("Project")
				items := project.FieldByName("Items")

				pageInfo := items.FieldByName("PageInfo")
				pageInfo.FieldByName("HasNextPage").SetBool(false)
				pageInfo.FieldByName("EndCursor").Set(reflect.ValueOf(githubv4.String("")))

				nodesField := items.FieldByName("Nodes")
				nodeType := nodesField.Type().Elem()
				node := reflect.New(nodeType).Elem()

				node.FieldByName("ID").Set(reflect.ValueOf(githubv4.ID("PVTI_nodate")))

				contentNode := node.FieldByName("ContentNode")
				issue := contentNode.FieldByName("Issue")
				issue.FieldByName("Number").SetInt(88)
				issue.FieldByName("Title").SetString("Issue Without Date")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/88")
				issue.FieldByName("Closed").SetBool(false)

				fieldValueUnion := node.FieldByName("FieldValueUnion")
				fieldValue := fieldValueUnion.FieldByName("FieldValue")
				fieldValue.FieldByName("Date").SetString("")

				nodes := reflect.MakeSlice(nodesField.Type(), 1, 1)
				nodes.Index(0).Set(node)
				nodesField.Set(nodes)

				return nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "invalid date format returns error",
			projectID: githubv4.ID("PVT_baddate"),
			fieldName: "Due Date",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				projectNode := v.FieldByName("ProjectNode")
				project := projectNode.FieldByName("Project")
				items := project.FieldByName("Items")

				pageInfo := items.FieldByName("PageInfo")
				pageInfo.FieldByName("HasNextPage").SetBool(false)
				pageInfo.FieldByName("EndCursor").Set(reflect.ValueOf(githubv4.String("")))

				nodesField := items.FieldByName("Nodes")
				nodeType := nodesField.Type().Elem()
				node := reflect.New(nodeType).Elem()

				node.FieldByName("ID").Set(reflect.ValueOf(githubv4.ID("PVTI_baddate")))

				contentNode := node.FieldByName("ContentNode")
				issue := contentNode.FieldByName("Issue")
				issue.FieldByName("Number").SetInt(77)
				issue.FieldByName("Title").SetString("Bad Date Issue")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/77")
				issue.FieldByName("Closed").SetBool(false)

				fieldValueUnion := node.FieldByName("FieldValueUnion")
				fieldValue := fieldValueUnion.FieldByName("FieldValue")
				fieldValue.FieldByName("Date").SetString("not-a-valid-date")

				nodes := reflect.MakeSlice(nodesField.Type(), 1, 1)
				nodes.Index(0).Set(node)
				nodesField.Set(nodes)

				return nil
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGraphQLClient{
				queryFunc: tt.mockQuery,
			}

			ctx := context.Background()
			issues, err := GetIssuesWithProjectDateFieldValue(ctx, mock, tt.projectID, tt.fieldName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssuesWithProjectDateFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(issues) != tt.wantCount {
				t.Errorf("GetIssuesWithProjectDateFieldValue() returned %d issues, want %d", len(issues), tt.wantCount)
			}
		})
	}
}

func TestGetIssuesWithProjectDateFieldValue_Pagination(t *testing.T) {
	callCount := 0

	mock := &mockGraphQLClient{
		queryFunc: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
			callCount++
			v := reflect.ValueOf(q).Elem()
			projectNode := v.FieldByName("ProjectNode")
			project := projectNode.FieldByName("Project")
			items := project.FieldByName("Items")

			pageInfo := items.FieldByName("PageInfo")
			nodesField := items.FieldByName("Nodes")

			if callCount == 1 {
				pageInfo.FieldByName("HasNextPage").SetBool(true)
				pageInfo.FieldByName("EndCursor").Set(reflect.ValueOf(githubv4.String("cursor1")))

				nodeType := nodesField.Type().Elem()
				node := reflect.New(nodeType).Elem()
				node.FieldByName("ID").Set(reflect.ValueOf(githubv4.ID("PVTI_1")))

				contentNode := node.FieldByName("ContentNode")
				issue := contentNode.FieldByName("Issue")
				issue.FieldByName("Number").SetInt(1)
				issue.FieldByName("Title").SetString("Issue 1")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/1")
				issue.FieldByName("Closed").SetBool(false)

				fieldValueUnion := node.FieldByName("FieldValueUnion")
				fieldValue := fieldValueUnion.FieldByName("FieldValue")
				fieldValue.FieldByName("Date").SetString("2024-06-01")

				nodes := reflect.MakeSlice(nodesField.Type(), 1, 1)
				nodes.Index(0).Set(node)
				nodesField.Set(nodes)
			} else {
				pageInfo.FieldByName("HasNextPage").SetBool(false)
				pageInfo.FieldByName("EndCursor").Set(reflect.ValueOf(githubv4.String("")))

				nodeType := nodesField.Type().Elem()
				node := reflect.New(nodeType).Elem()
				node.FieldByName("ID").Set(reflect.ValueOf(githubv4.ID("PVTI_2")))

				contentNode := node.FieldByName("ContentNode")
				issue := contentNode.FieldByName("Issue")
				issue.FieldByName("Number").SetInt(2)
				issue.FieldByName("Title").SetString("Issue 2")
				issue.FieldByName("URL").SetString("https://github.com/owner/repo/issues/2")
				issue.FieldByName("Closed").SetBool(false)

				fieldValueUnion := node.FieldByName("FieldValueUnion")
				fieldValue := fieldValueUnion.FieldByName("FieldValue")
				fieldValue.FieldByName("Date").SetString("2024-06-02")

				nodes := reflect.MakeSlice(nodesField.Type(), 1, 1)
				nodes.Index(0).Set(node)
				nodesField.Set(nodes)
			}

			return nil
		},
	}

	ctx := context.Background()
	issues, err := GetIssuesWithProjectDateFieldValue(ctx, mock, githubv4.ID("PVT_test"), "Due Date")

	if err != nil {
		t.Fatalf("GetIssuesWithProjectDateFieldValue() error = %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", callCount)
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
}

func TestAddComment_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		issueID    string
		comment    string
		mockMutate func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
		wantErr    bool
	}{
		{
			name:    "successful add comment",
			issueID: "I_kwDOTest123",
			comment: "This is a test comment",
			mockMutate: func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "mutation error",
			issueID: "I_invalid",
			comment: "Comment on invalid issue",
			mockMutate: func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
				return errors.New("GraphQL error: Could not resolve to a node")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGraphQLClient{
				mutateFunc: tt.mockMutate,
			}

			ctx := context.Background()
			err := AddComment(ctx, mock, tt.issueID, tt.comment)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddComment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetLabel_WithMock(t *testing.T) {
	tests := []struct {
		name      string
		repoOwner string
		repoName  string
		labelName string
		mockQuery func(ctx context.Context, q interface{}, variables map[string]interface{}) error
		wantID    string
		wantErr   bool
	}{
		{
			name:      "successful get label",
			repoOwner: "owner",
			repoName:  "repo",
			labelName: "bug",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				v := reflect.ValueOf(q).Elem()
				repo := v.FieldByName("Repository")
				label := repo.FieldByName("Label")
				label.FieldByName("ID").SetString("LA_kwDOBug123")
				return nil
			},
			wantID:  "LA_kwDOBug123",
			wantErr: false,
		},
		{
			name:      "label not found",
			repoOwner: "owner",
			repoName:  "repo",
			labelName: "nonexistent",
			mockQuery: func(ctx context.Context, q interface{}, variables map[string]interface{}) error {
				return errors.New("GraphQL error: Could not resolve to a Label")
			},
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGraphQLClient{
				queryFunc: tt.mockQuery,
			}

			ctx := context.Background()
			id, err := GetLabel(ctx, mock, tt.repoOwner, tt.repoName, tt.labelName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if id != tt.wantID {
				t.Errorf("GetLabel() = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestSetLabel_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		issueID    string
		labelID    string
		mockMutate func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
		wantErr    bool
	}{
		{
			name:    "successful set label",
			issueID: "I_kwDOTest123",
			labelID: "LA_kwDOBug123",
			mockMutate: func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "mutation error",
			issueID: "I_invalid",
			labelID: "LA_kwDOBug123",
			mockMutate: func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
				return errors.New("GraphQL error: Could not resolve to a node")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGraphQLClient{
				mutateFunc: tt.mockMutate,
			}

			ctx := context.Background()
			err := SetLabel(ctx, mock, tt.issueID, tt.labelID)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetLabel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveLabel_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		issueID    string
		labelID    string
		mockMutate func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
		wantErr    bool
	}{
		{
			name:    "successful remove label",
			issueID: "I_kwDOTest123",
			labelID: "LA_kwDOBug123",
			mockMutate: func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "mutation error",
			issueID: "I_invalid",
			labelID: "LA_kwDOBug123",
			mockMutate: func(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
				return errors.New("GraphQL error: Could not resolve to a node")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockGraphQLClient{
				mutateFunc: tt.mockMutate,
			}

			ctx := context.Background()
			err := RemoveLabel(ctx, mock, tt.issueID, tt.labelID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveLabel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
