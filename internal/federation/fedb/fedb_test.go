package fedb

import (
	"context"
	"net/url"
	"testing"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sidereusnuntius/gowiki/internal/config"
)

var configuration config.Configuration
var ctx = context.Background()
var host = "https://test.wiki"
var u, _ = url.Parse(host)

func TestMain(m *testing.M) {
	configuration = config.Configuration{
		FsRoot:     "./files",
		Language:   "en",
		License:    config.CcBy,
		MediaType:  config.Text,
		RsaKeySize: 2048,
		Debug:      true,
		Domain:     "test.wiki",
		Name:       "The test wiki",
		Url:        u,
	}

	m.Run()
}
