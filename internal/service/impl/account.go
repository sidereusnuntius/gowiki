package core

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/service"
	"github.com/sidereusnuntius/gowiki/internal/utils"
	"github.com/sidereusnuntius/gowiki/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

func (s *AppService) IsAdmin(ctx context.Context, accountId int64) (bool, error) {
	isAdmin, err := s.DB.IsAdmin(ctx, accountId)
	return isAdmin, err
}

// AuthenticateUser confirms the user's identity and, if their credentials are correct, returns data to be put
// in the login session, such as the user's name and id. user is either the user's Id or their
func (s *AppService) AuthenticateUser(ctx context.Context, user, password string) (u domain.Account, authenticated bool, err error) {
	user = strings.ToLower(strings.TrimSpace(user))

	err = validate.Email(user)
	if err == nil {
		u, err = s.DB.GetAuthDataByEmail(ctx, user)
	} else if err = validate.Username(user); err == nil {
		u, err = s.DB.GetAuthDataByUsername(ctx, user)
	} else {
		err = errors.New("invalid username or email")
	}

	err = errors.Join(err, validate.Password(password))
	if err != nil {
		err = fmt.Errorf("%w: %s", service.ErrInvalidInput, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	authenticated = err == nil
	return
}

func (s *AppService) CreateUser(ctx context.Context, username, password, email, reason string, admin bool, invitation string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	password = strings.ToLower(strings.TrimSpace(password))
	email = strings.ToLower(strings.TrimSpace(email))

	err := validate.SignUpForm(username, password, email, reason, s.Config.ApprovalRequired)
	if err != nil {
		return fmt.Errorf("%w: %s", service.ErrInvalidInput, err)
	}

	u, err := s.populateUser(username, "", "")
	if err != nil {
		return err
	}

	a, err := populateAccount(email, password, admin)
	if err != nil {
		return err
	}

	return s.DB.InsertUser(ctx, u, a, reason, invitation)
}

func populateAccount(email, password string, admin bool) (account domain.Account, err error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	account = domain.Account{
		Admin:    admin,
		Password: string(p),
		Email:    email,
	}
	return
}

func (s *AppService) populateUser(username, name, summary string) (user domain.UserFedInternal, err error) {
	apId := s.Config.Url.JoinPath("/u/" + username)
	_ = s.Config.Url.JoinPath("@" + username)

	pub, priv, err := utils.GenerateKeysPem(RsaKeySize)
	if err != nil {
		return
	}

	user = domain.UserFedInternal{
		UserFed: domain.UserFed{
			UserCore: domain.UserCore{
				Username: username,
				Name:     name,
				Host:     s.Config.Domain,
			},
			ApId:      apId,
			Inbox:     apId.JoinPath("/inbox"),
			Outbox:    apId.JoinPath("/outbox"),
			Followers: apId.JoinPath("/followers"),
			PublicKey: pub,
		},
		PrivateKey: priv,
		// If an invitation is required, then it is assumed that the people who sign up are trustworthy
		// and known to the instance's administrator.
		Trusted: s.Config.InvitationRequired,
	}

	return
}
