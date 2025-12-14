package authhelper

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func TestGenerateState(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "standard length",
			length:  32,
			wantErr: false,
		},
		{
			name:    "short length",
			length:  8,
			wantErr: false,
		},
		{
			name:    "long length",
			length:  64,
			wantErr: false,
		},
		{
			name:    "zero length",
			length:  0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateState(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify it's valid base64
				decoded, err := base64.URLEncoding.DecodeString(result)
				if err != nil {
					t.Errorf("GenerateState() returned invalid base64: %v", err)
				}
				if len(decoded) != tt.length {
					t.Errorf("GenerateState() decoded length = %v, want %v", len(decoded), tt.length)
				}
			}
		})
	}
}

func TestGenerateState_Uniqueness(t *testing.T) {
	// Generate multiple states and ensure they're unique
	states := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		state, err := GenerateState(32)
		if err != nil {
			t.Fatalf("GenerateState() error = %v", err)
		}
		if states[state] {
			t.Errorf("GenerateState() generated duplicate state on iteration %d", i)
		}
		states[state] = true
	}
}

func TestGeneratePKCEVerifier(t *testing.T) {
	// Test that verifier has correct length
	verifier := GeneratePKCEVerifier()
	if len(verifier) != lengthCodeVerifier {
		t.Errorf("GeneratePKCEVerifier() length = %v, want %v", len(verifier), lengthCodeVerifier)
	}

	// Test that verifier only contains allowed characters
	allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._"
	for _, char := range verifier {
		if !strings.ContainsRune(allowedChars, char) {
			t.Errorf("GeneratePKCEVerifier() contains invalid character: %c", char)
		}
	}
}

func TestGeneratePKCEVerifier_Uniqueness(t *testing.T) {
	verifiers := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		verifier := GeneratePKCEVerifier()
		if verifiers[verifier] {
			t.Errorf("GeneratePKCEVerifier() generated duplicate verifier on iteration %d", i)
		}
		verifiers[verifier] = true
	}
}

func TestGeneratePKCEChallenge(t *testing.T) {
	tests := []struct {
		name     string
		verifier string
	}{
		{
			name:     "standard verifier",
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
		},
		{
			name:     "another verifier",
			verifier: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		},
		{
			name:     "generated verifier",
			verifier: GeneratePKCEVerifier(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge := GeneratePKCEChallenge(tt.verifier)

			// Verify challenge is valid base64 URL encoded (no padding)
			if strings.Contains(challenge, "=") {
				t.Error("GeneratePKCEChallenge() should not contain padding")
			}

			// Verify challenge is not empty
			if challenge == "" {
				t.Error("GeneratePKCEChallenge() returned empty string")
			}

			// Verify the challenge can be decoded
			_, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(challenge)
			if err != nil {
				t.Errorf("GeneratePKCEChallenge() returned invalid base64: %v", err)
			}
		})
	}
}

func TestGeneratePKCEChallenge_Deterministic(t *testing.T) {
	// Same verifier should always produce the same challenge
	verifier := "test-verifier-string"
	challenge1 := GeneratePKCEChallenge(verifier)
	challenge2 := GeneratePKCEChallenge(verifier)

	if challenge1 != challenge2 {
		t.Errorf("GeneratePKCEChallenge() not deterministic: %v != %v", challenge1, challenge2)
	}
}

func TestGeneratePKCEChallenge_DifferentVerifiersDifferentChallenges(t *testing.T) {
	verifier1 := "verifier-one"
	verifier2 := "verifier-two"

	challenge1 := GeneratePKCEChallenge(verifier1)
	challenge2 := GeneratePKCEChallenge(verifier2)

	if challenge1 == challenge2 {
		t.Error("GeneratePKCEChallenge() should produce different challenges for different verifiers")
	}
}

func TestSaveTokenToViper(t *testing.T) {
	// Reset viper for testing
	viper.Reset()

	tests := []struct {
		name    string
		token   *oauth2.Token
		wantErr bool
	}{
		{
			name: "valid token",
			token: &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				TokenType:    "Bearer",
				Expiry:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:    "nil token",
			token:   nil,
			wantErr: true,
		},
		{
			name: "empty token",
			token: &oauth2.Token{
				AccessToken:  "",
				RefreshToken: "",
				TokenType:    "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			err := SaveTokenToViper(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveTokenToViper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.token != nil {
				// Verify values were saved
				if viper.GetString("access_token") != tt.token.AccessToken {
					t.Errorf("access_token = %v, want %v", viper.GetString("access_token"), tt.token.AccessToken)
				}
				if viper.GetString("refresh_token") != tt.token.RefreshToken {
					t.Errorf("refresh_token = %v, want %v", viper.GetString("refresh_token"), tt.token.RefreshToken)
				}
				if viper.GetString("token_type") != tt.token.TokenType {
					t.Errorf("token_type = %v, want %v", viper.GetString("token_type"), tt.token.TokenType)
				}
			}
		})
	}
}

