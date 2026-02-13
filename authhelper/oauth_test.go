package authhelper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// mockTokenResponse represents a mock OAuth token response
type mockTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// newMockOAuthServer creates a mock OAuth server for testing
func newMockOAuthServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			response := mockTokenResponse{
				AccessToken:  "mock-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "mock-refresh-token",
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode mock response: %v", err)
			}
		}
	}))
}

// newMockOAuthServerWithValidation creates a mock OAuth server that validates requests
func newMockOAuthServerWithValidation(t *testing.T, expectedCode string, expectedVerifier string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			code := r.FormValue("code")
			if code != expectedCode {
				http.Error(w, `{"error":"invalid_grant","error_description":"Invalid code"}`, http.StatusBadRequest)
				return
			}

			// If PKCE is expected, validate the verifier
			if expectedVerifier != "" {
				verifier := r.FormValue("code_verifier")
				if verifier != expectedVerifier {
					http.Error(w, `{"error":"invalid_grant","error_description":"Invalid code verifier"}`, http.StatusBadRequest)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			response := mockTokenResponse{
				AccessToken:  "validated-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "validated-refresh-token",
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode mock response: %v", err)
			}
		}
	}))
}

// newMockOAuthServerWithError creates a mock OAuth server that returns errors
func newMockOAuthServerWithError(t *testing.T, errorCode string, errorDescription string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]string{
				"error":             errorCode,
				"error_description": errorDescription,
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode mock error response: %v", err)
			}
		}
	}))
}

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

func TestGetTokenHandler_StateMismatch(t *testing.T) {
	expectedState := "expected-state-value"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, "")

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=wrong-state&code=test-code", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		if !strings.Contains(err.Error(), "state mismatch") {
			t.Errorf("Expected state mismatch error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestGetTokenHandler_CodeChallengeMismatch(t *testing.T) {
	expectedState := "test-state"
	expectedChallenge := "expected-challenge"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, expectedChallenge)
	ctx = context.WithValue(ctx, codeVerifierContextKey, "test-verifier")

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=test-code&code_challenge=wrong-challenge", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		if !strings.Contains(err.Error(), "code_challenge mismatch") {
			t.Errorf("Expected code_challenge mismatch error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestGetTokenHandler_CodeChallengeMethodMismatch(t *testing.T) {
	expectedState := "test-state"
	expectedChallenge := "expected-challenge"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, expectedChallenge)
	ctx = context.WithValue(ctx, codeVerifierContextKey, "test-verifier")

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=test-code&code_challenge=expected-challenge&code_challenge_method=plain", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		if !strings.Contains(err.Error(), "code_challenge_method mismatch") {
			t.Errorf("Expected code_challenge_method mismatch error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestGetTokenHandler_OAuthError(t *testing.T) {
	expectedState := "test-state"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, "")

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&error=access_denied&error_description=User%20denied%20access", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		if !strings.Contains(err.Error(), "access_denied") {
			t.Errorf("Expected access_denied error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "User denied access") {
			t.Errorf("Expected error description in error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestGetTokenHandler_OAuthErrorWithoutDescription(t *testing.T) {
	expectedState := "test-state"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, "")

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&error=server_error", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		if !strings.Contains(err.Error(), "server_error") {
			t.Errorf("Expected server_error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestGetTokenHandler_ExchangeError(t *testing.T) {
	expectedState := "test-state"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, "")

	// Use invalid token URL to trigger exchange error
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "http://invalid-url-that-will-fail.local/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=test-code", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		if !strings.Contains(err.Error(), "failed to exchange token") {
			t.Errorf("Expected exchange error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(5 * time.Second):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestGetTokenHandler_ExchangeWithPKCE(t *testing.T) {
	expectedState := "test-state"
	expectedChallenge := "test-challenge"
	expectedVerifier := "test-verifier"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, expectedChallenge)
	ctx = context.WithValue(ctx, codeVerifierContextKey, expectedVerifier)

	// Use invalid token URL to trigger exchange error (but after PKCE validation passes)
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "http://invalid-url-that-will-fail.local/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=test-code&code_challenge=test-challenge&code_challenge_method=S256", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		// We expect exchange error since the URL is invalid, but PKCE validation should pass
		if !strings.Contains(err.Error(), "failed to exchange token") {
			t.Errorf("Expected exchange error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(5 * time.Second):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestGetTokenHandler_SuccessfulExchange(t *testing.T) {
	mockServer := newMockOAuthServer(t)
	defer mockServer.Close()

	expectedState := "test-state"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, "")

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=valid-code", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case token := <-tokenChannel:
		if token.AccessToken != "mock-access-token" {
			t.Errorf("Expected access token 'mock-access-token', got: %v", token.AccessToken)
		}
		if token.TokenType != "Bearer" {
			t.Errorf("Expected token type 'Bearer', got: %v", token.TokenType)
		}
		if token.RefreshToken != "mock-refresh-token" {
			t.Errorf("Expected refresh token 'mock-refresh-token', got: %v", token.RefreshToken)
		}
	case err := <-errorChannel:
		t.Errorf("Expected token, got error: %v", err)
	case <-time.After(5 * time.Second):
		t.Error("Expected token channel to receive, timed out")
	}

	// Verify response was written
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got: %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "authenticated") {
		t.Errorf("Expected success message in response body, got: %v", w.Body.String())
	}
}

