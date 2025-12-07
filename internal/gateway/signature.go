package gateway

import (
	"context"
	"crypto"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/httpsig"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

func (g *FedGatewayImpl) GetPublicKey(ctx context.Context, keyId *url.URL) (crypto.PublicKey, error) {
	var keyStr string

	keyId.RawFragment = ""
	keyStr, err := g.db.GetPublicKeyByActorIRI(ctx, keyId)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, err
		}

		actor, err := g.client.Get(ctx, keyId)
		if err != nil {
			return nil, err
		}

		withKey, ok := actor.(conversions.WithPublicKeyProperty)
		if !ok {
			return nil, fmt.Errorf("%w: public key", federation.ErrMissingProperty)
		}

		keyStr, err = conversions.ExtractPublicKeyFromActor(withKey)
		if err != nil {
			return nil, err
		}

		if err = g.ProcessObject(ctx, actor); err != nil {
			return nil, err
		}
	}

	fmt.Println("Public key:\n", keyStr)
	block, _ := pem.Decode([]byte(keyStr))
	pubKey, err := conversions.ExtractPublicKeyFromPem(*block)

	return pubKey, err
}

func (g *FedGatewayImpl) Verify(ctx context.Context, r *http.Request) error {
	verifier, err := httpsig.NewVerifier(r)
	if err != nil {
		return err
	}

	keyId, err := url.Parse(verifier.KeyId())
	if err != nil {
		return fmt.Errorf("unable to parse keyId: %s", verifier.KeyId())
	}
	key, err := g.GetPublicKey(ctx, keyId)
	if err != nil {
		return err
	}

	err = verifier.Verify(key, httpsig.RSA_SHA256)
	if err == nil {
		fmt.Printf("authentication succeeded")
	}
	return err
}