package impl

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (d *dbImpl) GetRevisionList(ctx context.Context, title string) ([]domain.Revision, error) {
	list, err := d.queries.GetRevisionList(ctx, title)
	if err != nil {
		return nil, d.HandleError(err)
	}

	edits := make([]domain.Revision, 0, len(list))
	for _, r := range list {
		edits = append(edits, domain.Revision{
			ID:       r.ID,
			Reviewed: r.Reviewed,
			Title:    r.Title,
			Summary:  r.Summary.String,
			Username: r.Username.String,
			Created:  r.Created,
		})
	}

	return edits, nil
}

func (d *dbImpl) UpdateArticle(ctx context.Context, prevId, articleId, userId int64, summary, newContent string, apId *url.URL) (URI *url.URL, err error) {
	content, err := d.queries.GetArticleContent(ctx, articleId)
	if err != nil {
		return nil, d.HandleError(err)
	}

	err = d.WithTx(func(tx *queries.Queries) error {

		diff := d.getDiff(content, newContent)

		_, err = d.insertRevision(
			ctx,
			tx,
			articleId,
			sql.NullInt64{
				Valid: userId != 0,
				Int64: userId,
			},
			sql.NullInt64{
				Valid: prevId != 0,
				Int64: prevId,
			},
			sql.NullString{
				Valid: summary != "",
				String: summary,
			},
			diff,
			apId,
		)
		if err != nil {
			return fmt.Errorf("failed to insert revision: %w", err)
		}

		return tx.UpdateArticle(ctx, queries.UpdateArticleParams{
			Content: newContent,
			ID:      articleId,
		})
	})
	return URI, err
}

// GetArticleIds returns the article's ID, ActivityPub ID and the ID of its last revision, if the article exists.
func (d *dbImpl) GetLastRevisionID(ctx context.Context, title string) (int64, *url.URL, int64, error) {
	a, err := d.queries.GetArticleIDS(ctx, title)
	if err != nil {
		return 0, nil, 0, d.HandleError(err)
	}
	iri, err := url.Parse(a.ApID)
	return a.ArticleID, iri, a.RevID, d.HandleError(err)
}

// CreateArticle creates a new local article, also inserting the article's first revision.
func (d *dbImpl) CreateLocalArticle(ctx context.Context, userId int64, article domain.ArticleFed, initialEdit domain.Revision) (err error) {
	log.Debug().
		Str("title", article.Title).
		Str("iri", article.ApID.String()).
		Msg("creating article")
	return d.WithTx(func(tx *queries.Queries) error {
		articleId, err := insertArticleTx(tx, ctx, true, article, domain.FedObj{
			Iri: article.ApID,
			RawJSON: nil,
			ApType: "Article",
			Local: true,
		}, 0)
		if err != nil {
			return err
		}

		_, err = d.insertRevision(
			ctx,
			tx,
			articleId,
			sql.NullInt64{
				Valid: userId != 0,
				Int64: userId,
			},
			sql.NullInt64{},
			sql.NullString{
				Valid: initialEdit.Summary != "",
				String: initialEdit.Summary,
			},
			initialEdit.Diff,
			nil,
		)
		return err
	})
}

func (d *dbImpl) GetArticle(ctx context.Context, title string, host, author sql.NullString) (domain.ArticleFed, error) {
	if host.Valid {
		host.String = strings.ToLower(host.String)
	}
	a, err := d.queries.GetArticle(ctx, queries.GetArticleParams{
		Host: host,
		Username: author,
    	Title: strings.ToLower(title),
	})

	var u *url.URL
	if a.Url.Valid {
		u, err = url.Parse(a.Url.String)
		if err != nil {
			return domain.ArticleFed{}, fmt.Errorf("%w: invalid url property: %s", db.ErrInternal, a.Url.String)
		}
	}

	return domain.ArticleFed{
		ArticleCore: domain.ArticleCore{
			Title:     a.Title,
			Host: a.Host.String,
			Author: a.Author.String,
			Summary:   a.Summary.String,
			Content:   a.Content,
			Protected: a.Protected,
			MediaType: a.MediaType,
			License:   "", // TODO
			Language:  a.Language,
		},
		Url: u,
	}, d.HandleError(err)
}

