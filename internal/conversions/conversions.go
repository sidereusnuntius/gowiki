package conversions

import (
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func UserToActor(u domain.UserFed) vocab.Type {
	a := streams.NewActivityStreamsPerson()

	id := streams.NewJSONLDIdProperty()
	id.SetIRI(u.ApId)
	a.SetJSONLDId(id)

	username := streams.NewActivityStreamsPreferredUsernameProperty()
	username.SetXMLSchemaString(u.Username)
	a.SetActivityStreamsPreferredUsername(username)

	if u.Name != "" {
		name := streams.NewActivityStreamsNameProperty()
		name.AppendXMLSchemaString(u.Name)
		a.SetActivityStreamsName(name)
	}

	if u.Summary != "" {
		summary := streams.NewActivityStreamsSummaryProperty()
		summary.AppendXMLSchemaString(u.Summary)
		a.SetActivityStreamsSummary(summary)
	}

	if u.URL != nil {
		iri := streams.NewActivityStreamsUrlProperty()
		iri.AppendIRI(u.URL)
		a.SetActivityStreamsUrl(iri)
	}

	inbox := streams.NewActivityStreamsInboxProperty()
	inbox.SetIRI(u.Inbox)
	a.SetActivityStreamsInbox(inbox)

	outbox := streams.NewActivityStreamsOutboxProperty()
	outbox.SetIRI(u.Outbox)
	a.SetActivityStreamsOutbox(outbox)

	followers := streams.NewActivityStreamsFollowersProperty()
	followers.SetIRI(u.Followers)
	a.SetActivityStreamsFollowers(followers)

	created := streams.NewActivityStreamsPublishedProperty()
	created.Set(u.Created)

	updated := streams.NewActivityStreamsUpdatedProperty()
	updated.Set(u.LastUpdated)
	a.SetActivityStreamsPublished(created)
	a.SetActivityStreamsUpdated(updated)

	a.SetW3IDSecurityV1PublicKey(PublicKeyProp(u.ApId, u.PublicKey))

	return a
}

func PublicKeyProp(owner *url.URL, publicKeyPem string) vocab.W3IDSecurityV1PublicKeyProperty {
	keyProp := streams.NewW3IDSecurityV1PublicKeyProperty()
	key := streams.NewW3IDSecurityV1PublicKey()

	ownerProp := streams.NewW3IDSecurityV1OwnerProperty()
	ownerProp.SetIRI(owner)

	// TODO: improve this.
	keyURI := owner.JoinPath("#main-key")
	keyURIProp := streams.NewJSONLDIdProperty()
	keyURIProp.SetIRI(keyURI)

	pemProp := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	pemProp.Set(publicKeyPem)

	key.SetJSONLDId(keyURIProp)
	key.SetW3IDSecurityV1PublicKeyPem(pemProp)
	key.SetW3IDSecurityV1Owner(ownerProp)
	key.SetW3IDSecurityV1PublicKeyPem(pemProp)

	keyProp.AppendW3IDSecurityV1PublicKey(key)
	return keyProp
}
