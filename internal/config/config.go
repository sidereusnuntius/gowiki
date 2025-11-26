package config

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/viper"
)

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
	// FsRoot is the root of the directory on which files, such as the images and videos present in articles,
	// are stored.
	FsRoot string `mapstructure:"fs_root"`
	// FixedArticle is a list of the titles of articles deemed important by the wiki administrator, which are
	// displayed at the left side bar.
	FixedArticles []string
	// StaticDir is the directory on which the wiki's favicon, stylesheet, logo and other static files can be found.
	StaticDir string `mapstructure:"static_dir"`
	Language  string `mapstructure:"lang"`
	License   string `mapstructure:"license"`
	MediaType string `mapstructure:"markup"`
	// AutoPublish defines whether edits to articles are published automatically, or they should first be
	// reviewed and accepted by a trusted user before being published to readers. Will be removed.
	AutoPublish bool
	// InvitationRequired specifies whether new accounts on the instance can only be created through an invitation
	// link.
	InvitationRequired bool
	// ApprovalRequired specifies whether new accounts need to be reviewed and approved by an administrator before
	// being able to edit and create articles. Both InvitationRequired and AutoPublish cannot be true. Will be removed;
	// if the wiki does not require an invitation, it will automatically ask for a reason.
	ApprovalRequired bool
	// RsaKeySize specifies the size of the RSA keys to be used by the wiki in signing its outgoing activities.
	RsaKeySize int
	// Debug, if true, will make the application log all HTTP requests and other events.
	Debug bool `mapstructure:"debug"`
	// DbUrl is the path to the database file. Perhaps change this?
	DbUrl            string `mapstructure:"db_path"`
	MigrationsFolder string `mapstructure:"migrations"`
	// Name of the wiki.
	Name  string `mapstructure:"wiki_name"`
	Https bool   `mapstructure:"https"`
	// The name of the host runnig the application. Rename to host.
	Domain string `mapstructure:"hostname"`
	Port   uint16 `mapstructure:"port"`
	// Url is the instance's url.
	Url *url.URL
}

func ReadConfig() (Configuration, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	viper.SetDefault("fs_root", "./files")
	viper.SetDefault("debug", true)
	viper.SetDefault("lang", "en")
	viper.SetDefault("wiki_name", "Test")
	viper.SetDefault("hostname", "localhost")
	viper.SetDefault("port", uint16(8080))
	viper.SetDefault("markup", Markdown)
	viper.SetDefault("https", false)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Read error:", err)
		return Configuration{}, err
	}
	var cfg Configuration
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, err
	}

	var schema string
	if cfg.Https {
		schema = "https"
	} else {
		schema = "http"
	}

	port := strconv.FormatUint(uint64(cfg.Port), 10)
	u, err := url.Parse(schema + "://" + cfg.Domain + ":" + port + "/")
	if err != nil {
		return cfg, err
	}
	cfg.Url = u
	cfg.Domain = cfg.Domain + ":" + port
	return cfg, err
}
