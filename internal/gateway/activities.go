package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/diff"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

var ErrNotExists = errors.New("not found")

func (g *FedGatewayImpl) ProcessObject(ctx context.Context, asType vocab.Type) error {
	b, _ := streams.Serialize(asType)
	fmt.Printf("%+v\n", b)
	switch asType.GetTypeName() {
	case streams.ActivityStreamsArticleName:
		article, err := conversions.ConvertArticle(asType)
		if err != nil {
			return fmt.Errorf("invalid AS type")
		}

		raw, err := conversions.SerializeActivity(asType)
		if err != nil {
			return fmt.Errorf("serialization error: %w", err)
		}

		return g.db.PersistRemoteArticle(ctx, article, raw)
	case streams.ActivityStreamsFollowName:
		follow, ok := asType.(vocab.ActivityStreamsFollow)
		if !ok {
			return fmt.Errorf("failed to convert follow activity")
		}
		return g.processFollow(ctx, follow)
	case streams.ActivityStreamsUpdateName:
		update, ok := asType.(vocab.ActivityStreamsUpdate)
		if !ok {
			return fmt.Errorf("failed to convert update activity")
		}
		return g.processUpdate(ctx, update)
	default:
		return fmt.Errorf("%w: %s", errors.ErrUnsupported, asType.GetTypeName())
	}
}

func (g *FedGatewayImpl) ProcessOutbox(ctx context.Context, asType vocab.Type) error {
	switch asType.GetTypeName() {
	case streams.ActivityStreamsUpdateName:
		update, ok := asType.(vocab.ActivityStreamsUpdate)
		if !ok {
			return federation.ErrUnsupported
		}
		return g.processUpdateOutbox(ctx, update)
	case streams.ActivityStreamsCreateName:
		create, ok := asType.(vocab.ActivityStreamsCreate)
		if !ok {
			return federation.ErrUnsupported
		}
		return g.processCreateOutbox(ctx, create)
	default:
		return fmt.Errorf("%w: %s activity", federation.ErrUnsupported, asType.GetTypeName())
	}
}

func (g *FedGatewayImpl) processCreateOutbox(ctx context.Context, create vocab.ActivityStreamsCreate) error {
	actorIRI, err := g.processActor(create.GetActivityStreamsActor())
	if err != nil {
		return err
	}

	userId, err := g.db.GetUserIdByIRI(ctx, actorIRI)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return err
		}

		payload, err := streams.Serialize(create)
		if err != nil {
			return err
		}

		task := Task{
			Type: Fetch,
			To:   actorIRI.String(),
			Next: &Task{
				Type:    ProcessOutbox,
				Payload: payload,
			},
		}
		_, err = g.queue.Add(task).Save()
		return err
	}

	article, err := g.processArticleObject(ctx, create.GetActivityStreamsObject())
	if err != nil {
		return err
	}

	if article.Title == "" {
		return fmt.Errorf("%w: name", federation.ErrMissingProperty)
	}

	exists, err := g.db.ArticleTitleExists(ctx, article.Title)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: article \"%s\" already exists in this wiki", federation.ErrConflict, article.Title)
	}

	var summary string
	if summaryProp := create.GetActivityStreamsSummary(); summaryProp != nil && summaryProp.Len() != 0 {
		summary = summaryProp.Begin().GetXMLSchemaString()
	}

	articleIRI := g.cfg.Url.JoinPath("a", article.Title)
	articleInternal := domain.ArticleFed{
		ArticleCore: domain.ArticleCore{
			Title:     article.Title,
			Author:    g.cfg.Name,
			Host:      g.cfg.Domain,
			Content:   article.Content,
			MediaType: g.cfg.MediaType,
			Published: time.Now(),
		},
		ApID: articleIRI,
		To: []*url.URL{
			domain.Public,
			g.cfg.Url,
		},
		AttributedTo: g.cfg.Url,
		Url:          articleIRI,
	}

	diffs := diff.FindPatches("", article.Content)
	revision := domain.Revision{
		Summary:  summary,
		Diff:     diffs,
		Reviewed: false,
	}

	err = g.db.CreateLocalArticle(
		ctx,
		userId,
		articleInternal,
		revision,
	)
	if err != nil {
		return err
	}

	return g.CreateLocalArticle(ctx, articleInternal, actorIRI, summary)
}

