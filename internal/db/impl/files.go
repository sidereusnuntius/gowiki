package impl

import (
	"context"
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (d *dbImpl) Save(ctx context.Context, file domain.File) (id int64, err error) {
	err = d.WithTx(func (tx *queries.Queries) error {
		id, err = d.queries.InsertFile(ctx, queries.InsertFileParams{
			Local: file.Local,
			Digest: file.Digest,
			Path: file.Path,
			ApID: file.ApId.String(),
			Name: sql.NullString{
				Valid: file.Name != "",
				String: file.Name,
			},
			Filename: sql.NullString{
				Valid: file.Filename != "",
				String: file.Filename,
			},
			Type: file.Type,
			MimeType: file.MimeType,
			SizeBytes: sql.NullInt64{
				Valid: file.SizeBytes != 0,
				Int64: file.SizeBytes,
			},
			UploadedBy: sql.NullInt64{
				Valid: file.Local,
				Int64: file.UploaderId,
			},
			Url: file.Url.String(),
		})

		return err
	})

	return id, err
}