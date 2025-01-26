package identity

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/youmark/pkcs8"
)

// GetTokenString retrieves bearer token string from the specified request
func GetTokenString(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	temp := strings.Split(header, " ")
	if len(temp) != 2 {
		return ""
	}
	if temp[0] != "Bearer" {
		return ""
	}
	return temp[1]
}

// GetPublicKey extracts RSA public key from PEM file of public key
func GetPublicKey(path string) (*rsa.PublicKey, error) {
	pemBlock, errFile := getPemBlock(path)
	if errFile != nil {
		return nil, errFile
	}

	key, errParse := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("unable to parse pulbic key from file %s with error %w", path, errParse)
	}

	switch keyType := key.(type) {
	case *rsa.PublicKey:
		return keyType, nil
	default:
		return nil, fmt.Errorf("unsupported key type %T", key)
	}
}

// GetPublicKeyFromCertificate extracts RSA public key from PEM file of certificate
func GetPublicKeyFromCertificate(path string) (*rsa.PublicKey, error) {
	pemBlock, errFile := getPemBlock(path)
	if errFile != nil {
		return nil, errFile
	}

	cert, errParse := x509.ParseCertificate(pemBlock.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("unable to parse pulbic key from file %s with error %w", path, errParse)
	}

	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return cert.PublicKey.(*rsa.PublicKey), nil
	default:
		return nil, fmt.Errorf("unsupported key type %T", cert.PublicKey)
	}
}

// GetPrivateKey extract encrypted RSA private key from PEM file
func GetPrivateKey(path string, password string) (*rsa.PrivateKey, error) {
	pemBlock, errFile := getPemBlock(path)
	if errFile != nil {
		return nil, errFile
	}

	key, errDecrpt := pkcs8.ParsePKCS8PrivateKey(pemBlock.Bytes, []byte(password))
	if errDecrpt != nil {
		return nil, fmt.Errorf("unable to decrypted key from file %s with error %w", path, errDecrpt)
	}

	switch keyType := key.(type) {
	case *rsa.PrivateKey:
		return keyType, nil
	default:
		return nil, fmt.Errorf("unsupported key type %T", key)
	}
}

func getPemBlock(path string) (*pem.Block, error) {
	file, errOpen := os.Open(path)
	if errOpen != nil {
		return nil, fmt.Errorf("unable to open file %s with error %w", path, errOpen)
	}
	defer file.Close()

	bytes, errRead := io.ReadAll(file)
	if errRead != nil {
		return nil, fmt.Errorf("unable to read file %s with error %w", path, errRead)
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, fmt.Errorf("unable to decode PEM file %s", path)
	}
	return block, nil
}
