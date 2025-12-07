package impl

import (
	"context"
	"crypto"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

const PageSize = 20

func (d *dbImpl) UpdateFedArticle(ctx context.Context, articleIRI, updateIRI, actorIRI *url.URL, newContent, summary string) (*url.URL, error) {
	article, err := d.queries.GetArticleContentByIRI(ctx, articleIRI.String())
	if err != nil {
		return nil, fmt.Errorf("%w: article with IRI %s", d.HandleError(err), articleIRI)
	}
	diff := d.getDiff(article.Content, newContent)

	userId, err := d.queries.GetUserId(ctx, actorIRI.String())
	if err != nil {
		return nil, fmt.Errorf("%w: user with IRI %s", d.HandleError(err), actorIRI)
	}

	err = d.WithTx(func(tx *queries.Queries) error {
		err = tx.UpdateArticleByIRI(ctx, queries.UpdateArticleByIRIParams{
			Content: newContent,
			ApID:    articleIRI.String(),
		})
		if err != nil {
			return err
		}

		var revisionId int64
		revisionId, updateIRI, err = d.insertRevision(
			ctx,
			tx,
			article.ID,
			sql.NullInt64{
				Valid: true,
				Int64: userId,
			},
			sql.NullInt64{},
			sql.NullString{
				Valid:  summary != "",
				String: summary,
			},
			diff,
			updateIRI,
		)
		if err != nil {
			return fmt.Errorf("failed to insert revision: %w", err)
		}

		fmt.Printf("%s\n", updateIRI)
		return tx.InsertApObject(ctx, queries.InsertApObjectParams{
			ApID: updateIRI.String(),
			LocalTable: sql.NullString{
				Valid:  true,
				String: "revisions",
			},
			LocalID: sql.NullInt64{
				Valid: true,
				Int64: revisionId,
			},
			Type: "Update",
		})
	})
	return updateIRI, err
}

func (d *dbImpl) PersistRemoteArticle(ctx context.Context, article domain.ArticleFed, articleRaw domain.FedObj) error {
	author := article.AttributedTo.String()
	userId, err := d.queries.GetUserId(ctx, author)
	if err != nil {
		return fmt.Errorf("could not query internal id for IRI %s: %w", author, d.HandleError(err))
	}
	return d.WithTx(func(tx *queries.Queries) error {
		fetched := time.Now().Unix()
		id, err := insertArticleTx(tx, ctx, false, article, articleRaw, fetched)
		if err != nil {
			return err
		}

		diff := d.getDiff("", article.Content)
		_, _, err = d.insertRevision(
			ctx,
			tx,
			id,
			sql.NullInt64{
				Valid: true,
				Int64: userId,
			},
			sql.NullInt64{},
			//TODO: get summary
			sql.NullString{},
			diff,
			nil,
		)
		return err
	})
}

func (d *dbImpl) GetCollectionStart(ctx context.Context, collectionIRI *url.URL) (size, start int64, err error) {
	result, err := d.queries.GetCollectionStart(ctx, collectionIRI.String())
	if err != nil {
		err = d.HandleError(err)
		return
	}
	return result.Size, result.Start + 1, nil
}

func (d *dbImpl) GetCollectionActivities(ctx context.Context, collectionIRI *url.URL, last int64) (activities []map[string]any, err error) {
	// Maybe verify first if the collection exists?
	results, err := d.queries.GetCollectionActivitiesPage(ctx, queries.GetCollectionActivitiesPageParams{
		CollectionID: collectionIRI.String(),
		LastID: sql.NullInt64{
			Valid: last != 0,
			Int64: last,
		},
		PageSize: PageSize,
	})

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			activities = make([]map[string]any, 0)
			err = nil
			return
		}
		log.Error().Err(err).Str("collection IRI", collectionIRI.String()).Msg("error querying collection's activities")
		err = d.HandleError(err)
		return
	}

	activities = make([]map[string]any, len(results))
	for i, r := range results {
		err = json.Unmarshal(r.RawJson, &activities[i])
		if err != nil {
			log.Error().Err(err).
				Str("activity IRI", r.ApID).
				Bytes("raw JSON", r.RawJson).
				Msg("error unmarshaling activity into intermediate map")
			return
		}
	}
	return activities, nil
}

