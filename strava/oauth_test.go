package strava

import (
	"testing"
)

func TestGetOAuthEndpoint(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	expectedAuthURL := "https://www.strava.com/oauth/authorize"
	expectedTokenURL := "https://www.strava.com/oauth/token"

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

func TestGetOAuthEndpointContainsStravaHost(t *testing.T) {
	endpoint := GetOAuthEndpoint()

	stravaHost := "www.strava.com"

	if len(endpoint.AuthURL) < len(stravaHost) || !containsSubstring(endpoint.AuthURL, stravaHost) {
		t.Errorf("GetOAuthEndpoint().AuthURL should contain %q, got %q", stravaHost, endpoint.AuthURL)
	}

	if len(endpoint.TokenURL) < len(stravaHost) || !containsSubstring(endpoint.TokenURL, stravaHost) {
		t.Errorf("GetOAuthEndpoint().TokenURL should contain %q, got %q", stravaHost, endpoint.TokenURL)
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
