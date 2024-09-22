package signer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log/slog"

	"github.com/spf13/viper"
)

type Signer interface {
	Sign(ctx context.Context, data []byte) ([]byte, error)
	GetPublicKey(ctx context.Context, format string) ([]byte, error)
}

func NewWithConfig(config *viper.Viper) (Signer, error) {
	var privateKey *rsa.PrivateKey
	var err error

	keyStr := config.GetString("signing.key")
	if keyStr == "" {
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		slog.Warn("A private signing key has been generated. To make it permanent, specify the valid RSA private key in the config parameter signing.key")
	} else {
		keyBytes := []byte(keyStr)
		rawPem, _ := pem.Decode(keyBytes)
		if rawPem == nil {
			return nil, errors.New("unable to decode pem key")
		}

		privateKey, err = x509.ParsePKCS1PrivateKey(rawPem.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return NewLocal(privateKey), nil
}
