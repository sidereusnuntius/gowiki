package impl

import (
	"context"
	"database/sql"
	"errors"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)



func (d *dbImpl) GetProfile(ctx context.Context, name string, host sql.NullString) (p domain.Profile, err error) {
	userRow, err := d.queries.GetActorData(ctx, queries.GetActorDataParams{
		LOWER: name,
		Host: host,
	})
	if err != nil {
		return
	}

	var userURL *url.URL
	if userRow.Url.Valid {
		if userURL, err = url.Parse(userRow.Url.String); err != nil {
			return domain.Profile{}, err
		}
	}

	u := domain.UserCore{
		Username: userRow.Name.String,
		Host: userRow.Host.String,
		Summary: userRow.Summary.String,
		URL: userURL,
	}
	
	articleRows, err := d.queries.GetArticlesByActorId(ctx, sql.NullString{
		Valid: true,
		String: userRow.ApID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows){
		return domain.Profile{}, d.HandleError(err)
	}

	articles := make([]domain.ArticlePreview, len(articleRows))
	for i, r := range articleRows {
		articles[i] = domain.ArticlePreview{
			Title: r.Title,
		}
	}

	r, err := d.queries.GetRevisionsByUserId(ctx, sql.NullInt64{
		Valid: true,
		Int64: u.ID,
	})
	edits := make([]domain.Revision, len(r))
	for i, r := range r {
		edits[i] = domain.Revision{
			ID:       r.ID,
			Title:    r.Title,
			Reviewed: r.Reviewed,
			Summary:  r.Summary.String,
			Created:  r.Created,
		}
	}
	p = domain.Profile{
		UserCore: u,
		Articles: articles,
		Edits:    edits,
	}

	if err != nil {
		err = d.HandleError(err)
	}
	return
}
