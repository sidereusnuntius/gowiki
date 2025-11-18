package config

import "net/url"

const (
	Text     = "text/plain"
	Markdown = "text/markdown"
)

const (
	CcBy     = "CC BY"
	CcBySa   = "CC BY-SA"
	CcByNc   = "CC BY-NC"
	CcByNcSa = "CC BY-NC-SA"
)

type Configuration struct {
	// FixedArticle is a list of the titles of articles deemed important by the wiki administrator, which are
	// displayed at the left side bar.
	FixedArticles []string
	// StaticDir is the directory on which the wiki's favicon, stylesheet, logo and other static files can be found.
	StaticDir string
	Language  string
	License   string
	MediaType string
	// AutoPublish defines whether edits to articles are published automatically, or they should first be
	// reviewed and accepted by a trusted user before being published to readers.
	AutoPublish bool
	// InvitationRequired specifies whether new accounts on the instance can only be created through an invitation
	// link.
	InvitationRequired bool
	// ApprovalRequired specifies whether new accounts need to be reviewed and approved by an administrator before
	// being able to edit and create articles. Both InvitationRequired and AutoPublish cannot be true.
	ApprovalRequired bool
	RsaKeySize       int
	Debug            bool
	DbUrl            string
	// Name of the wiki.
	Name   string
	Https  bool
	Domain string
	Url    *url.URL
}