func (d *dbImpl) GetCollectionMemberIRIS(ctx context.Context, collectionIRI *url.URL) ([]*url.URL, error) {
	result, err := d.queries.CollectionMembersIRIs(ctx, collectionIRI.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*url.URL{}, nil
		}
		return nil, d.HandleError(err)
	}

	uris := make([]*url.URL, len(result))
	for i, u := range result {
		uris[i], err = url.Parse(u)
		if err != nil {
			return nil, err
		}
	}
	return uris, nil
}

func (d *dbImpl) AddOutbox(ctx context.Context, apType string, raw []byte, id, outbox *url.URL) error {
	exists, err := d.Exists(ctx, id)
	if err != nil {
		return d.HandleError(err)
	}
	idStr := id.String()
	return d.WithTx(func(tx *queries.Queries) error {
		if !exists {
			err := tx.InsertApObject(ctx, queries.InsertApObjectParams{
				ApID:    idStr,
				Type:    apType,
				RawJson: raw,
			})
			if err != nil {
				return err
			}
		}

		_, err = tx.AddToCollection(ctx, queries.AddToCollectionParams{
			CollectionApID: outbox.String(),
			MemberApID:     idStr,
		})
		return err
	})
}

func (d *dbImpl) GetFollowing(ctx context.Context, actorIRI *url.URL) ([]*url.URL, error) {
	result, err := d.queries.GetFollowing(ctx, actorIRI.String())
	var following []*url.URL
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*url.URL{}, nil
		}
		return nil, d.HandleError(err)
	}

	following = make([]*url.URL, len(result))
	for i, f := range result {
		following[i], err = url.Parse(f)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to parse IRI: %s", db.ErrInternal, f)
		}
	}
	return following, nil
}

func (d *dbImpl) GetFollowers(ctx context.Context, id *url.URL) ([]*url.URL, error) {
	followers, err := d.queries.GetFollowers(ctx, id.String())
	if err != nil {
		d.HandleError(err)
	}

	followersIRIS := make([]*url.URL, 0, len(followers))
	var u *url.URL
	for _, f := range followers {
		u, err = url.Parse(f)
		if err != nil {
			return nil, err
		}
		followersIRIS = append(followersIRIS, u)
	}

	return followersIRIS, nil
}

func (d *dbImpl) GetCollectionPage(ctx context.Context, iri *url.URL, last int64) (ids []*url.URL, err error) {
	var members []string

	if last == 0 {
		members, err = d.queries.GetCollectionFirstPage(ctx, iri.String())
	} else {
		// TODO: query with pagination.
	}

	if err != nil {
		err = d.HandleError(err)
		return
	}

	var temp *url.URL
	ids = make([]*url.URL, 0, len(members))
	for _, id := range members {
		temp, err = url.Parse(id)
		if err != nil {
			return
		}

		ids = append(ids, temp)
	}
	return
}

func (d *dbImpl) CollectionContains(ctx context.Context, collection, id *url.URL) (bool, error) {
	exists, err := d.queries.CollectionContains(ctx, queries.CollectionContainsParams{
		CollectionApID: collection.String(),
		MemberApID:     id.String(),
	})
	if err != nil {
		err = d.HandleError(err)
	}
	return exists != 0, err
}

func (d *dbImpl) DeleteAp(ctx context.Context, id *url.URL) error {
	err := d.queries.DeleteAp(ctx, id.String())
	return d.HandleError(err)
}

func (d *dbImpl) UpdateAp(ctx context.Context, id *url.URL, rawJSON []byte) error {
	err := d.queries.UpdateAp(ctx, queries.UpdateApParams{
		RawJson: rawJSON,
		ApID:    id.String(),
	})
	return d.HandleError(err)
}

func (d *dbImpl) Exists(ctx context.Context, id *url.URL) (bool, error) {
	exists, err := d.queries.ApExists(ctx, id.String())
	if err != nil {
		err = d.HandleError(err)
	}

	return exists != 0, err
}

