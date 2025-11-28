package conversions

import (
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

func NewAccept(id, actor, object *url.URL) (a vocab.ActivityStreamsAccept) {
	a = streams.NewActivityStreamsAccept()
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(id)
	a.SetJSONLDId(idProp)

	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(actor)
	a.SetActivityStreamsActor(actorProp)

	objProp := streams.NewActivityStreamsObject()
	objId := streams.NewJSONLDIdProperty()
	objId.SetIRI(object)
	objProp.SetJSONLDId(objId)

	return
}

func GroupToActor(g domain.Collective) vocab.Type {
	obj := streams.NewActivityStreamsGroup()
	
	id := streams.NewJSONLDIdProperty()
	id.SetIRI(g.Url)
	obj.SetJSONLDId(id)

	name := streams.NewActivityStreamsPreferredUsernameProperty()
	name.SetXMLSchemaString(g.Name)
	obj.SetActivityStreamsPreferredUsername(name)


	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("")
	obj.SetActivityStreamsSummary(summary)

	var keyFragment, _ = url.Parse("#main-key")
	keyProp := streams.NewW3IDSecurityV1PublicKeyProperty()

	key := streams.NewW3IDSecurityV1PublicKey()

	keyId := streams.NewJSONLDIdProperty()
	keyURI := id.GetIRI().ResolveReference(keyFragment)
	log.Debug().Str("key", keyURI.String()).Msg("at conversions")
	keyId.SetIRI(keyURI)
	key.SetJSONLDId(keyId)

	keyPem := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	keyPem.Set(g.Public_key)
	key.SetW3IDSecurityV1PublicKeyPem(keyPem)

	owner := streams.NewW3IDSecurityV1OwnerProperty()
	owner.SetIRI(g.Url)
	key.SetW3IDSecurityV1Owner(owner)

	keyProp.AppendW3IDSecurityV1PublicKey(key)

	obj.SetW3IDSecurityV1PublicKey(keyProp)

	inbox := streams.NewActivityStreamsInboxProperty()
	inbox.SetIRI(g.Inbox)
	obj.SetActivityStreamsInbox(inbox)

	outbox := streams.NewActivityStreamsOutboxProperty()
	outbox.SetIRI(g.Outbox)
	obj.SetActivityStreamsOutbox(outbox)

	if g.Followers != nil {
		followers := streams.NewActivityStreamsFollowersProperty()
		followers.SetIRI(g.Followers)
		obj.SetActivityStreamsFollowers(followers)
	}

	if g.Url != nil {
		url := streams.NewActivityStreamsUrlProperty()
		url.AppendIRI(g.Url)
		obj.SetActivityStreamsUrl(url)
	}

	return obj
}


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
