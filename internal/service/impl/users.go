package core

import (
	"context"
	"strings"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (s *AppService) GetUserProfile(ctx context.Context, username, domain string) (p domain.Profile, err error) {
	username = strings.TrimSpace(username)
	domain = strings.TrimSpace(domain)

	p, err = s.DB.GetProfile(ctx, username, domain)
	return
}