func (d *dbImpl) ActorIdByOutbox(ctx context.Context, iri *url.URL) (*url.URL, error) {
	id, err := d.queries.ActorIdByOutbox(ctx, iri.String())

	if err != nil {
		return nil, d.HandleError(err)
	}
	iri, err = url.Parse(id)

	return iri, d.HandleError(err)
}

func (d *dbImpl) ActorIdByInbox(ctx context.Context, iri *url.URL) (*url.URL, error) {
	id, err := d.queries.ActorIdByInbox(ctx, iri.String())

	if err != nil {
		return nil, d.HandleError(err)
	}
	iri, err = url.Parse(id)

	return iri, d.HandleError(err)
}

func (d *dbImpl) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (*url.URL, error) {
	id, err := d.queries.OutboxForInbox(ctx, inboxIRI.String())
	if err != nil {
		return nil, d.HandleError(err)
	}
	outboxIRI, err := url.Parse(id)
	return outboxIRI, d.HandleError(err)
}

func (d *dbImpl) GetUserFed(ctx context.Context, id *url.URL) (user domain.UserFed, err error) {
	defer func() {
		if err != nil {
			err = d.HandleError(err)
		}
	}()

	u, err := d.queries.GetUserFull(ctx, id.String())
	if err != nil {
		return
	}

	apId, err := url.Parse(u.ApID)
	if err != nil {
		return
	}

	inbox, err := url.Parse(u.Inbox)
	if err != nil {
		return
	}

	outbox, err := url.Parse(u.Outbox)
	if err != nil {
		return
	}

	followers, err := url.Parse(u.Followers)
	if err != nil {
		return
	}

	user = domain.UserFed{
		UserCore: domain.UserCore{
			Username: u.Username.String,
			Name:     u.Name.String,
			Summary:  u.Summary.String,
			//URL: , TODO
		},
		ApId:        apId,
		Inbox:       inbox,
		Outbox:      outbox,
		Followers:   followers,
		PublicKey:   u.PublicKey,
		Created:     time.Unix(u.Created, 0),
		LastUpdated: time.Unix(u.LastUpdated, 0),
	}
	return
}

func (d *dbImpl) GetApObject(ctx context.Context, iri *url.URL) (domain.FedObj, error) {
	log.Debug().Str("iri", iri.String()).Msg("querying ap cache table")
	obj, err := d.queries.GetApObject(ctx, iri.String())
	if err != nil {
		log.Error().Err(err).Msg("at GetApObject")
		err = d.HandleError(err)
	}

	return domain.FedObj{
		Iri:        iri,
		RawJSON:    obj.RawJson,
		ApType:     obj.Type,
		Local:      !obj.LastFetched.Valid,
		LocalTable: obj.LocalTable.String,
		LocalId:    obj.LocalID.Int64,
	}, err
}

func (d *dbImpl) CreateApObject(ctx context.Context, obj domain.FedObj, fetched int64) error {
	err := d.queries.InsertApObject(ctx, queries.InsertApObjectParams{
		ApID: obj.Iri.String(),
		LocalTable: sql.NullString{
			Valid:  obj.LocalTable != "",
			String: obj.LocalTable,
		},
		LocalID: sql.NullInt64{
			Valid: obj.LocalTable != "" && obj.LocalId != 0,
			Int64: obj.LocalId,
		},
		Type:    obj.ApType,
		RawJson: obj.RawJSON,
		LastFetched: sql.NullInt64{
			Valid: !obj.Local,
			Int64: fetched,
		},
	})

	if err != nil {
		err = d.HandleError(err)
	}
	return err
}