func TestLoadTokenFromViper(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		wantAccess  string
		wantRefresh string
		wantType    string
		wantErr     bool
	}{
		{
			name: "valid token in viper",
			setup: func() {
				viper.Reset()
				viper.Set("access_token", "loaded-access-token")
				viper.Set("refresh_token", "loaded-refresh-token")
				viper.Set("token_type", "Bearer")
				viper.Set("expiry", time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC))
			},
			wantAccess:  "loaded-access-token",
			wantRefresh: "loaded-refresh-token",
			wantType:    "Bearer",
			wantErr:     false,
		},
		{
			name: "empty viper",
			setup: func() {
				viper.Reset()
			},
			wantAccess:  "",
			wantRefresh: "",
			wantType:    "",
			wantErr:     false,
		},
		{
			name: "partial token in viper",
			setup: func() {
				viper.Reset()
				viper.Set("access_token", "partial-access")
			},
			wantAccess:  "partial-access",
			wantRefresh: "",
			wantType:    "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			token, err := LoadTokenFromViper()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTokenFromViper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token.AccessToken != tt.wantAccess {
				t.Errorf("AccessToken = %v, want %v", token.AccessToken, tt.wantAccess)
			}
			if token.RefreshToken != tt.wantRefresh {
				t.Errorf("RefreshToken = %v, want %v", token.RefreshToken, tt.wantRefresh)
			}
			if token.TokenType != tt.wantType {
				t.Errorf("TokenType = %v, want %v", token.TokenType, tt.wantType)
			}
		})
	}
}

func TestSaveAndLoadTokenFromViper_RoundTrip(t *testing.T) {
	viper.Reset()

	originalToken := &oauth2.Token{
		AccessToken:  "roundtrip-access",
		RefreshToken: "roundtrip-refresh",
		TokenType:    "Bearer",
		Expiry:       time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	err := SaveTokenToViper(originalToken)
	if err != nil {
		t.Fatalf("SaveTokenToViper() error = %v", err)
	}

	loadedToken, err := LoadTokenFromViper()
	if err != nil {
		t.Fatalf("LoadTokenFromViper() error = %v", err)
	}

	if loadedToken.AccessToken != originalToken.AccessToken {
		t.Errorf("AccessToken = %v, want %v", loadedToken.AccessToken, originalToken.AccessToken)
	}
	if loadedToken.RefreshToken != originalToken.RefreshToken {
		t.Errorf("RefreshToken = %v, want %v", loadedToken.RefreshToken, originalToken.RefreshToken)
	}
	if loadedToken.TokenType != originalToken.TokenType {
		t.Errorf("TokenType = %v, want %v", loadedToken.TokenType, originalToken.TokenType)
	}
}

func TestGetServer(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantAddr string
	}{
		{
			name:     "standard port",
			port:     8080,
			wantAddr: ":8080",
		},
		{
			name:     "port 3000",
			port:     3000,
			wantAddr: ":3000",
		},
		{
			name:     "high port",
			port:     65535,
			wantAddr: ":65535",
		},
		{
			name:     "low port",
			port:     80,
			wantAddr: ":80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := getServer(tt.port)
			if server.Addr != tt.wantAddr {
				t.Errorf("getServer() Addr = %v, want %v", server.Addr, tt.wantAddr)
			}
			if server.ReadHeaderTimeout != 10*time.Second {
				t.Errorf("getServer() ReadHeaderTimeout = %v, want %v", server.ReadHeaderTimeout, 10*time.Second)
			}
		})
	}
}

func TestGetToken_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *OAuthConfig
		wantErr string
	}{
		{
			name: "missing client_id",
			config: &OAuthConfig{
				ClientId:     "",
				ClientSecret: "secret",
				RedirectURI:  "/callback",
			},
			wantErr: "client_id is not configured",
		},
		{
			name: "missing client_secret",
			config: &OAuthConfig{
				ClientId:     "client-id",
				ClientSecret: "",
				RedirectURI:  "/callback",
			},
			wantErr: "client_secret is not configured",
		},
		{
			name: "missing redirect_uri",
			config: &OAuthConfig{
				ClientId:     "client-id",
				ClientSecret: "secret",
				RedirectURI:  "",
			},
			wantErr: "redirect_uri is not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := GetToken(ctx, tt.config, false)
			if err == nil {
				t.Errorf("GetToken() expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("GetToken() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRefreshToken_NilToken(t *testing.T) {
	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}

	// Test with expired/invalid token - should fail when trying to refresh
	expiredToken := &oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "invalid-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	_, err := RefreshToken(ctx, config, expiredToken)
	// This should return an error because we can't actually refresh against a fake endpoint
	if err == nil {
		t.Error("RefreshToken() expected error for invalid refresh, got nil")
	}
}

// Benchmark tests
func BenchmarkGenerateState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateState(32)
	}
}

func BenchmarkGeneratePKCEVerifier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GeneratePKCEVerifier()
	}
}

func BenchmarkGeneratePKCEChallenge(b *testing.B) {
	verifier := GeneratePKCEVerifier()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GeneratePKCEChallenge(verifier)
	}
}

func BenchmarkSaveTokenToViper(b *testing.B) {
	token := &oauth2.Token{
		AccessToken:  "benchmark-access-token",
		RefreshToken: "benchmark-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SaveTokenToViper(token)
	}
}

func BenchmarkLoadTokenFromViper(b *testing.B) {
	viper.Set("access_token", "benchmark-access")
	viper.Set("refresh_token", "benchmark-refresh")
	viper.Set("token_type", "Bearer")
	viper.Set("expiry", time.Now())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LoadTokenFromViper()
	}
}
