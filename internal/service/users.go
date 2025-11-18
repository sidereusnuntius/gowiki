package service

import (
	"context"
	"strings"

	"github.com/sidereusnuntius/wiki/internal/domain"
)

func (s *Service) GetUserProfile(ctx context.Context, username, domain string) (p domain.Profile, err error) {
	username = strings.TrimSpace(username)
	domain = strings.TrimSpace(domain)

	p, err = s.DB.GetProfile(ctx, username, domain)
	return
}