func (d *dbImpl) GetUserByID(ctx context.Context, id int64) (user domain.UserFed, err error) {
	log.Debug().Int64("id", id).Msg("GetUserByID")
	u, err := d.queries.GetUserFullByID(ctx, id)

	if err != nil {
		log.Error().Err(err).Msg("at GetUserByID")
		err = d.HandleError(err)
		return
	}

	apId, err := url.Parse(u.ApID)
	if err != nil {
		err = db.ErrInternal
		return
	}

	inbox, err := url.Parse(u.Inbox)
	if err != nil {
		err = db.ErrInternal
		return
	}

	outbox, err := url.Parse(u.Outbox)
	if err != nil {
		err = db.ErrInternal
		return
	}

	followers, err := url.Parse(u.Followers)
	if err != nil {
		err = db.ErrInternal
		return
	}

	return domain.UserFed{
		UserCore: domain.UserCore{
			Username: u.Username.String,
			Name:     u.Name.String,
			Summary:  u.Summary.String,
			//URL: , TODO
		},
		ApId:        apId,
		Inbox:       inbox,
		Outbox:      outbox,
		Followers:   followers,
		PublicKey:   u.PublicKey,
		Created:     time.Unix(u.Created, 0),
		LastUpdated: time.Unix(u.LastUpdated, 0),
	}, err
}

func (d *dbImpl) Follow(ctx context.Context, follow domain.Follow) (int64, *url.URL, error) {
	var inbox string
	if follow.FollowerInbox != nil {
		inbox = follow.FollowerInbox.String()
	}

	var followIRI sql.NullString
	if follow.IRI != nil {
		followIRI.Valid = true
		followIRI.String = follow.IRI.String()
	}

	var followID int64
	err := d.WithTx(func(tx *queries.Queries) error {
		var err error
		followID, err = tx.Follow(ctx, queries.FollowParams{
			FollowApID:   followIRI,
			FollowerApID: follow.Follower.String(),
			FolloweeApID: follow.Followee.String(),
			FollowerInboxUrl: sql.NullString{
				Valid:  follow.FollowerInbox != nil,
				String: inbox,
			},
		})
		if err != nil {
			return err
		}

		if !followIRI.Valid {
			follow.IRI = d.Config.Url.JoinPath("follows", strconv.FormatInt(followID, 10))
			followIRI.String = follow.IRI.String()
			followIRI.Valid = true
			err = tx.UpdateFollowApId(ctx, queries.UpdateFollowApIdParams{
				FollowApID: followIRI,
				ID:         followID,
			})
			if err != nil {
				return err
			}
		}

		if follow.Followee.Hostname() == d.Config.Url.Hostname() {
			_, err = tx.AddToCollection(ctx, queries.AddToCollectionParams{
				CollectionApID: follow.Followee.JoinPath("followers").String(),
				MemberApID:     follow.IRI.String(),
			})
			if err != nil {
				return err
			}
		}

		return tx.InsertApObject(ctx, queries.InsertApObjectParams{
			ApID: followIRI.String,
			LocalTable: sql.NullString{
				Valid:  true,
				String: "follows",
			},
			LocalID: sql.NullInt64{
				Valid: true,
				Int64: followID,
			},
			Type:    "Follow",
			RawJson: follow.Raw,
		})
	})

	return followID, follow.IRI, err
}

func (d *dbImpl) GetFollow(ctx context.Context, followIRI *url.URL) (domain.Follow, error) {
	f, err := d.queries.GetFollow(ctx, sql.NullString{
		Valid:  true,
		String: followIRI.String(),
	})
	if err != nil {
		return domain.Follow{}, d.HandleError(err)
	}

	follower, err := url.Parse(f.FollowerApID)
	if err != nil {
		return domain.Follow{}, fmt.Errorf("%w: failed to parse follower's IRI: %s", db.ErrInternal, f.FollowerApID)
	}

	followee, err := url.Parse(f.FolloweeApID)
	if err != nil {
		return domain.Follow{}, fmt.Errorf("%w: failed to parse followee's IRI: %s", db.ErrInternal, f.FolloweeApID)
	}

	return domain.Follow{
		IRI:      followIRI,
		Follower: follower,
		Followee: followee,
	}, nil
}

func (d *dbImpl) Follows(ctx context.Context, actor, object *url.URL) (bool, error) {
	result, err := d.queries.Follows(ctx, queries.FollowsParams{
		FollowerApID: actor.String(),
		FolloweeApID: object.String(),
	})
	if err != nil {
		err = d.HandleError(err)
	}

	return result != 0, err
}

