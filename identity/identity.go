package identity

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
		return nil, fmt.Errorf("Unable to parse pulbic key from file %s with error %w", path, errParse)
	}

	switch key.(type) {
	case *rsa.PublicKey:
		return key.(*rsa.PublicKey), nil
	default:
		return nil, fmt.Errorf("Unsupported key type %T", key)
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
		return nil, fmt.Errorf("Unable to parse pulbic key from file %s with error %w", path, errParse)
	}

	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return cert.PublicKey.(*rsa.PublicKey), nil
	default:
		return nil, fmt.Errorf("Unsupported key type %T", cert.PublicKey)
	}
}

// GetPrivateKey extract encrypted RSA private key from PEM file
func GetPrivateKey(path string, password string) (*rsa.PrivateKey, error) {
	pemBlock, errFile := getPemBlock(path)
	if errFile != nil {
		return nil, errFile
	}

	decryptedBytes, errDecrpt := x509.DecryptPEMBlock(pemBlock, []byte(password))
	if errDecrpt != nil {
		return nil, fmt.Errorf("Unable to decrypted key from file %s with error %w", path, errDecrpt)
	}

	key, errParse := x509.ParsePKCS1PrivateKey(decryptedBytes)
	if errParse != nil {
		return nil, fmt.Errorf("Unable to parse private key from file %s with error %w", path, errParse)
	}
	return key, nil
}

func getPemBlock(path string) (*pem.Block, error) {
	file, errOpen := os.Open(path)
	if errOpen != nil {
		return nil, fmt.Errorf("Unable to open file %s with error %w", path, errOpen)
	}
	defer file.Close()

	bytes, errRead := ioutil.ReadAll(file)
	if errRead != nil {
		return nil, fmt.Errorf("Unable to read file %s with error %w", path, errRead)
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, fmt.Errorf("Unable to decode PEM file %s", path)
	}
	return block, nil
}