func TestGetTokenHandler_SuccessfulExchangeWithPKCE(t *testing.T) {
	expectedCode := "pkce-valid-code"
	expectedVerifier := "pkce-test-verifier-12345"
	mockServer := newMockOAuthServerWithValidation(t, expectedCode, expectedVerifier)
	defer mockServer.Close()

	expectedState := "test-state"
	expectedChallenge := GeneratePKCEChallenge(expectedVerifier)
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, expectedChallenge)
	ctx = context.WithValue(ctx, codeVerifierContextKey, expectedVerifier)

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code="+expectedCode+"&code_challenge="+expectedChallenge+"&code_challenge_method=S256", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case token := <-tokenChannel:
		if token.AccessToken != "validated-access-token" {
			t.Errorf("Expected access token 'validated-access-token', got: %v", token.AccessToken)
		}
		if token.RefreshToken != "validated-refresh-token" {
			t.Errorf("Expected refresh token 'validated-refresh-token', got: %v", token.RefreshToken)
		}
	case err := <-errorChannel:
		t.Errorf("Expected token, got error: %v", err)
	case <-time.After(5 * time.Second):
		t.Error("Expected token channel to receive, timed out")
	}
}

func TestGetTokenHandler_ServerReturnsError(t *testing.T) {
	mockServer := newMockOAuthServerWithError(t, "invalid_client", "Client authentication failed")
	defer mockServer.Close()

	expectedState := "test-state"
	ctx := context.WithValue(context.Background(), stateContextKey, expectedState)
	ctx = context.WithValue(ctx, codeChallengeContextKey, "")

	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "wrong-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	tokenChannel := make(chan oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	handler := getTokenHandler(ctx, config, tokenChannel, errorChannel)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=some-code", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errorChannel:
		if !strings.Contains(err.Error(), "failed to exchange token") {
			t.Errorf("Expected exchange error, got: %v", err)
		}
	case <-tokenChannel:
		t.Error("Expected error, got token")
	case <-time.After(5 * time.Second):
		t.Error("Expected error channel to receive, timed out")
	}
}

func TestRefreshToken_WithMockServer(t *testing.T) {
	mockServer := newMockOAuthServer(t)
	defer mockServer.Close()

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	// Create an expired token with a valid refresh token
	expiredToken := &oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // Expired
	}

	newToken, err := RefreshToken(ctx, config, expiredToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if newToken.AccessToken != "mock-access-token" {
		t.Errorf("Expected new access token 'mock-access-token', got: %v", newToken.AccessToken)
	}
	if newToken.RefreshToken != "mock-refresh-token" {
		t.Errorf("Expected new refresh token 'mock-refresh-token', got: %v", newToken.RefreshToken)
	}
}