func (d *dbImpl) ApproveFollow(ctx context.Context, followIRI *url.URL, accept domain.FedObj) error {
	return d.WithTx(func(tx *queries.Queries) error {
		err := tx.AcceptFollow(ctx, sql.NullString{
			Valid:  true,
			String: followIRI.String(),
		})
		if err != nil {
			return err
		}

		now := sql.NullInt64{
			Valid: true,
			Int64: time.Now().Unix(),
		}
		return tx.InsertApObject(ctx, queries.InsertApObjectParams{
			ApID:        accept.Iri.String(),
			Type:        accept.ApType,
			RawJson:     accept.RawJSON,
			LastUpdated: now,
			LastFetched: now,
		})
	})
}

func (d *dbImpl) GetUserPrivateKey(ctx context.Context, id int64) (owner *url.URL, key crypto.PrivateKey, err error) {
	res, err := d.queries.GetUserKeys(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch actor's private key")
		err = d.HandleError(err)
		return
	}

	block, _ := pem.Decode([]byte(res.PrivateKey))
	if block == nil || block.Type != "PRIVATE KEY" {
		err = errors.New("failure to parse private key")
		return
	}

	key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse private key")
		return
	}

	owner, err = url.Parse(res.ApID)
	if err != nil {
		log.Error().Err(err).Msg("parse error")
	}
	return
}

func (d *dbImpl) GetPublicKeyByActorIRI(ctx context.Context, IRI *url.URL) (string, error) {
	pubKey, err := d.queries.GetPublicKey(ctx, IRI.String())
	if err != nil {
		err = d.HandleError(err)
	}
	return pubKey, err
}

func (d *dbImpl) GetUserPrivateKeyByURI(ctx context.Context, url *url.URL) (key crypto.PrivateKey, err error) {
	log.Debug().Str("id", url.String()).Send()
	k, err := d.queries.GetPrivateKeyByID(ctx, url.String())

	if err != nil {
		err = d.HandleError(err)
		return
	}

	block, _ := pem.Decode([]byte(k))
	if block == nil || block.Type != "PRIVATE KEY" {
		err = errors.New("failure to parse private key")
		return
	}

	key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	return
}

func (d *dbImpl) InsertOrUpdateUser(ctx context.Context, u domain.UserFed, fetched time.Time) error {
	return d.WithTx(func(tx *queries.Queries) error {
		id, err := tx.InsertOrUpdateUser(ctx, queries.InsertOrUpdateUserParams{
			ApID: u.ApId.String(),
			Url: sql.NullString{
				Valid:  u.URL != nil,
				String: u.URL.String(),
			},
			Username: sql.NullString{
				Valid:  u.Username != "",
				String: u.Username,
			},
			Name: sql.NullString{
				Valid:  u.Name != "",
				String: u.Name,
			},
			Host: sql.NullString{
				Valid:  true,
				String: u.ApId.Host,
			},
			Summary: sql.NullString{
				Valid:  u.Summary != "",
				String: u.Summary,
			},
			Inbox:     u.Inbox.String(),
			Outbox:    u.Outbox.String(),
			Followers: u.Followers.String(),
			PublicKey: u.PublicKey,
			LastFetched: sql.NullInt64{
				Valid: true,
				Int64: fetched.Unix(),
			},
		})

		if err != nil {
			return err
		}

		err = tx.InsertOrUpdateApObject(ctx, queries.InsertOrUpdateApObjectParams{
			ApID: u.ApId.String(),
			LocalTable: sql.NullString{
				Valid:  true,
				String: "users",
			},
			LocalID: sql.NullInt64{
				Valid: true,
				Int64: id,
			},
			Type:    "Person",
			RawJson: nil,
			LastFetched: sql.NullInt64{
				Valid: true,
				Int64: fetched.Unix(),
			},
		})
		if err != nil {
			return err
		}

		return tx.UpdateFollowInbox(ctx, queries.UpdateFollowInboxParams{
			FollowerInboxUrl: sql.NullString{
				Valid:  u.Inbox != nil,
				String: u.Inbox.String(),
			},
			FollowerApID: u.ApId.String(),
		})
	})
}

