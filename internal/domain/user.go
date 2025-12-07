package domain

import (
	"net/url"
	"time"
)

type Collective struct {
	Type       string
	Name       string
	Hostname   string
	Summary string
	Url        *url.URL
	PublicKey string
	Inbox      *url.URL
	Outbox     *url.URL
	Followers  *url.URL
}

type Account struct {
	ApId      *url.URL
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
	Host     string
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
	Articles []ArticlePreview
	Edits    []Revision
}

type ArticlePreview struct {
	Title string
}

type UserInternal struct {
	UserCore
}

type Session struct {
	UserID    int64
	AccountID int64
	Username  string
	ApId      *url.URL
}
