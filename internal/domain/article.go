package domain

import (
	"net/url"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
)

type ArticleCore struct {
	Title       string
	Host string
	Summary     string
	Content     string
	Protected   bool
	MediaType   string
	License     string
	Language    string
	LastUpdated time.Time
	Published   time.Time
}

type ArticleFed struct {
	ArticleCore
	AttributedTo *url.URL
	ApID         *url.URL
	Url          *url.URL
	To           []*url.URL
}

type Revision struct {
	ID       int64
	Title    string
	Reviewed bool
	Diff     string
	Summary  string
	Username string
	Created  int64
}

func (a *ArticleFed) ConvertToAp() vocab.ActivityStreamsArticle {
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

	if a.AttributedTo != nil {
		att := streams.NewActivityStreamsAttributedToProperty()
		att.AppendIRI(a.AttributedTo)
		o.SetActivityStreamsAttributedTo(att)
	}

	content := streams.NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString(a.Content)
	o.SetActivityStreamsContent(content)

	toProp := streams.NewActivityStreamsToProperty()
	for _, uri := range a.To {
		toProp.AppendIRI(uri)
	}
	o.SetActivityStreamsTo(toProp)

	pub := streams.NewActivityStreamsPublishedProperty()
	pub.Set(a.Published)
	o.SetActivityStreamsPublished(pub)

	if !a.LastUpdated.IsZero() {
		updated := streams.NewActivityStreamsUpdatedProperty()
		updated.Set(a.LastUpdated)
		o.SetActivityStreamsUpdated(updated)
	}

	return o
}

func (a *ArticleFed) UpdateAP(id, author, wiki *url.URL, summary string) (update vocab.ActivityStreamsUpdate) {
	update = streams.NewActivityStreamsUpdate()

	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(id)
	update.SetJSONLDId(idProp)

	actor := streams.NewActivityStreamsActorProperty()
	actor.AppendIRI(author)
	update.SetActivityStreamsActor(actor)

	att := streams.NewActivityStreamsAttributedToProperty()
	att.AppendIRI(wiki)
	update.SetActivityStreamsAttributedTo(att)

	obj := streams.NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsArticle(a.ConvertToAp())
	update.SetActivityStreamsObject(obj)

	published := streams.NewActivityStreamsPublishedProperty()
	published.Set(a.LastUpdated)
	update.SetActivityStreamsPublished(published)

	cc := streams.NewActivityStreamsCcProperty()
	cc.AppendIRI(wiki.JoinPath("followers"))
	update.SetActivityStreamsCc(cc)

	return
}

func (a *ArticleFed) CreateAP(id, author, wikiId *url.URL, summary string) (c vocab.ActivityStreamsCreate) {
	c = streams.NewActivityStreamsCreate()

	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(id)
	c.SetJSONLDId(idProp)

	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(wikiId)
	c.SetActivityStreamsActor(actorProp)

	as := a.ConvertToAp()
	objProp := streams.NewActivityStreamsObjectProperty()
	objProp.AppendActivityStreamsArticle(as)
	c.SetActivityStreamsObject(objProp)

	to := streams.NewActivityStreamsToProperty()
	to.AppendIRI(Public)
	c.SetActivityStreamsTo(to)

	cc := streams.NewActivityStreamsCcProperty()
	cc.AppendIRI(wikiId.JoinPath("followers"))
	c.SetActivityStreamsCc(cc)

	if summary != "" {
		summaryProp := streams.NewActivityStreamsSummaryProperty()
		summaryProp.AppendXMLSchemaString(summary)
		c.SetActivityStreamsSummary(summaryProp)
	}

	return
}
