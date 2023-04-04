package cryptohelper

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/exp/slog"
)

func GetEcdsaKey(privateKeyBytes []byte, passphraseBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("unable to decode private key")
	}

	if !x509.IsEncryptedPEMBlock(block) {
		return nil, fmt.Errorf("private key is not encrypted")
	}

	slog.Debug("Decoded private key",
		slog.String("key_type", block.Type),
	)

	decryptedBlockBytes, err := x509.DecryptPEMBlock(block, passphraseBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt private key: %w", err)
	}

	return x509.ParseECPrivateKey(decryptedBlockBytes)
}
