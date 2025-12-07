package conversions

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/sidereusnuntius/gowiki/internal/federation"
)

func ExtractPublicKeyFromActor(actor WithPublicKeyProperty) (string, error) {
	pubKeyProp := actor.GetW3IDSecurityV1PublicKey()
	if pubKeyProp == nil || pubKeyProp.Len() == 0 {
		return "", fmt.Errorf("%w: public key", federation.ErrMissingProperty)
	}

	keyPemProp := pubKeyProp.Begin().Get().GetW3IDSecurityV1PublicKeyPem()
	if keyPemProp == nil {
		return "", fmt.Errorf("%w: publicKeyPem", federation.ErrMissingProperty)
	}
	return keyPemProp.Get(), nil
}

func ExtractPublicKeyFromPem(block pem.Block) (crypto.PublicKey, error) {
	var pubKey crypto.PublicKey
	var err error
	switch block.Type {
	case "PUBLIC KEY":
		pubKey, err = x509.ParsePKIXPublicKey(block.Bytes)
	case "RSA PUBLIC KEY":
		pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		err = fmt.Errorf("unsupported type: %s", block.Type)
	}

	if err != nil {
		return nil, err
	}
	return pubKey, nil
}