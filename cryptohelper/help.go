package cryptohelper

import (
	"crypto/ecdsa"
	"encoding/pem"
	"fmt"

	"log/slog"

	"github.com/youmark/pkcs8"
)

func GetEcdsaKey(privateKeyBytes []byte, passphraseBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("unable to decode private key")
	}

	slog.Debug("Decoded private key",
		slog.String("key_type", block.Type),
	)

	decryptedBlockBytes, err := pkcs8.ParsePKCS8PrivateKeyECDSA(block.Bytes, passphraseBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt private key: %w", err)
	}

	return decryptedBlockBytes, nil
}
