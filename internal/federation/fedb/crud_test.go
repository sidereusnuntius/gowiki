package fedb

import (
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/google/go-cmp/cmp"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	mock_db "github.com/sidereusnuntius/gowiki/internal/mocks"
	"go.uber.org/mock/gomock"
)

type comparison struct {
	property string
	expect   string
	got      any
}

func TestGet_UserSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	DB := mock_db.NewMockDB(ctrl)
	fdb := New(DB, configuration)

	sarah := makeUser(1, "sarah")
	article := makeArticle(
		"tests",
		"an introduction to software testing",
		"something beautiful will happen",
	)

	noteIRI, _ := url.Parse("https://elysium.social/statuses/247189")
	noteJSON, note := makeNote(noteIRI)

	cases := []struct {
		name             string
		iri              *url.URL
		cacheEntry       domain.FedObj
		returnedValue    any
		expectedErr      error
		returnedAsObject vocab.Type
		expectErr        bool
		expectedType     string
	}{
		{
			name:             "UserSuccess",
			iri:              u.JoinPath("u", "sarah"),
			cacheEntry:       domain.FedObj{Iri: u.JoinPath("u", "sarah"), RawJSON: "", ApType: "Person", Local: true, LocalTable: "users", LocalId: 1},
			returnedValue:    sarah,
			returnedAsObject: conversions.UserToActor(sarah),
			expectErr:        false,
			expectedType:     streams.ActivityStreamsPersonName,
		},
		{
			name:             "ArticleSuccess",
			iri:              u.JoinPath("a", "tests"),
			cacheEntry:       domain.FedObj{Iri: u.JoinPath("a", "tests"), RawJSON: "", ApType: "Article", Local: true, LocalTable: "articles", LocalId: 5},
			returnedValue:    article,
			returnedAsObject: conversions.ArticleToObject(article),
			expectedType:     streams.ActivityStreamsArticleName,
		},
		{
			name:             "RawJSON",
			iri:              u.JoinPath("notes", "123987"),
			cacheEntry:       domain.FedObj{Iri: noteIRI, RawJSON: noteJSON, ApType: "Note", Local: false},
			returnedAsObject: note,
			expectedType:     streams.ActivityStreamsNoteName,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			DB.EXPECT().
				GetApObject(gomock.Any(), c.iri).
				Return(c.cacheEntry, nil)

			switch c.cacheEntry.LocalTable {
			case "users":
				DB.EXPECT().
					GetUserByID(gomock.Any(), c.cacheEntry.LocalId).
					Return(c.returnedValue, nil).
					Times(1)
			case "articles":
				DB.EXPECT().
					GetArticleById(gomock.Any(), c.cacheEntry.LocalId).
					Return(c.returnedValue, nil).
					Times(1)
			}

			returned, err := fdb.Get(ctx, c.iri)
			if err != nil {
				if c.expectedErr == nil {
					t.Errorf("unexpected error: %v", err)
				} else if !errors.Is(err, c.expectedErr) {
					t.Errorf("expected error \"%s\", got \"%s\"", c.expectedErr, err)
				}
				return
			}

			if returned.GetTypeName() != c.expectedType {
				t.Errorf("expected an object of type %s, got a %s", c.expectedType, returned.GetTypeName())
			}

			returnedSerial, err := streams.Serialize(returned)
			if err != nil {
				t.Fatalf("Serialize() on returned value failed: %v", err)
			}
			expectedSerial, err := streams.Serialize(c.returnedAsObject)
			if err != nil {
				t.Fatalf("Serialize() on expected value failed: %v", err)
			}

			if diff := cmp.Diff(expectedSerial, returnedSerial); diff != "" {
				t.Error(diff)
			}
		})
	}
	//	Return(user, nil)
}

func makeNote(iri *url.URL) (string, vocab.Type) {
	note := streams.NewActivityStreamsNote()
	id := streams.NewJSONLDIdProperty()
	id.SetIRI(iri)
	note.SetJSONLDId(id)

	content := streams.NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("Hello, world!")
	note.SetActivityStreamsContent(content)

	n2, _ := streams.Serialize(note)
	b, _ := json.Marshal(n2)
	return string(b), note
}

func makeArticle(title, summary, content string) domain.ArticleFed {
	iri := u.JoinPath("a", title)
	return domain.ArticleFed{
		ApID: iri,
		Url:  iri,
		ArticleCore: domain.ArticleCore{
			Title:     title,
			Summary:   summary,
			Content:   content,
			MediaType: config.Text,
		},
	}
}

func makeUser(id int64, name string) domain.UserFed {
	apId := u.JoinPath("u", name)
	return domain.UserFed{
		UserCore: domain.UserCore{
			ID:       id,
			Username: name,
			Name:     name,
			Domain:   host,
			Summary:  "A test user",
			URL:      apId,
		},
		ApId:        apId,
		Inbox:       apId.JoinPath("inbox"),
		Outbox:      apId.JoinPath("outbox"),
		Followers:   apId.JoinPath("followers"),
		PublicKey:   "",
		Created:     time.Now(),
		LastUpdated: time.Now(),
	}
}
