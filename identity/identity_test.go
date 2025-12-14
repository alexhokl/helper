package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/youmark/pkcs8"
)

func TestGetTokenString(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
	}{
		{
			name:       "valid bearer token",
			authHeader: "Bearer abc123token",
			want:       "abc123token",
		},
		{
			name:       "valid bearer token with long token",
			authHeader: "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature",
			want:       "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature",
		},
		{
			name:       "empty header",
			authHeader: "",
			want:       "",
		},
		{
			name:       "missing Bearer prefix",
			authHeader: "abc123token",
			want:       "",
		},
		{
			name:       "wrong auth type Basic",
			authHeader: "Basic abc123token",
			want:       "",
		},
		{
			name:       "wrong auth type lowercase bearer",
			authHeader: "bearer abc123token",
			want:       "",
		},
		{
			name:       "too many parts",
			authHeader: "Bearer token extra",
			want:       "",
		},
		{
			name:       "Bearer only no token",
			authHeader: "Bearer",
			want:       "",
		},
		{
			name:       "Bearer with empty token",
			authHeader: "Bearer ",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			got := GetTokenString(req)
			if got != tt.want {
				t.Errorf("GetTokenString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTokenStringWithRealRequest(t *testing.T) {
	// Test with a real http.Request object
	req, err := http.NewRequest("GET", "http://example.com/api", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer mytoken123")

	got := GetTokenString(req)
	if got != "mytoken123" {
		t.Errorf("GetTokenString() = %q, want %q", got, "mytoken123")
	}
}

// generateTestRSAKey generates a test RSA key pair
func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	return privateKey
}

// createTestPublicKeyPEM creates a PEM file with a public key
func createTestPublicKeyPEM(t *testing.T, dir string, publicKey *rsa.PublicKey) string {
	t.Helper()
	path := filepath.Join(dir, "public.pem")

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		t.Fatalf("Failed to encode PEM: %v", err)
	}

	return path
}

// createTestCertificatePEM creates a PEM file with a self-signed certificate
func createTestCertificatePEM(t *testing.T, dir string, privateKey *rsa.PrivateKey) string {
	t.Helper()
	path := filepath.Join(dir, "cert.pem")

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		t.Fatalf("Failed to encode PEM: %v", err)
	}

	return path
}

// createTestEncryptedPrivateKeyPEM creates a PEM file with an encrypted private key
func createTestEncryptedPrivateKeyPEM(t *testing.T, dir string, privateKey *rsa.PrivateKey, password string) string {
	t.Helper()
	path := filepath.Join(dir, "private_encrypted.pem")

	// Encrypt using PKCS8
	encryptedBytes, err := pkcs8.MarshalPrivateKey(privateKey, []byte(password), nil)
	if err != nil {
		t.Fatalf("Failed to encrypt private key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "ENCRYPTED PRIVATE KEY",
		Bytes: encryptedBytes,
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		t.Fatalf("Failed to encode PEM: %v", err)
	}

	return path
}

func TestGetPublicKey(t *testing.T) {
	// Generate test key
	privateKey := generateTestRSAKey(t)

	// Create temp directory
	tmpDir := t.TempDir()

	// Create test public key PEM file
	pubKeyPath := createTestPublicKeyPEM(t, tmpDir, &privateKey.PublicKey)

	// Test loading the public key
	loadedKey, err := GetPublicKey(pubKeyPath)
	if err != nil {
		t.Fatalf("GetPublicKey() error: %v", err)
	}

	if loadedKey == nil {
		t.Fatal("GetPublicKey() returned nil key")
	}

	// Verify the key matches
	if loadedKey.N.Cmp(privateKey.PublicKey.N) != 0 {
		t.Error("GetPublicKey() returned different key")
	}
	if loadedKey.E != privateKey.PublicKey.E {
		t.Error("GetPublicKey() returned different exponent")
	}
}

func TestGetPublicKeyFileNotFound(t *testing.T) {
	_, err := GetPublicKey("/nonexistent/path/to/key.pem")
	if err == nil {
		t.Error("GetPublicKey() with non-existent file should return error")
	}
}

func TestGetPublicKeyInvalidPEM(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.pem")

	// Create file with invalid content
	if err := os.WriteFile(invalidPath, []byte("not a valid PEM file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := GetPublicKey(invalidPath)
	if err == nil {
		t.Error("GetPublicKey() with invalid PEM should return error")
	}
}

func TestGetPublicKeyFromCertificate(t *testing.T) {
	// Generate test key
	privateKey := generateTestRSAKey(t)

	// Create temp directory
	tmpDir := t.TempDir()

	// Create test certificate PEM file
	certPath := createTestCertificatePEM(t, tmpDir, privateKey)

	// Test loading the public key from certificate
	loadedKey, err := GetPublicKeyFromCertificate(certPath)
	if err != nil {
		t.Fatalf("GetPublicKeyFromCertificate() error: %v", err)
	}

	if loadedKey == nil {
		t.Fatal("GetPublicKeyFromCertificate() returned nil key")
	}

	// Verify the key matches
	if loadedKey.N.Cmp(privateKey.PublicKey.N) != 0 {
		t.Error("GetPublicKeyFromCertificate() returned different key")
	}
	if loadedKey.E != privateKey.PublicKey.E {
		t.Error("GetPublicKeyFromCertificate() returned different exponent")
	}
}

func TestGetPublicKeyFromCertificateFileNotFound(t *testing.T) {
	_, err := GetPublicKeyFromCertificate("/nonexistent/path/to/cert.pem")
	if err == nil {
		t.Error("GetPublicKeyFromCertificate() with non-existent file should return error")
	}
}

func TestGetPublicKeyFromCertificateInvalidPEM(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.pem")

	// Create file with invalid content
	if err := os.WriteFile(invalidPath, []byte("not a valid PEM file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := GetPublicKeyFromCertificate(invalidPath)
	if err == nil {
		t.Error("GetPublicKeyFromCertificate() with invalid PEM should return error")
	}
}

func TestGetPrivateKey(t *testing.T) {
	// Generate test key
	privateKey := generateTestRSAKey(t)
	password := "testpassword123"

	// Create temp directory
	tmpDir := t.TempDir()

	// Create test encrypted private key PEM file
	privKeyPath := createTestEncryptedPrivateKeyPEM(t, tmpDir, privateKey, password)

	// Test loading the private key
	loadedKey, err := GetPrivateKey(privKeyPath, password)
	if err != nil {
		t.Fatalf("GetPrivateKey() error: %v", err)
	}

	if loadedKey == nil {
		t.Fatal("GetPrivateKey() returned nil key")
	}

	// Verify the key matches
	if loadedKey.D.Cmp(privateKey.D) != 0 {
		t.Error("GetPrivateKey() returned different key")
	}
}

func TestGetPrivateKeyWrongPassword(t *testing.T) {
	// Generate test key
	privateKey := generateTestRSAKey(t)
	password := "correctpassword"

	// Create temp directory
	tmpDir := t.TempDir()

	// Create test encrypted private key PEM file
	privKeyPath := createTestEncryptedPrivateKeyPEM(t, tmpDir, privateKey, password)

	// Test loading with wrong password
	_, err := GetPrivateKey(privKeyPath, "wrongpassword")
	if err == nil {
		t.Error("GetPrivateKey() with wrong password should return error")
	}
}

func TestGetPrivateKeyFileNotFound(t *testing.T) {
	_, err := GetPrivateKey("/nonexistent/path/to/key.pem", "password")
	if err == nil {
		t.Error("GetPrivateKey() with non-existent file should return error")
	}
}

func TestGetPrivateKeyInvalidPEM(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.pem")

	// Create file with invalid content
	if err := os.WriteFile(invalidPath, []byte("not a valid PEM file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := GetPrivateKey(invalidPath, "password")
	if err == nil {
		t.Error("GetPrivateKey() with invalid PEM should return error")
	}
}

func TestGetPemBlock(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid PEM file
	validPath := filepath.Join(tmpDir, "valid.pem")
	pemContent := `-----BEGIN TEST-----
dGVzdCBkYXRh
-----END TEST-----`
	if err := os.WriteFile(validPath, []byte(pemContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	block, err := getPemBlock(validPath)
	if err != nil {
		t.Fatalf("getPemBlock() error: %v", err)
	}

	if block == nil {
		t.Fatal("getPemBlock() returned nil block")
	}

	if block.Type != "TEST" {
		t.Errorf("getPemBlock() block.Type = %q, want %q", block.Type, "TEST")
	}
}

func TestGetPemBlockEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyPath := filepath.Join(tmpDir, "empty.pem")

	// Create empty file
	if err := os.WriteFile(emptyPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := getPemBlock(emptyPath)
	if err == nil {
		t.Error("getPemBlock() with empty file should return error")
	}
}

func TestGetPemBlockInvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.pem")

	// Create file with non-PEM content
	if err := os.WriteFile(invalidPath, []byte("this is not PEM data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := getPemBlock(invalidPath)
	if err == nil {
		t.Error("getPemBlock() with invalid content should return error")
	}
}

// Benchmark tests
func BenchmarkGetTokenString(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature")

	for i := 0; i < b.N; i++ {
		GetTokenString(req)
	}
}

func BenchmarkGetTokenStringNoHeader(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)

	for i := 0; i < b.N; i++ {
		GetTokenString(req)
	}
}

func BenchmarkGetPublicKey(b *testing.B) {
	// Generate test key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Create temp directory and public key file
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "public.pem")

	pubKeyBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	pemBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyBytes}
	file, _ := os.Create(path)
	pem.Encode(file, pemBlock)
	file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetPublicKey(path)
	}
}
