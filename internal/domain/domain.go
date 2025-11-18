package domain

import (
	"net/url"
)

type UserCore struct {
	ID       int64
	Username string
	Name     string
	Domain   string
	Summary  string
	URL      *url.URL
}

type Revision struct {
	ID       int64
	Title    string
	Reviewed bool
	Summary  string
	Username string
	Created  int64
}

type Profile struct {
	UserCore
	Edits []Revision
}

type UserInternal struct {
	UserCore
}
