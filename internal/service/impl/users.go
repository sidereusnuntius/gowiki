package core

import (
	"context"
	"database/sql"
	"strings"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (s *AppService) GetProfile(ctx context.Context, name, host string) (p domain.Profile, err error) {
	name = strings.TrimSpace(name)
	host = strings.TrimSpace(host)

	p, err = s.DB.GetProfile(ctx, name, sql.NullString{
		Valid: host != "",
		String: host,
	})
	return
}
