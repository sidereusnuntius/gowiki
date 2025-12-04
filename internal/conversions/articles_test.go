package conversions

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"github.com/google/go-cmp/cmp"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

//go:embed articles
var articles []byte

func TestConvertArticle(t *testing.T) {
	var objects []map[string]any
	err := json.Unmarshal(articles, &objects)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{
		name string
		expected domain.ArticleFed
	}{
		{
			"valid article",
			domain.ArticleFed{
				ArticleCore: domain.ArticleCore{
					Title: "Prism",
					Host: "test.wiki",
					Summary: "Test article",
					Content: "Hello, world!",
					MediaType: "text/markdown",
					LastUpdated: toTime("2025-09-25T22:27:56.722186Z"),
					Published: toTime("2025-09-25T22:27:56.617343Z"),
				},
				AttributedTo: toURL("https://test.wiki/"),
				ApID: toURL("https://test.wiki/a/Prism"),
				Url: toURL("https://test.wiki/a/Prism"),
			},
		},
	}

	if len(cases) != len(objects) {
		t.Fatalf("length mismatch: there are %d test cases, but %d test AS objects", len(cases), len(objects))
	}

	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, err := streams.ToType(context.Background(), objects[i])
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			af, err := ConvertArticle(a)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(af, c.expected); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func toURL(u string) *url.URL {
	url, _ := url.Parse(u)
	return url
}

func toTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}