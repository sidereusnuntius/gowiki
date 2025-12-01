package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"code.superseriousbusiness.org/httpsig"
	"github.com/rs/zerolog/log"
	mock_db "github.com/sidereusnuntius/gowiki/internal/mocks"
	"go.uber.org/mock/gomock"
)

var client HttpClient
var key *rsa.PrivateKey
var algo = httpsig.RSA_SHA256
var ctx = context.Background()

func TestMain(m *testing.M) {
	var err error
	key, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal().Err(err).Msg("tests setup failure")
		return
	}

	m.Run()
}

func verify(t *testing.T, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifier, err := httpsig.NewVerifier(r)
		if err != nil {
			t.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if path != r.URL.Path {
			t.Errorf("expected path %s, got %s", path, r.URL.Path)
		}

		err = verifier.Verify(&key.PublicKey, algo)
		if err != nil {
			t.Error("signature validation error:", err)
			return
		}
		w.Write([]byte("hello!"))
	})
}

func TestDereference(t *testing.T) {
	ctrl := gomock.NewController(t)
	DB := mock_db.NewMockDB(ctrl)

	kId, _ := url.Parse("http://localhost:8080")
	client, err := New(DB, &http.Client{}, key, []httpsig.Algorithm{algo}, kId)
	if err != nil {
		t.Fatal(err)
	}

	path := "/someguy"
	server := httptest.NewServer(verify(t, path))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	url := u.JoinPath(path)
	res, err := client.Dereference(ctx, url)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if b := string(body); b != "hello!" {
		t.Errorf("unexpected response: \"%s\"", b)
	}
}