func (d *dbImpl) GetArticleById(ctx context.Context, id int64) (domain.ArticleFed, error) {
	a, err := d.queries.GetArticleByID(ctx, id)
	if err != nil {
		return domain.ArticleFed{}, d.HandleError(err)
	}

	var attributedTo *url.URL
	if a.AttributedTo.Valid {
		attributedTo, _ = url.Parse(a.AttributedTo.String)
	}

	iri, err := url.Parse(a.ApID)
	if err != nil {
		return domain.ArticleFed{}, err
	}
	uri, _ := url.Parse(a.Url.String)

	return domain.ArticleFed{
		ApID:         iri,
		AttributedTo: attributedTo,
		To: []*url.URL{
			domain.Public,
			d.Config.Url,
		},
		Url: uri,
		ArticleCore: domain.ArticleCore{
			Title:       a.Title,
			Summary:     a.Summary.String,
			Content:     a.Content,
			Protected:   a.Protected,
			MediaType:   a.MediaType,
			License:     "", // TODO
			Language:    a.Language,
			Published:   time.Unix(a.Published.Int64, 0),
			LastUpdated: time.Unix(a.LastUpdated, 0),
		},
	}, err
}

func insertArticleTx(tx *queries.Queries, ctx context.Context, local bool, article domain.ArticleFed, raw domain.FedObj, fetched int64) (int64, error) {
	var url sql.NullString
	if article.Url != nil {
		url.Valid = true
		url.String = article.Url.String()
	}

	var attributedTo sql.NullString
	if article.AttributedTo != nil {
		attributedTo.Valid = true
		attributedTo.String = article.AttributedTo.String()
	}
	id, err := tx.CreateArticle(ctx, queries.CreateArticleParams{
		Local: local,
		ApID: article.ApID.String(),
		Author: attributedTo,
		AttributedTo: attributedTo,
		Url: url,
		Language: article.Language,
		MediaType: article.MediaType,
		Title: article.Title,
		Host: sql.NullString{
			Valid: true,
			String: article.Host,
		},
		Type: raw.ApType,
		Summary: sql.NullString{
			Valid: article.Summary != "",
			String: article.Summary,
		},
		Content: article.Content,
		Published: sql.NullInt64{
			Valid: !article.Published.IsZero(),
			Int64: article.Published.Unix(),
		},
		LastUpdated: article.LastUpdated.Unix(),
		LastFetched: sql.NullInt64{
			Valid: fetched != 0,
			Int64: fetched,
		},
	})
	if err != nil {
		return 0, err
	}

	err = tx.InsertApObject(ctx, queries.InsertApObjectParams{
		ApID: article.ApID.String(),
		LocalTable: sql.NullString{
			Valid: true,
			String: "articles",
		},
		LocalID: sql.NullInt64{
			Valid: true,
			Int64: id,
		},
		Type: raw.ApType,
		RawJson: raw.RawJSON,
		LastFetched: sql.NullInt64{
			Valid: fetched != 0,
			Int64: fetched,
		},
	})
	return id, err
}

func (d *dbImpl) insertRevision(ctx context.Context, tx *queries.Queries, articleId int64, userId, prevId sql.NullInt64, summary sql.NullString, diff string, apId *url.URL) (int64, error) {
	var URI sql.NullString
	if apId != nil {
		URI.Valid = true
		URI.String = apId.String()
	}
	id, err := tx.InsertRevision(ctx, queries.InsertRevisionParams{
		ApID: URI,
		ArticleID: articleId,
		UserID:    userId,
		Summary: summary,
		Diff: diff,
		Prev: prevId,
	})
	if err != nil || URI.Valid {
		return 0, err
	}


	URI.String = d.Config.Url.JoinPath("revisions", strconv.FormatInt(id, 10)).String()
	return id, tx.UpdateRevisionApId(ctx, queries.UpdateRevisionApIdParams{
		ApID: URI,
		ID: id,
	})
}

func (d *dbImpl) getDiff(prev, new string) string {
	diffs := d.DMP.DiffMain(prev, new, false)
	patches := d.DMP.PatchMake(diffs)
	return d.DMP.PatchToText(patches)
}