func (d *dbImpl) GetActorIRI(ctx context.Context, name, host string) (*url.URL, error) {
	iriStr, err := d.queries.GetActorIRI(ctx, queries.GetActorIRIParams{
		LOWER: name,
		Host: sql.NullString{
			Valid:  true,
			String: host,
		},
	})
	if err != nil {
		return nil, d.HandleError(err)
	}

	iri, err := url.Parse(iriStr)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to parse IRI %s", db.ErrInternal, iriStr)
	}
	return iri, nil
}

func (d *dbImpl) GetActorInbox(ctx context.Context, actor *url.URL) (*url.URL, error) {
	inbox, err := d.queries.GetInboxByActorId(ctx, actor.String())
	if err != nil {
		return nil, d.HandleError(err)
	}
	if inbox == "" {
		return nil, fmt.Errorf("%w: empty inbox", db.ErrNotFound)
	}

	iri, err := url.Parse(inbox)
	if err != nil {
		err = db.ErrInternal
	}
	return iri, err
}

func (d *dbImpl) InsertOrUpdateCollective(ctx context.Context, collective domain.Collective, fetched time.Time) error {
	return d.WithTx(func(tx *queries.Queries) error {
		id, err := tx.InsertOrUpdateCollective(ctx, queries.InsertOrUpdateCollectiveParams{
			Name: sql.NullString{
				Valid: collective.Name != "",
				String: collective.Name,
			},
    		Host: collective.Hostname,
    		Url: sql.NullString{
				Valid: collective.Url != nil,
				String: collective.Url.String(),
			},
    		Summary: sql.NullString{
				Valid: collective.Summary != "",
				String: collective.Summary,
			},
    		PublicKey: sql.NullString{
				Valid: collective.PublicKey != "",
				String: collective.PublicKey,
			},
    		Inbox: sql.NullString{
				Valid: collective.Inbox != nil,
				String: collective.Inbox.String(),
			},
    		Outbox: sql.NullString{
				Valid: collective.Outbox != nil,
				String: collective.Outbox.String(),
			},
    		Followers: sql.NullString{
				Valid: collective.Followers != nil,
				String: collective.Followers.String(),
			},
		})
		if err != nil {
			return err
		}

		return tx.InsertOrUpdateApObject(ctx, queries.InsertOrUpdateApObjectParams{
			ApID: collective.Url.String(),
			LocalTable: sql.NullString{
				Valid: true,
				String: "collectives",
			},
			LocalID: sql.NullInt64{
				Valid: true,
				Int64: id,
			},
			Type: collective.Type,
			RawJson: nil,
			LastFetched: sql.NullInt64{
				Valid: !fetched.IsZero(),
				Int64: fetched.Unix(),
			},
		})
	})
}

func (d *dbImpl) GetCollectiveById(ctx context.Context, id int64) (c domain.Collective, err error) {
	obj, err := d.queries.GetCollectiveByID(ctx, id)
	if err != nil {
		return domain.Collective{}, err
	}
	c = domain.Collective{
		Type:       obj.Type,
		Name:       obj.Name.String,
		Hostname:   obj.Host,
		PublicKey: obj.PublicKey.String,
	}

	c.Inbox, err = url.Parse(obj.Inbox.String)
	if err != nil {
		return
	}

	if obj.Outbox.Valid {
		c.Outbox, err = url.Parse(obj.Outbox.String)
		if err != nil {
			return
		}
	}

	if obj.Followers.Valid {
		c.Followers, err = url.Parse(obj.Followers.String)
		if err != nil {
			return
		}
	}

	if obj.Url.Valid {
		c.Url, err = url.Parse(obj.Url.String)
	}

	return
}

func (d *dbImpl) GetUserApId(ctx context.Context, username string) (*url.URL, error) {
	log.Debug().Str("username", username).Send()
	id, err := d.queries.GetUserApId(ctx, username)
	if err != nil {
		return nil, d.HandleError(err)
	}

	uri, err := url.Parse(id)
	return uri, err
}
