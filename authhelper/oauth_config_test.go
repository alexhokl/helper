package authhelper

import (
	"testing"

	"golang.org/x/oauth2"
)

func TestOAuthConfig_GetOAuthConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   OAuthConfig
		expected *oauth2.Config
	}{
		{
			name: "basic config",
			config: OAuthConfig{
				ClientId:     "test-client-id",
				ClientSecret: "test-client-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
				Scopes:      []string{"read", "write"},
				RedirectURI: "/callback",
				Port:        8080,
			},
			expected: &oauth2.Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
				Scopes:      []string{"read", "write"},
				RedirectURL: "http://localhost:8080/callback",
			},
		},
		{
			name: "different port",
			config: OAuthConfig{
				ClientId:     "another-client",
				ClientSecret: "another-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://auth.example.com/authorize",
					TokenURL: "https://auth.example.com/token",
				},
				Scopes:      []string{"profile"},
				RedirectURI: "/oauth/callback",
				Port:        3000,
			},
			expected: &oauth2.Config{
				ClientID:     "another-client",
				ClientSecret: "another-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://auth.example.com/authorize",
					TokenURL: "https://auth.example.com/token",
				},
				Scopes:      []string{"profile"},
				RedirectURL: "http://localhost:3000/oauth/callback",
			},
		},
		{
			name: "empty scopes",
			config: OAuthConfig{
				ClientId:     "client",
				ClientSecret: "secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
				Scopes:      []string{},
				RedirectURI: "/cb",
				Port:        9000,
			},
			expected: &oauth2.Config{
				ClientID:     "client",
				ClientSecret: "secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
				Scopes:      []string{},
				RedirectURL: "http://localhost:9000/cb",
			},
		},
		{
			name: "multiple scopes",
			config: OAuthConfig{
				ClientId:     "multi-scope-client",
				ClientSecret: "multi-scope-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
				Scopes:      []string{"openid", "profile", "email", "offline_access"},
				RedirectURI: "/auth/callback",
				Port:        8000,
			},
			expected: &oauth2.Config{
				ClientID:     "multi-scope-client",
				ClientSecret: "multi-scope-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
				Scopes:      []string{"openid", "profile", "email", "offline_access"},
				RedirectURL: "http://localhost:8000/auth/callback",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetOAuthConfig()

			if result.ClientID != tt.expected.ClientID {
				t.Errorf("ClientID = %v, want %v", result.ClientID, tt.expected.ClientID)
			}
			if result.ClientSecret != tt.expected.ClientSecret {
				t.Errorf("ClientSecret = %v, want %v", result.ClientSecret, tt.expected.ClientSecret)
			}
			if result.RedirectURL != tt.expected.RedirectURL {
				t.Errorf("RedirectURL = %v, want %v", result.RedirectURL, tt.expected.RedirectURL)
			}
			if result.Endpoint.AuthURL != tt.expected.Endpoint.AuthURL {
				t.Errorf("Endpoint.AuthURL = %v, want %v", result.Endpoint.AuthURL, tt.expected.Endpoint.AuthURL)
			}
			if result.Endpoint.TokenURL != tt.expected.Endpoint.TokenURL {
				t.Errorf("Endpoint.TokenURL = %v, want %v", result.Endpoint.TokenURL, tt.expected.Endpoint.TokenURL)
			}
			if len(result.Scopes) != len(tt.expected.Scopes) {
				t.Errorf("len(Scopes) = %v, want %v", len(result.Scopes), len(tt.expected.Scopes))
			}
			for i, scope := range result.Scopes {
				if scope != tt.expected.Scopes[i] {
					t.Errorf("Scopes[%d] = %v, want %v", i, scope, tt.expected.Scopes[i])
				}
			}
		})
	}
}

func TestOAuthConfig_GetOAuthConfig_RedirectURLFormat(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		redirectURI string
		wantURL     string
	}{
		{
			name:        "standard port and path",
			port:        8080,
			redirectURI: "/callback",
			wantURL:     "http://localhost:8080/callback",
		},
		{
			name:        "port 80",
			port:        80,
			redirectURI: "/oauth",
			wantURL:     "http://localhost:80/oauth",
		},
		{
			name:        "high port number",
			port:        65535,
			redirectURI: "/auth",
			wantURL:     "http://localhost:65535/auth",
		},
		{
			name:        "nested path",
			port:        8080,
			redirectURI: "/api/v1/oauth/callback",
			wantURL:     "http://localhost:8080/api/v1/oauth/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OAuthConfig{
				ClientId:     "test",
				ClientSecret: "test",
				Port:         tt.port,
				RedirectURI:  tt.redirectURI,
			}
			result := config.GetOAuthConfig()
			if result.RedirectURL != tt.wantURL {
				t.Errorf("RedirectURL = %v, want %v", result.RedirectURL, tt.wantURL)
			}
		})
	}
}
