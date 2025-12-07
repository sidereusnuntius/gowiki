package conversions

import (
	"net/url"

	"code.superseriousbusiness.org/activity/streams/vocab"
)

type IriProperty interface {
	IsIRI() bool
	GetIRI() *url.URL
}

type WithPublicKeyProperty interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
	SetW3IDSecurityV1PublicKey(i vocab.W3IDSecurityV1PublicKeyProperty)
}
