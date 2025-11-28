package domain

import (
	"net/url"
	"time"
)

type Collective struct {
	Type string
	Name string
	Hostname string
	Url *url.URL
	Public_key string
	Inbox *url.URL
	Outbox *url.URL
	Followers *url.URL
}

type Account struct {
	UserID    int64
	AccountID int64
	Username  string
	Email     string
	Password  string
	Admin     bool
}

type UserCore struct {
	ID       int64
	Username string
	Name     string
	Domain   string
	Summary  string
	URL      *url.URL
}

type UserFed struct {
	UserCore
	ApId        *url.URL
	Inbox       *url.URL
	Outbox      *url.URL
	Followers   *url.URL
	PublicKey   string
	Created     time.Time
	LastUpdated time.Time
}

type UserFedInternal struct {
	UserFed
	Trusted    bool
	PrivateKey string
}

type Profile struct {
	UserCore
	Edits []Revision
}

type UserInternal struct {
	UserCore
}
