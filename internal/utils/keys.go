package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func GenerateKeysPem(size int) (pub string, priv string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return
	}

	priv, err = privateKeyPem(key)
	if err != nil {
		return
	}

	pub, err = publicKeyPem(&key.PublicKey)
	return
}

func privateKeyPem(key *rsa.PrivateKey) (string, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})), nil
}

func publicKeyPem(key *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})), err
}
