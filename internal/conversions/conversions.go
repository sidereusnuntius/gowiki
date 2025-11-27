package conversions

import (
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

func ActorToUser(a vocab.ActivityStreamsPerson) (u domain.UserFed, err error) {
	idProp := a.GetJSONLDId()
	if idProp == nil {
		err = fmt.Errorf("%w: id", federation.ErrMissingProperty)
	}
	id := idProp.Get()
	u.ApId = id

	if name := a.GetActivityStreamsName(); name != nil && name.Len() != 0 {
		u.Name = name.Begin().GetXMLSchemaString()
	}

	if username := a.GetActivityStreamsPreferredUsername(); username != nil {
		u.Username = username.GetXMLSchemaString()
	}

	u.Domain = id.Host
	if summary := a.GetActivityStreamsSummary(); summary != nil && summary.Len() != 0 {
		u.Summary = summary.Begin().GetXMLSchemaString()
	}

	if url := a.GetActivityStreamsUrl(); url != nil && url.Len() != 0 {
		u.URL = url.Begin().GetIRI()
	}

	inbox := a.GetActivityStreamsInbox()

	if inbox == nil {
		err = fmt.Errorf("%w: inbox", federation.ErrMissingProperty)
		return
	}
	if !inbox.IsIRI() {
		err = fmt.Errorf("%w: inbox", federation.ErrUnprocessablePropValue)
	}
	u.Inbox = inbox.GetIRI()

	if outbox := a.GetActivityStreamsOutbox(); outbox != nil && outbox.IsIRI() {
		u.Outbox = outbox.GetIRI()
	}

    if followers := a.GetActivityStreamsFollowers(); followers != nil && followers.IsIRI() {
		u.Followers = followers.GetIRI()
	}

    if key := a.GetW3IDSecurityV1PublicKey(); key != nil && key.Len() != 0 {
		k := key.Begin().Get()
		keyPem := k.GetW3IDSecurityV1PublicKeyPem()
		u.PublicKey = keyPem.Get()
	}
    
	if created := a.GetActivityStreamsPublished(); created != nil {
		// Perhaps use it?
	}
    
	if updated := a.GetActivityStreamsUpdated(); updated != nil {

	}
	
	return
}

func ArticleToObject(a domain.ArticleFed) vocab.Type {
	o := streams.NewActivityStreamsArticle()
	id := streams.NewJSONLDIdProperty()
	id.SetIRI(a.ApID)
	o.SetJSONLDId(id)

	if a.Title != "" {
		title := streams.NewActivityStreamsNameProperty()
		title.AppendXMLSchemaString(a.Title)
		o.SetActivityStreamsName(title)
	}

	if a.Summary != "" {
		summary := streams.NewActivityStreamsSummaryProperty()
		summary.AppendXMLSchemaString(a.Summary)
		o.SetActivityStreamsSummary(summary)
	}

	if a.MediaType != "" {
		mt := streams.NewActivityStreamsMediaTypeProperty()
		mt.Set(a.MediaType)
		o.SetActivityStreamsMediaType(mt)
	}

	if a.Url != nil {
		u := streams.NewActivityStreamsUrlProperty()
		u.AppendIRI(a.Url)
		o.SetActivityStreamsUrl(u)
	}

	content := streams.NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString(a.Content)
	o.SetActivityStreamsContent(content)

	return o
}

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
