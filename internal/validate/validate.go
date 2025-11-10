package validate

import (
	"errors"
	"fmt"
	"net/mail"
)

const (
	MinPasswordLen = 8
	MaxPasswordLen = 72
	MaxUsernameLen = 64
)

func SignUpForm(name, password, email, reason string, approvalRequired bool) error {

	// TODO: validate reason
	
	var errs = []error{}
	
	errs = append(errs, Username(name))
	
	errs = append(errs, Email(email))

	errs = append(errs, Password(password))

	if approvalRequired {

	}

	return errors.Join(errs...)
}

func Password(password string) error {
	l := len(password)
	switch {
	case l == 0:
		return errors.New("empty password")
	case l < MinPasswordLen:
		return fmt.Errorf("password too short; min %d characters", MinPasswordLen)
	case l > MaxPasswordLen:
		return fmt.Errorf("password too long; max %d characters", MaxPasswordLen)
	}
	return nil
}

func Email(email string) error {
	if len(email) == 0 {
		return errors.New("empty email")
	}
	_, err := mail.ParseAddress(email)

	return err
}

func Username(username string) error {
	if l := len(username); l == 0 {
		return errors.New("empty username")
	} else if l > MaxUsernameLen {
		return fmt.Errorf("username too long; max %d characters", MaxUsernameLen)
	}
	return nil
}