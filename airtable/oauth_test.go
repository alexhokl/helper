package airtable

import (
	"testing"
)

func TestGetOAuthEndpoint(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	expectedAuthURL := "https://airtable.com/oauth2/v1/authorize"
	expectedTokenURL := "https://airtable.com/oauth2/v1/token"

	if endpoint.AuthURL != expectedAuthURL {
		t.Errorf("GetOAuthEndpoint().AuthURL = %q, want %q", endpoint.AuthURL, expectedAuthURL)
	}

	if endpoint.TokenURL != expectedTokenURL {
		t.Errorf("GetOAuthEndpoint().TokenURL = %q, want %q", endpoint.TokenURL, expectedTokenURL)
	}
}

func TestGetOAuthEndpointAuthURLNotEmpty(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	if endpoint.AuthURL == "" {
		t.Error("GetOAuthEndpoint().AuthURL should not be empty")
	}
}

func TestGetOAuthEndpointTokenURLNotEmpty(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	if endpoint.TokenURL == "" {
		t.Error("GetOAuthEndpoint().TokenURL should not be empty")
	}
}

func TestGetOAuthEndpointURLsAreHTTPS(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	if len(endpoint.AuthURL) < 8 || endpoint.AuthURL[:8] != "https://" {
		t.Errorf("GetOAuthEndpoint().AuthURL should use HTTPS, got %q", endpoint.AuthURL)
	}

	if len(endpoint.TokenURL) < 8 || endpoint.TokenURL[:8] != "https://" {
		t.Errorf("GetOAuthEndpoint().TokenURL should use HTTPS, got %q", endpoint.TokenURL)
	}
}

func TestGetOAuthEndpointConsistency(t *testing.T) {
	// Call GetOAuthEndpoint multiple times and verify consistency
	endpoint1 := GetOAuthEndpoint()
	endpoint2 := GetOAuthEndpoint()

	if endpoint1.AuthURL != endpoint2.AuthURL {
		t.Errorf("GetOAuthEndpoint() returned inconsistent AuthURL: %q vs %q", endpoint1.AuthURL, endpoint2.AuthURL)
	}

	if endpoint1.TokenURL != endpoint2.TokenURL {
		t.Errorf("GetOAuthEndpoint() returned inconsistent TokenURL: %q vs %q", endpoint1.TokenURL, endpoint2.TokenURL)
	}
}

func TestGetOAuthEndpointContainsAirtableHost(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	airtableHost := "airtable.com"

	if !containsSubstring(endpoint.AuthURL, airtableHost) {
		t.Errorf("GetOAuthEndpoint().AuthURL should contain %q, got %q", airtableHost, endpoint.AuthURL)
	}

	if !containsSubstring(endpoint.TokenURL, airtableHost) {
		t.Errorf("GetOAuthEndpoint().TokenURL should contain %q, got %q", airtableHost, endpoint.TokenURL)
	}
}

func TestGetOAuthEndpointContainsOAuth2Path(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	oauth2Path := "oauth2/v1"

	if !containsSubstring(endpoint.AuthURL, oauth2Path) {
		t.Errorf("GetOAuthEndpoint().AuthURL should contain %q, got %q", oauth2Path, endpoint.AuthURL)
	}

	if !containsSubstring(endpoint.TokenURL, oauth2Path) {
		t.Errorf("GetOAuthEndpoint().TokenURL should contain %q, got %q", oauth2Path, endpoint.TokenURL)
	}
}

// Test apiURL constant
func TestAPIURLConstant(t *testing.T) {
	if apiURL != "https://airtable.com" {
		t.Errorf("apiURL = %q, want %q", apiURL, "https://airtable.com")
	}
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark test
func BenchmarkGetOAuthEndpoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetOAuthEndpoint()
	}
}