func TestRefreshToken_ServerError(t *testing.T) {
	mockServer := newMockOAuthServerWithError(t, "invalid_grant", "Refresh token expired")
	defer mockServer.Close()

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	expiredToken := &oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "expired-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	_, err := RefreshToken(ctx, config, expiredToken)
	if err == nil {
		t.Error("RefreshToken() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to retrieve a new token") {
		t.Errorf("Expected refresh error message, got: %v", err)
	}
}

func TestRefreshToken_ValidToken(t *testing.T) {
	mockServer := newMockOAuthServer(t)
	defer mockServer.Close()

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	// Create a valid (non-expired) token
	validToken := &oauth2.Token{
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour), // Not expired
	}

	// TokenSource should return the same token if it's still valid
	newToken, err := RefreshToken(ctx, config, validToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	// The token should be returned (either same or refreshed)
	if newToken.AccessToken == "" {
		t.Error("RefreshToken() returned empty access token")
	}
}

// Tests for TokenOptions and functional options

func TestWithBrowserOpener(t *testing.T) {
	customOpener := func(url string) error {
		return errors.New("custom opener called")
	}

	opts := defaultTokenOptions()
	WithBrowserOpener(customOpener)(opts)

	err := opts.browserOpener("http://example.com")
	if err == nil || err.Error() != "custom opener called" {
		t.Errorf("WithBrowserOpener did not set custom opener, got error: %v", err)
	}
}

func TestWithOutput(t *testing.T) {
	var buf bytes.Buffer
	customOutput := func(w io.Writer, format string, args ...interface{}) (int, error) {
		return fmt.Fprintf(w, "CUSTOM: "+format, args...)
	}

	opts := defaultTokenOptions()
	WithOutput(customOutput)(opts)
	WithOutputWriter(&buf)(opts)

	_, _ = opts.output(opts.outputWriter, "test %s", "message")

	if !strings.Contains(buf.String(), "CUSTOM: test message") {
		t.Errorf("WithOutput did not set custom output, got: %s", buf.String())
	}
}

func TestWithOutputWriter(t *testing.T) {
	var buf bytes.Buffer

	opts := defaultTokenOptions()
	WithOutputWriter(&buf)(opts)

	if opts.outputWriter != &buf {
		t.Error("WithOutputWriter did not set custom writer")
	}
}

func TestWithSleepDuration(t *testing.T) {
	opts := defaultTokenOptions()
	WithSleepDuration(500 * time.Millisecond)(opts)

	if opts.sleepDuration != 500*time.Millisecond {
		t.Errorf("WithSleepDuration did not set duration, got: %v", opts.sleepDuration)
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	opts := defaultTokenOptions()
	WithShutdownTimeout(10 * time.Second)(opts)

	if opts.shutdownTimeout != 10*time.Second {
		t.Errorf("WithShutdownTimeout did not set timeout, got: %v", opts.shutdownTimeout)
	}
}

func TestDefaultTokenOptions(t *testing.T) {
	opts := defaultTokenOptions()

	if opts.browserOpener == nil {
		t.Error("defaultTokenOptions browserOpener is nil")
	}
	if opts.output == nil {
		t.Error("defaultTokenOptions output is nil")
	}
	if opts.outputWriter == nil {
		t.Error("defaultTokenOptions outputWriter is nil")
	}
	if opts.sleepDuration != 1*time.Second {
		t.Errorf("defaultTokenOptions sleepDuration = %v, want 1s", opts.sleepDuration)
	}
	if opts.shutdownTimeout != 5*time.Second {
		t.Errorf("defaultTokenOptions shutdownTimeout = %v, want 5s", opts.shutdownTimeout)
	}
}

// Tests for RefreshTokenOptions

func TestWithTokenSourceFactory(t *testing.T) {
	factoryCalled := false
	customFactory := func(ctx context.Context, config *oauth2.Config, token *oauth2.Token) oauth2.TokenSource {
		factoryCalled = true
		return config.TokenSource(ctx, token)
	}

	opts := defaultRefreshTokenOptions()
	WithTokenSourceFactory(customFactory)(opts)

	// Call the factory to verify it was set
	ctx := context.Background()
	config := &oauth2.Config{}
	token := &oauth2.Token{}
	_ = opts.tokenSourceFactory(ctx, config, token)

	if !factoryCalled {
		t.Error("WithTokenSourceFactory did not set custom factory")
	}
}

func TestDefaultRefreshTokenOptions(t *testing.T) {
	opts := defaultRefreshTokenOptions()

	if opts.tokenSourceFactory == nil {
		t.Error("defaultRefreshTokenOptions tokenSourceFactory is nil")
	}
}

// Integration tests for GetToken with mocked dependencies

func TestGetToken_FullFlow(t *testing.T) {
	// Create a mock OAuth server
	mockServer := newMockOAuthServer(t)
	defer mockServer.Close()

	// Track if browser was "opened"
	browserOpenedURL := ""
	mockBrowserOpener := func(url string) error {
		browserOpenedURL = url
		// Simulate the OAuth callback by making a request to our local server
		// We need to extract the state from the URL and call the callback
		go func() {
			// Give the server time to start
			time.Sleep(50 * time.Millisecond)
			// Parse the auth URL to get the state
			// The callback will be made to our local server
			resp, err := http.Get("http://localhost:19876/callback?state=" + extractState(url) + "&code=test-auth-code")
			if err != nil {
				t.Logf("Callback request failed: %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()
		}()
		return nil
	}

	// Capture output
	var outputBuf bytes.Buffer

	config := &OAuthConfig{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
		RedirectURI: "/callback",
		Port:        19876,
		Scopes:      []string{"read", "write"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := GetToken(ctx, config, false,
		WithBrowserOpener(mockBrowserOpener),
		WithOutput(fmt.Fprintf),
		WithOutputWriter(&outputBuf),
		WithSleepDuration(0),
		WithShutdownTimeout(1*time.Second),
	)

	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}

	if token == nil {
		t.Fatal("GetToken() returned nil token")
	}

	if token.AccessToken != "mock-access-token" {
		t.Errorf("GetToken() AccessToken = %v, want mock-access-token", token.AccessToken)
	}

	if browserOpenedURL == "" {
		t.Error("Browser was not opened")
	}

	if !strings.Contains(outputBuf.String(), "browser for authentication") {
		t.Errorf("Expected authentication message in output, got: %s", outputBuf.String())
	}
}

func TestGetToken_FullFlowWithPKCE(t *testing.T) {
	// Create a mock OAuth server
	mockServer := newMockOAuthServer(t)
	defer mockServer.Close()

	mockBrowserOpener := func(url string) error {
		go func() {
			time.Sleep(50 * time.Millisecond)
			// Extract state and code_challenge from URL for PKCE flow
			state := extractState(url)
			codeChallenge := extractParam(url, "code_challenge")
			callbackURL := fmt.Sprintf("http://localhost:19877/callback?state=%s&code=test-auth-code&code_challenge=%s&code_challenge_method=S256", state, codeChallenge)
			resp, err := http.Get(callbackURL)
			if err != nil {
				t.Logf("Callback request failed: %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()
		}()
		return nil
	}

	var outputBuf bytes.Buffer

	config := &OAuthConfig{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
		RedirectURI: "/callback",
		Port:        19877,
		Scopes:      []string{"read"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := GetToken(ctx, config, true,
		WithBrowserOpener(mockBrowserOpener),
		WithOutput(fmt.Fprintf),
		WithOutputWriter(&outputBuf),
		WithSleepDuration(0),
		WithShutdownTimeout(1*time.Second),
	)

	if err != nil {
		t.Fatalf("GetToken() with PKCE error = %v", err)
	}

	if token == nil {
		t.Fatal("GetToken() with PKCE returned nil token")
	}

	if token.AccessToken != "mock-access-token" {
		t.Errorf("GetToken() with PKCE AccessToken = %v, want mock-access-token", token.AccessToken)
	}
}

func TestGetToken_BrowserOpenError(t *testing.T) {
	mockBrowserOpener := func(url string) error {
		return errors.New("failed to open browser")
	}

	var outputBuf bytes.Buffer

	config := &OAuthConfig{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
		RedirectURI: "/callback",
		Port:        19878,
		Scopes:      []string{"read"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := GetToken(ctx, config, false,
		WithBrowserOpener(mockBrowserOpener),
		WithOutput(fmt.Fprintf),
		WithOutputWriter(&outputBuf),
		WithSleepDuration(0),
		WithShutdownTimeout(1*time.Second),
	)

	if err == nil {
		t.Fatal("GetToken() expected error when browser fails to open")
	}

	if !strings.Contains(err.Error(), "failed to open browser") {
		t.Errorf("GetToken() error = %v, want error containing 'failed to open browser'", err)
	}
}

func TestGetToken_ContextCancellation(t *testing.T) {
	mockBrowserOpener := func(url string) error {
		// Don't simulate callback - let context timeout
		return nil
	}

	var outputBuf bytes.Buffer

	config := &OAuthConfig{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
		RedirectURI: "/callback",
		Port:        19879,
		Scopes:      []string{"read"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := GetToken(ctx, config, false,
		WithBrowserOpener(mockBrowserOpener),
		WithOutput(fmt.Fprintf),
		WithOutputWriter(&outputBuf),
		WithSleepDuration(0),
		WithShutdownTimeout(100*time.Millisecond),
	)

	if err == nil {
		t.Fatal("GetToken() expected error on context cancellation")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("GetToken() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestGetToken_OAuthServerError(t *testing.T) {
	// Create a mock OAuth server that returns errors
	mockServer := newMockOAuthServerWithError(t, "access_denied", "User denied access")
	defer mockServer.Close()

	mockBrowserOpener := func(url string) error {
		go func() {
			time.Sleep(50 * time.Millisecond)
			state := extractState(url)
			// Simulate OAuth error response
			callbackURL := fmt.Sprintf("http://localhost:19880/callback?state=%s&error=access_denied&error_description=User%%20denied%%20access", state)
			resp, err := http.Get(callbackURL)
			if err != nil {
				t.Logf("Callback request failed: %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()
		}()
		return nil
	}

	var outputBuf bytes.Buffer

	config := &OAuthConfig{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
		RedirectURI: "/callback",
		Port:        19880,
		Scopes:      []string{"read"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := GetToken(ctx, config, false,
		WithBrowserOpener(mockBrowserOpener),
		WithOutput(fmt.Fprintf),
		WithOutputWriter(&outputBuf),
		WithSleepDuration(0),
		WithShutdownTimeout(1*time.Second),
	)

	if err == nil {
		t.Fatal("GetToken() expected error on OAuth server error")
	}

	if !strings.Contains(err.Error(), "access_denied") {
		t.Errorf("GetToken() error = %v, want error containing 'access_denied'", err)
	}
}

func TestRefreshToken_WithCustomTokenSource(t *testing.T) {
	expectedToken := &oauth2.Token{
		AccessToken:  "custom-refreshed-token",
		RefreshToken: "custom-new-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	customFactory := func(ctx context.Context, config *oauth2.Config, token *oauth2.Token) oauth2.TokenSource {
		return oauth2.StaticTokenSource(expectedToken)
	}

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}
	oldToken := &oauth2.Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	newToken, err := RefreshToken(ctx, config, oldToken,
		WithTokenSourceFactory(customFactory),
	)

	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if newToken.AccessToken != "custom-refreshed-token" {
		t.Errorf("RefreshToken() AccessToken = %v, want custom-refreshed-token", newToken.AccessToken)
	}
}

// Helper functions for extracting URL parameters

func extractState(url string) string {
	return extractParam(url, "state")
}

func extractParam(url string, param string) string {
	// Simple extraction - find param= and extract value until & or end
	search := param + "="
	idx := strings.Index(url, search)
	if idx == -1 {
		return ""
	}
	start := idx + len(search)
	end := strings.Index(url[start:], "&")
	if end == -1 {
		return url[start:]
	}
	return url[start : start+end]
}
