package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/wiki/internal/config"
	"github.com/sidereusnuntius/wiki/internal/db"
	"github.com/sidereusnuntius/wiki/internal/db/queries"
	"github.com/sidereusnuntius/wiki/internal/state"
	"github.com/sidereusnuntius/wiki/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidInput = errors.New("invalid")

const (
	RsaKeySize = 2048
	BcryptCost = 10
)

type Service struct {
	Config config.Configuration
	DB     *db.DB
	DMP    *diffmatchpatch.DiffMatchPatch
}

func New(state state.State) Service {
	dmp := diffmatchpatch.New()
	DB := db.New(state)
	return Service{
		Config: state.Config,
		DB:     &DB,
		DMP:    dmp,
	}
}

// AuthenticateUser confirms the user's identity and, if their credentials are correct, returns data to be put
// in the login session, such as the user's name and id. user is either the user's Id or their
func (s *Service) AuthenticateUser(ctx context.Context, user, password string) (u db.UserData, authenticated bool, err error) {
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
		err = fmt.Errorf("%w: %s", ErrInvalidInput, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	authenticated = err == nil
	return
}

func (s *Service) CreateUser(ctx context.Context, username, password, email, reason string, admin bool, invitation string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	password = strings.ToLower(strings.TrimSpace(password))
	email = strings.ToLower(strings.TrimSpace(email))

	err := validate.SignUpForm(username, password, email, reason, s.Config.ApprovalRequired)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidInput, err)
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

func populateAccount(email, password string, admin bool) (account queries.CreateAccountParams, err error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	account = queries.CreateAccountParams{
		Admin:    admin,
		Password: string(p),
		Email:    email,
	}
	return
}

func (s *Service) populateUser(username, name, summary string) (user queries.CreateLocalUserParams, err error) {
	apId := s.Config.Url.JoinPath("/u/" + username)

	key, err := rsa.GenerateKey(rand.Reader, RsaKeySize)
	if err != nil {
		return
	}

	priv, err := privateKeyPem(key)
	if err != nil {
		return
	}

	pub, err := publicKeyPem(&key.PublicKey)
	if err != nil {
		return
	}

	user = queries.CreateLocalUserParams{
		ApID:       apId.String(),
		Username:   username,
		Name:       name,
		Inbox:      apId.JoinPath("/inbox").String(),
		Outbox:     apId.JoinPath("/outbox").String(),
		Followers:  apId.JoinPath("/followers").String(),
		PublicKey:  pub,
		PrivateKey: priv,
		// If an invitation is required, then it is assumed that the people who sign up are trustworthy
		// and known to the instance's administrator.
		Trusted: s.Config.InvitationRequired,
	}

	return
}

func privateKeyPem(key *rsa.PrivateKey) (string, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})), nil
}

func publicKeyPem(key *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})), err
}
