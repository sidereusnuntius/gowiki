package domain

import "net/url"

type ArticleCore struct {
	Title     string
	Summary   string
	Content   string
	Protected bool
	MediaType string
	License   string
	Language  string
}

type ArticleFed struct {
	ArticleCore
	ApID *url.URL
	Url  *url.URL
}

type Revision struct {
	ID       int64
	Title    string
	Reviewed bool
	Diff     string
	Summary  string
	Username string
	Created  int64
}

//InstanceID sql.NullInt64
