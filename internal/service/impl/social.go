package core

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/service"
)

func (s *AppService) FollowRemote(ctx context.Context, followerIRI, followeeIRI *url.URL) error {
	exists, err := s.DB.Exists(ctx, followerIRI)
	if err != nil {
		return fmt.Errorf("error checking if IRI exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("IRI %s does not exist", followerIRI)
	}

	return s.gateway.FollowRemoteActor(ctx, followerIRI, followeeIRI)
}

func (s *AppService) Follows(ctx context.Context, actor, object *url.URL) (bool, error) {
	if actor == nil || object == nil {
		return false, fmt.Errorf("%w: nil IRI", service.ErrInvalidInput)
	}

	follows, err := s.DB.Follows(ctx, actor, object)
	return follows, err
}

func (s *AppService) GetActorIRI(ctx context.Context, name, host string) (*url.URL, error) {
	name = strings.TrimSpace(name)
	host = strings.TrimSpace(host)
	if name == "" {
		return nil, fmt.Errorf("%w: empty name", service.ErrInvalidInput)
	}
	if host == "" {
		return nil, fmt.Errorf("%w: empty host", service.ErrInvalidInput)
	}

	iri, err := s.DB.GetActorIRI(ctx, name, host)
	return iri, err
}

func (s *AppService) GetProfile(ctx context.Context, name, host string) (p domain.Profile, err error) {
	name = strings.TrimSpace(name)
	host = strings.TrimSpace(host)
	if host == "" {
		host = s.Config.Domain
	}

	p, err = s.DB.GetProfile(ctx, name, sql.NullString{
		Valid:  true,
		String: host,
	})
	return
}
