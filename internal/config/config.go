package config

import "net/url"

type Configuration struct {
	InvitationRequired bool
	ApprovalRequired bool
	RsaKeySize int
	Https bool
	Debug bool
	DbUrl string
	Domain string
	Url *url.URL
}