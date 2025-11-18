package db

import (
	"context"
	"database/sql"

	"github.com/sidereusnuntius/wiki/internal/db/queries"
)

func (d *DB) GetRevisionList(ctx context.Context, title string) ([]queries.GetRevisionListRow, error) {
	list, err := d.queries.GetRevisionList(ctx, title)
	return list, d.HandleError(err)
}

func (d *DB) UpdateArticle(ctx context.Context, prevId, articleId, userId int64, summary, newContent string) (err error) {
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
func (d *DB) GetLastRevisionID(ctx context.Context, title string) (int64, string, int64, error) {
	a, err := d.queries.GetArticleIDS(ctx, title)
	return a.ArticleID, a.ApID, a.RevID, d.HandleError(err)
}

// CreateArticle creates a new local article, also inserting the article's first revision.
func (d *DB) CreateLocalArticle(ctx context.Context, article queries.CreateArticleParams, initialEdit queries.EditArticleParams) (err error) {
	t, err := d.db.Begin()
	if err != nil {
		// TODO: handle error.
		return
	}

	defer func() {
		if err != nil {
			t.Rollback()
		} else {
			err = t.Commit()
		}
	}()

	tx := d.queries.WithTx(t)
	articleId, err := tx.CreateArticle(ctx, article)
	if err != nil {
		return
	}
	initialEdit.ArticleID = articleId

	_, err = tx.EditArticle(ctx, initialEdit)
	return
}

func (d *DB) GetLocalArticle(ctx context.Context, title string) (queries.GetLocalArticleByTitleRow, error) {
	article, err := d.queries.GetLocalArticleByTitle(ctx, title)
	return article, d.HandleError(err)
}
