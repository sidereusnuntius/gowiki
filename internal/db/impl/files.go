package impl

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (d *dbImpl) Save(ctx context.Context, file domain.File) (id int64, err error) {
	err = d.WithTx(func (tx *queries.Queries) error {
		id, err = d.queries.InsertFile(ctx, queries.InsertFileParams{
			Local: file.Local,
			Digest: file.Digest,
			Path: sql.NullString{
				String: file.Path,
				Valid: file.Path != "",
			},
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

func (d *dbImpl) FileExists(ctx context.Context, digest string) (exists bool, err error) {
	exists, err = d.queries.FileExists(ctx, digest)
	if err != nil {
		err = d.HandleError(err)
	}
	return
}

func (d *dbImpl) GetFile(ctx context.Context, digest string) (file domain.File, err error) {
	f, err := d.queries.GetFile(ctx, digest)
	if err != nil {
		err = d.HandleError(err)
		return
	}

	id, err := url.Parse(f.ApID)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse stored url: " + f.ApID)
		err = db.ErrInternal
		return
	}
	u, err := url.Parse(f.Url)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse store url: " + f.Url)
		err = db.ErrInternal
		return
	}

	file = domain.File{
		FileMetadata: domain.FileMetadata{
			Name: f.Name.String,
			Filename: f.Filename.String,
			Type: f.Type,
			MimeType: f.MimeType,
			SizeBytes: f.SizeBytes.Int64,
			Local: f.Local,
		},
		Digest: f.Digest,
		Path: f.Path.String,
		ApId: id,
		Url: u,
	}
	return
}