type TransientArticle struct {
	Title   string
	IRI     *url.URL
	Content string
}

// processArticleObjects processes the object present in a create or update activity received at the wiki's outbox. It returns a
// transient article, containing the information needed to create or update an article.
func (g *FedGatewayImpl) processArticleObject(ctx context.Context, prop vocab.ActivityStreamsObjectProperty) (TransientArticle, error) {
	if prop == nil || prop.Len() == 0 {
		return TransientArticle{}, fmt.Errorf("%w: object property", federation.ErrMissingProperty)
	}

	obj := prop.Begin().GetType()
	if obj == nil {
		return TransientArticle{}, fmt.Errorf("%w: object property", federation.ErrMissingProperty)
	}

	if obj.GetTypeName() != streams.ActivityStreamsArticleName {
		return TransientArticle{}, fmt.Errorf("%w: %s", federation.ErrUnsupported, obj.GetTypeName())
	}

	article, ok := obj.(vocab.ActivityStreamsArticle)
	if !ok {
		return TransientArticle{}, federation.ErrUnprocessablePropValue
	}

	transient := TransientArticle{}
	id, err := g.processId(article.GetJSONLDId())
	if err == nil {
		transient.IRI = id
	}

	contentProp := article.GetActivityStreamsContent()
	if contentProp == nil || contentProp.Len() == 0 {
		return TransientArticle{}, fmt.Errorf("%w: content", federation.ErrMissingProperty)
	}

	transient.Content = contentProp.Begin().GetXMLSchemaString()

	titleProp := article.GetActivityStreamsName()
	if titleProp != nil && titleProp.Len() != 0 {
		transient.Title = strings.TrimSpace(titleProp.Begin().GetXMLSchemaString())
	}

	return transient, nil
}

func (g *FedGatewayImpl) processUpdateOutbox(ctx context.Context, update vocab.ActivityStreamsUpdate) error {
	actorIRI, err := g.processActor(update.GetActivityStreamsActor())
	if err != nil {
		return err
	}

	exists, err := g.db.Exists(ctx, actorIRI)
	if err != nil {
		return err
	}

	if !exists {
		payload, err := streams.Serialize(update)
		if err != nil {
			return err
		}

		task := Task{
			Type: Fetch,
			To:   actorIRI.String(),
			Next: &Task{
				Type:    ProcessOutbox,
				Payload: payload,
			},
		}
		_, err = g.queue.Add(task).Save()
		return err
	}

	var summary string
	if summaryProp := update.GetActivityStreamsSummary(); summaryProp != nil && summaryProp.Len() != 0 {
		// TODO: handle RFDLangString
		summary = summaryProp.Begin().GetXMLSchemaString()
	}

	article, err := g.processArticleObject(ctx, update.GetActivityStreamsObject())
	if err != nil {
		return err
	}

	if article.IRI == nil {
		return fmt.Errorf("%w: article IRI", federation.ErrMissingProperty)
	}

	exists, err = g.db.Exists(ctx, article.IRI)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%w: %s", ErrNotExists, article.IRI)
	}

	updateIRI, err := g.db.UpdateFedArticle(ctx, article.IRI, nil, actorIRI, article.Content, summary)
	if err != nil {
		return fmt.Errorf("%w: failed to update article", err)
	}

	articleId, err := g.db.GetArticleIdByIRI(ctx, article.IRI)
	if err != nil {
		return err
	}

	return g.UpdateLocalArticle(ctx, updateIRI, g.cfg.Url, summary, articleId)
}

