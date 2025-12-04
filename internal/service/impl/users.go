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
	if host == "" {
		host = s.Config.Domain
	}

	p, err = s.DB.GetProfile(ctx, name, sql.NullString{
		Valid: true,
		String: host,
	})
	return
}
