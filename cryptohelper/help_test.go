package cryptohelper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/youmark/pkcs8"
)

// generateTestECDSAKey generates a test ECDSA key pair
func generateTestECDSAKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}
	return privateKey
}

// createTestEncryptedECDSAPrivateKeyPEM creates a PEM file with an encrypted ECDSA private key
func createTestEncryptedECDSAPrivateKeyPEM(t *testing.T, dir string, privateKey *ecdsa.PrivateKey, password string) string {
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

// readPEMFile reads a PEM file and returns its bytes
func readPEMFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	return data
}

func TestGetEcdsaKey(t *testing.T) {
	// Generate test key
	privateKey := generateTestECDSAKey(t)
	password := "testpassword123"

	// Create temp directory
	tmpDir := t.TempDir()

	// Create test encrypted private key PEM file
	privKeyPath := createTestEncryptedECDSAPrivateKeyPEM(t, tmpDir, privateKey, password)
	privateKeyBytes := readPEMFile(t, privKeyPath)

	// Test loading the private key
	loadedKey, err := GetEcdsaKey(privateKeyBytes, []byte(password))
	if err != nil {
		t.Fatalf("GetEcdsaKey() error: %v", err)
	}

	if loadedKey == nil {
		t.Fatal("GetEcdsaKey() returned nil key")
	}

	// Verify the key matches
	if loadedKey.D.Cmp(privateKey.D) != 0 {
		t.Error("GetEcdsaKey() returned different D value")
	}
	if loadedKey.X.Cmp(privateKey.X) != 0 {
		t.Error("GetEcdsaKey() returned different X value")
	}
	if loadedKey.Y.Cmp(privateKey.Y) != 0 {
		t.Error("GetEcdsaKey() returned different Y value")
	}
}

func TestGetEcdsaKeyWrongPassword(t *testing.T) {
	// Generate test key
	privateKey := generateTestECDSAKey(t)
	password := "correctpassword"

	// Create temp directory
	tmpDir := t.TempDir()

	// Create test encrypted private key PEM file
	privKeyPath := createTestEncryptedECDSAPrivateKeyPEM(t, tmpDir, privateKey, password)
	privateKeyBytes := readPEMFile(t, privKeyPath)

	// Test loading with wrong password
	_, err := GetEcdsaKey(privateKeyBytes, []byte("wrongpassword"))
	if err == nil {
		t.Error("GetEcdsaKey() with wrong password should return error")
	}
}

func TestGetEcdsaKeyInvalidPEM(t *testing.T) {
	// Test with invalid PEM content
	invalidPEM := []byte("not a valid PEM file")

	_, err := GetEcdsaKey(invalidPEM, []byte("password"))
	if err == nil {
		t.Error("GetEcdsaKey() with invalid PEM should return error")
	}
}

func TestGetEcdsaKeyEmptyInput(t *testing.T) {
	// Test with empty input
	_, err := GetEcdsaKey([]byte{}, []byte("password"))
	if err == nil {
		t.Error("GetEcdsaKey() with empty input should return error")
	}
}

func TestGetEcdsaKeyNilInput(t *testing.T) {
	// Test with nil input
	_, err := GetEcdsaKey(nil, []byte("password"))
	if err == nil {
		t.Error("GetEcdsaKey() with nil input should return error")
	}
}

func TestGetEcdsaKeyEmptyPassword(t *testing.T) {
	// Generate test key with empty password (unencrypted)
	privateKey := generateTestECDSAKey(t)

	// Create temp directory
	tmpDir := t.TempDir()

	// Create an unencrypted PKCS8 PEM file
	path := filepath.Join(tmpDir, "private_unencrypted.pem")

	// Marshal as unencrypted PKCS8
	unencryptedBytes, err := pkcs8.MarshalPrivateKey(privateKey, nil, nil)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: unencryptedBytes,
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		t.Fatalf("Failed to encode PEM: %v", err)
	}

	privateKeyBytes := readPEMFile(t, path)

	// Test loading with empty password (should work for unencrypted key)
	loadedKey, err := GetEcdsaKey(privateKeyBytes, nil)
	if err != nil {
		t.Fatalf("GetEcdsaKey() with unencrypted key error: %v", err)
	}

	if loadedKey == nil {
		t.Fatal("GetEcdsaKey() returned nil key")
	}

	// Verify the key matches
	if loadedKey.D.Cmp(privateKey.D) != 0 {
		t.Error("GetEcdsaKey() returned different D value")
	}
}

func TestGetEcdsaKeyDifferentCurves(t *testing.T) {
	curves := []struct {
		name  string
		curve elliptic.Curve
	}{
		{"P224", elliptic.P224()},
		{"P256", elliptic.P256()},
		{"P384", elliptic.P384()},
		{"P521", elliptic.P521()},
	}

	for _, tc := range curves {
		t.Run(tc.name, func(t *testing.T) {
			// Generate key with specific curve
			privateKey, err := ecdsa.GenerateKey(tc.curve, rand.Reader)
			if err != nil {
				t.Fatalf("Failed to generate ECDSA key with curve %s: %v", tc.name, err)
			}

			password := "testpassword"

			// Create temp directory
			tmpDir := t.TempDir()

			// Create test encrypted private key PEM file
			path := filepath.Join(tmpDir, "private.pem")

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

			if err := pem.Encode(file, pemBlock); err != nil {
				file.Close()
				t.Fatalf("Failed to encode PEM: %v", err)
			}
			file.Close()

			privateKeyBytes := readPEMFile(t, path)

			// Test loading the private key
			loadedKey, err := GetEcdsaKey(privateKeyBytes, []byte(password))
			if err != nil {
				t.Fatalf("GetEcdsaKey() error: %v", err)
			}

			if loadedKey == nil {
				t.Fatal("GetEcdsaKey() returned nil key")
			}

			// Verify the key matches
			if loadedKey.D.Cmp(privateKey.D) != 0 {
				t.Error("GetEcdsaKey() returned different D value")
			}

			// Verify curve matches
			if loadedKey.Curve.Params().Name != tc.curve.Params().Name {
				t.Errorf("GetEcdsaKey() curve = %s, want %s", loadedKey.Curve.Params().Name, tc.curve.Params().Name)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetEcdsaKey(b *testing.B) {
	// Generate test key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatalf("Failed to generate ECDSA key: %v", err)
	}
	password := "benchmarkpassword"

	// Create temp directory and encrypted key file
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "private.pem")

	encryptedBytes, _ := pkcs8.MarshalPrivateKey(privateKey, []byte(password), nil)
	pemBlock := &pem.Block{Type: "ENCRYPTED PRIVATE KEY", Bytes: encryptedBytes}
	file, _ := os.Create(path)
	pem.Encode(file, pemBlock)
	file.Close()

	privateKeyBytes, _ := os.ReadFile(path)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetEcdsaKey(privateKeyBytes, []byte(password))
	}
}

func BenchmarkGetEcdsaKeyUnencrypted(b *testing.B) {
	// Generate test key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	// Create temp directory and unencrypted key file
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "private.pem")

	unencryptedBytes, _ := pkcs8.MarshalPrivateKey(privateKey, nil, nil)
	pemBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: unencryptedBytes}
	file, _ := os.Create(path)
	pem.Encode(file, pemBlock)
	file.Close()

	privateKeyBytes, _ := os.ReadFile(path)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetEcdsaKey(privateKeyBytes, nil)
	}
}