func (g *FedGatewayImpl) processUpdate(ctx context.Context, update vocab.ActivityStreamsUpdate) error {
	id, err := g.processId(update.GetJSONLDId())
	if err != nil {
		return err
	}

	actor, err := g.processActor(update.GetActivityStreamsActor())
	if err != nil {
		return err
	}

	exists, err := g.db.Exists(ctx, actor)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Printf("%s not present in db\n\n", actor)
		payload, err := streams.Serialize(update)
		if err != nil {
			return err
		}
		task := Task{
			Type: Fetch,
			To:   actor.String(),
			Next: &Task{
				Type:    Process,
				Payload: payload,
			},
		}
		_, err = g.queue.Add(task).Save()
		return err
	}

	// Handle multiple objects.
	objProp := update.GetActivityStreamsObject()
	if objProp == nil || objProp.Len() == 0 {
		return fmt.Errorf("%w: object", federation.ErrMissingProperty)
	}

	var summary string
	if summaryProp := update.GetActivityStreamsSummary(); summaryProp != nil && summaryProp.Len() != 0 {
		summary = summaryProp.Begin().GetXMLSchemaString()
	}

	obj := objProp.Begin()
	if obj.IsIRI() {
		contentProp := update.GetActivityStreamsContent()
		if contentProp == nil || contentProp.Len() == 0 {
			return g.Fetch(obj.GetIRI())
		}

		//TODO: handle case when property is not an XMLSchemaString
		content := contentProp.Begin().GetXMLSchemaString()

		_, err = g.db.UpdateFedArticle(ctx, obj.GetIRI(), id, actor, content, summary)
		return err
	}

	if t := obj.GetType(); t != nil {
		return g.ProcessObject(ctx, t)
	}

	return fmt.Errorf("%w: updated object", federation.ErrUnprocessablePropValue)
}

// processUpdatedObject handles the object of an update activity, which typically will be an article or note.
// Updating an article requires some extra logic that is not covered in the ProcessObject method, such as
// inserting a revision having the update's IRI.
func (g *FedGatewayImpl) processUpdatedObject(ctx context.Context, obj vocab.Type, actorIRI, updateIRI *url.URL, summary string) error {
	switch obj.GetTypeName() {
	case streams.ActivityStreamsArticleName:
		article, err := conversions.ConvertArticle(obj)
		if err != nil {
			return err
		}

		exists, err := g.db.Exists(ctx, article.ApID)
		if err != nil {
			return err
		}
		if !exists {
			raw, err := conversions.SerializeActivity(obj)
			if err != nil {
				return err
			}
			return g.db.PersistRemoteArticle(ctx, article, raw)
		}

		_, err = g.db.UpdateFedArticle(ctx, article.ApID, updateIRI, actorIRI, article.Content, summary)
		return err
	default:
		return fmt.Errorf("%w: %s", federation.ErrUnsupported, obj.GetTypeName())
	}
}

func (g *FedGatewayImpl) processFollow(ctx context.Context, follow vocab.ActivityStreamsFollow) error {
	id, err := g.processId(follow.GetJSONLDId())
	if err != nil {
		return err
	}

	actor, err := g.processActor(follow.GetActivityStreamsActor())
	if err != nil {
		return err
	}

	obj, err := g.processObjectProperty(ctx, follow.GetActivityStreamsObject())
	if err != nil {
		return err
	}

	props, err := streams.Serialize(follow)
	if err != nil {
		return err
	}

	rawJSON, err := json.Marshal(props)
	if err != nil {
		return err
	}

	fmt.Printf("Actor: %s\nObject: %s\n", actor, obj)
	returnedId, err := g.db.Follow(ctx, domain.Follow{
		IRI:           id,
		Follower:      actor,
		Followee:      obj,
		FollowerInbox: nil,
		Raw:           rawJSON,
	})
	if err != nil {
		return err
	}

	// TODO: Ibis and Mastodon repeat the follow activity in the accept's object property.
	acceptId := g.cfg.Url.JoinPath("accept", strconv.Itoa(int(returnedId)))
	accept := conversions.NewAccept(acceptId, obj, id)

	return g.Deliver(ctx, accept, actor, obj)
}
