package impl

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/rs/zerolog/log"
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
			Username: r.Username,
			Created:  r.Created,
		})
	}

	return edits, nil
}

func (d *dbImpl) UpdateArticle(ctx context.Context, prevId, articleId, userId int64, summary, newContent string) (err error) {
	content, err := d.queries.GetArticleContent(ctx, articleId)
	if err != nil {
		return d.HandleError(err)
	}

	t, err := d.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		// Handle error.
		if err != nil {
			t.Rollback()
		} else {
			err = t.Commit()
		}
	}()
	tx := d.queries.WithTx(t)

	diffs := d.DMP.DiffMain(content, newContent, false)
	patches := d.DMP.PatchMake(diffs)

	err = tx.InsertRevision(ctx, queries.InsertRevisionParams{
		//TODO: generate apId. Perhaps use the generated id?
		//ApID: ,
		ArticleID: articleId,
		UserID:    userId,
		Summary: sql.NullString{
			String: summary,
			Valid:  summary != "",
		},
		Diff: d.DMP.PatchToText(patches),
		Prev: sql.NullInt64{
			Int64: prevId,
			Valid: true,
		},
	})
	if err != nil {
		return
	}

	err = tx.UpdateArticle(ctx, queries.UpdateArticleParams{
		Content: newContent,
		ID:      articleId,
	})
	return
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
		var apid string
		if article.ApID != nil {
			apid = article.ApID.String()
		}
		articleId, err := tx.CreateArticle(ctx, queries.CreateArticleParams{
			ApID: article.ApID.String(),
			Url: sql.NullString{
				Valid:  article.Url != nil,
				String: apid,
			},
			InstanceID: sql.NullInt64{},
			Language:   article.Language,
			MediaType:  article.MediaType,
			Title:      article.Title,
			Content:    article.Content,
		})
		if err != nil {
			return err
		}

		err = tx.InsertApObject(ctx, queries.InsertApObjectParams{
			ApID: apid,
			LocalTable: sql.NullString{
				Valid: true,
				String: "articles",
			},
			LocalID: sql.NullInt64{
				Valid: true,
				Int64: articleId,
			},
			Type: "Article",
		})
		if err != nil {
			return err
		}

		_, err = tx.EditArticle(ctx, queries.EditArticleParams{
			ApID: sql.NullString{
				// TODO
			},
			ArticleID: articleId,
			UserID:    userId,
			Summary: sql.NullString{
				Valid:  initialEdit.Summary != "",
				String: initialEdit.Summary,
			},
			Diff: initialEdit.Diff,
		})
		return err
	})
}

func (d *dbImpl) GetLocalArticle(ctx context.Context, title string) (domain.ArticleCore, error) {
	a, err := d.queries.GetLocalArticleByTitle(ctx, title)
	return domain.ArticleCore{
		Title:     a.Title,
		Summary:   a.Summary.String,
		Content:   a.Content,
		Protected: a.Protected,
		MediaType: a.MediaType,
		License:   "", // TODO
		Language:  a.Language,
	}, d.HandleError(err)
}
