package googleapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func TestNewFromClientIDSecret(t *testing.T) {
	ctx := context.Background()
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	scopes := []string{"scope1", "scope2"}

	client, err := NewFromClientIDSecret(ctx, clientID, clientSecret, scopes, nil)
	if err != nil {
		t.Fatalf("NewFromClientIDSecret() error: %v", err)
	}

	if client == nil {
		t.Fatal("NewFromClientIDSecret() returned nil client")
	}

	if client.Config.ClientID != clientID {
		t.Errorf("ClientID = %q, want %q", client.Config.ClientID, clientID)
	}

	if client.Config.ClientSecret != clientSecret {
		t.Errorf("ClientSecret = %q, want %q", client.Config.ClientSecret, clientSecret)
	}

	if len(client.Config.Scopes) != len(scopes) {
		t.Errorf("Scopes length = %d, want %d", len(client.Config.Scopes), len(scopes))
	}

	for i, scope := range scopes {
		if client.Config.Scopes[i] != scope {
			t.Errorf("Scopes[%d] = %q, want %q", i, client.Config.Scopes[i], scope)
		}
	}
}

func TestNewFromClientIDSecretWithToken(t *testing.T) {
	ctx := context.Background()
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	scopes := []string{"scope1"}
	token := &oauth2.Token{
		AccessToken: "test-access-token",
	}

	client, err := NewFromClientIDSecret(ctx, clientID, clientSecret, scopes, token)
	if err != nil {
		t.Fatalf("NewFromClientIDSecret() error: %v", err)
	}

	if client.Token == nil {
		t.Fatal("Token should not be nil")
	}

	if client.Token.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %q, want %q", client.Token.AccessToken, token.AccessToken)
	}
}

func TestNewFromClientIDSecretEmptyScopes(t *testing.T) {
	ctx := context.Background()
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	scopes := []string{}

	client, err := NewFromClientIDSecret(ctx, clientID, clientSecret, scopes, nil)
	if err != nil {
		t.Fatalf("NewFromClientIDSecret() error: %v", err)
	}

	if len(client.Config.Scopes) != 0 {
		t.Errorf("Scopes length = %d, want 0", len(client.Config.Scopes))
	}
}

func TestNewFromClientIDSecretRedirectURL(t *testing.T) {
	ctx := context.Background()
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	scopes := []string{"scope1"}

	client, err := NewFromClientIDSecret(ctx, clientID, clientSecret, scopes, nil)
	if err != nil {
		t.Fatalf("NewFromClientIDSecret() error: %v", err)
	}

	expectedRedirectURL := "http://localhost:9998/callback"
	if client.Config.RedirectURL != expectedRedirectURL {
		t.Errorf("RedirectURL = %q, want %q", client.Config.RedirectURL, expectedRedirectURL)
	}
}

func TestNewFromClientIDSecretEndpoint(t *testing.T) {
	ctx := context.Background()
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	scopes := []string{"scope1"}

	client, err := NewFromClientIDSecret(ctx, clientID, clientSecret, scopes, nil)
	if err != nil {
		t.Fatalf("NewFromClientIDSecret() error: %v", err)
	}

	// Verify Google endpoint is set
	if client.Config.Endpoint.AuthURL == "" {
		t.Error("Endpoint.AuthURL should not be empty")
	}
	if client.Config.Endpoint.TokenURL == "" {
		t.Error("Endpoint.TokenURL should not be empty")
	}
}

func TestNewNonExistentFile(t *testing.T) {
	ctx := context.Background()
	nonExistentPath := "/nonexistent/path/to/client_secret.json"
	scopes := []string{"scope1"}

	_, err := New(ctx, nonExistentPath, nil, scopes)
	if err == nil {
		t.Error("New() with non-existent file should return error")
	}
}

func TestNewInvalidJSONFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	invalidJSONPath := filepath.Join(tmpDir, "invalid.json")

	// Create file with invalid JSON
	if err := os.WriteFile(invalidJSONPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	scopes := []string{"scope1"}
	_, err := New(ctx, invalidJSONPath, nil, scopes)
	if err == nil {
		t.Error("New() with invalid JSON should return error")
	}
}

func TestNewEmptyFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	emptyFilePath := filepath.Join(tmpDir, "empty.json")

	// Create empty file
	if err := os.WriteFile(emptyFilePath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	scopes := []string{"scope1"}
	_, err := New(ctx, emptyFilePath, nil, scopes)
	if err == nil {
		t.Error("New() with empty file should return error")
	}
}

func TestGoogleClientNewHttpClient(t *testing.T) {
	ctx := context.Background()
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	scopes := []string{"scope1"}
	token := &oauth2.Token{
		AccessToken: "test-access-token",
	}

	client, err := NewFromClientIDSecret(ctx, clientID, clientSecret, scopes, token)
	if err != nil {
		t.Fatalf("NewFromClientIDSecret() error: %v", err)
	}

	httpClient := client.NewHttpClient()
	if httpClient == nil {
		t.Error("NewHttpClient() returned nil")
	}
}

func TestGoogleClientStruct(t *testing.T) {
	client := &GoogleClient{
		Context: context.Background(),
		Config: &oauth2.Config{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
		},
		Token: &oauth2.Token{
			AccessToken: "test-token",
		},
	}

	if client.Context == nil {
		t.Error("Context should not be nil")
	}
	if client.Config == nil {
		t.Error("Config should not be nil")
	}
	if client.Token == nil {
		t.Error("Token should not be nil")
	}
}

func TestNewMapClientEmptyAPIKey(t *testing.T) {
	// Empty API key returns an error from the maps library
	_, err := NewMapClient("")
	if err == nil {
		t.Error("NewMapClient(\"\") should return error for empty API key")
	}
}

func TestNewMapClientValidAPIKey(t *testing.T) {
	client, err := NewMapClient("test-api-key")
	if err != nil {
		t.Fatalf("NewMapClient() error: %v", err)
	}
	if client == nil {
		t.Error("NewMapClient() returned nil client")
	}
}

// Test constants
func TestConstants(t *testing.T) {
	if port != 9998 {
		t.Errorf("port = %d, want 9998", port)
	}
	if callbackUri != "/callback" {
		t.Errorf("callbackUri = %q, want %q", callbackUri, "/callback")
	}
}
