package domain

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"

	"os"
)

func GetPrivateRSAKey(name string) (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(filepath.Join(".ssh", name))
	if err != nil {
		return nil, fmt.Errorf("Failed to read key file: %w", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("Failed to decode PEM block: invalid PEM data")
	}

	switch block.Type {
	case "RSA PRIVATE KEY": // PKCS#1 format
		rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse PKCS#1 private key: %w", err)
		}
		return rsaKey, nil

	case "PRIVATE KEY": // PKCS#8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse PKCS#8 private key: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("Key in PKCS#8 block is not an RSA private key")
		}
		return rsaKey, nil

	default:
		return nil, fmt.Errorf("Unsupported PEM block type: %s", block.Type)
	}
}

func GetPublicRSAKey(name string) (*rsa.PublicKey, error) {
	pemBytes, err := os.ReadFile(filepath.Join(".ssh", name))
	if err != nil {
		return nil, fmt.Errorf("Failed to read key file: %w", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("Failed to decode PEM block: invalid or missing PEM data")
	}

	switch block.Type {
	case "PUBLIC KEY":
		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse PKIX public key: %w", err)
		}

		rsaPub, ok := pubInterface.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("Key in PEM block is not an RSA public key")
		}
		return rsaPub, nil

	case "RSA PUBLIC KEY":
		rsaPub, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse PKCS#1 public key: %w", err)
		}
		return rsaPub, nil

	default:
		return nil, fmt.Errorf("Unsupported PEM block type: %s", block.Type)
	}
}
