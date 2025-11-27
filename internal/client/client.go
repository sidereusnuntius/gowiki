package client

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/httpsig"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

var getHeaders = []string{httpsig.RequestTarget, "date"}
var postHeaders = []string{httpsig.RequestTarget, "date", "digest"}

// HttpClient is a client to be used by the wiki instance actor itself, such as when announcing edits.
// When it is needed to dereference or deliver objects on behalf of a particular actor, the client
// can create a new Transport for the given actor.
type HttpClient struct {
	db db.DB
	client *http.Client
	key crypto.PrivateKey
	pubKeyId *url.URL
	getSigner httpsig.Signer
	getSignerMutex sync.Mutex
	postSigner httpsig.Signer
	postSignerMutex sync.Mutex
}

func New(db db.DB, client *http.Client, key crypto.PrivateKey, prefs []httpsig.Algorithm, keyId *url.URL) (*HttpClient, error) {
	getSigner, _, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, getHeaders, httpsig.Signature, 3600)
	if err != nil {
		return nil, err
	}

	postSigner, _, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, postHeaders, httpsig.Signature, 3600)
	if err != nil {
		return nil, err
	}

	return &HttpClient{
		db: db,
		client: client,
		key: key,
		pubKeyId: keyId,
		getSigner: getSigner,
		postSigner: postSigner,
	}, nil
}

func (c *HttpClient) Dereference(ctx context.Context, iri *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iri.String(), nil)
	if err != nil {
		return nil, err
	}

	c.getSignerMutex.Lock()
	defer c.getSignerMutex.Unlock()
	req.Header.Set("Date", time.Now().Format(time.RFC3339))
	err = c.getSigner.SignRequest(c.key, c.pubKeyId.String(), req, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	return res, err
}

func (c *HttpClient) Deliver(ctx context.Context, obj map[string]interface{}, to *url.URL) error {
	body, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, to.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	c.postSignerMutex.Lock()
	defer c.postSignerMutex.Unlock()
	req.Header.Add("Date", time.Now().Format(time.RFC3339))
	err = c.postSigner.SignRequest(c.key, c.pubKeyId.String(), req, body)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		return errors.New("error: response status:" + res.Status)
	}
	return nil
}

func (c *HttpClient) BatchDeliver(ctx context.Context, obj map[string]interface{}, recipients []*url.URL) error {
	return errors.New("not yet implemented")
}

func (c *HttpClient) NewTransport(ctx context.Context, prefs []httpsig.Algorithm, id int64) (transport pub.Transport, err error) {
	owner, key, err := c.db.GetUserPrivateKey(ctx, id)
	if err != nil {
		return
	}
	
	getSigner, chosenAlgo, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, getHeaders, httpsig.Signature, 3600)
	if err != nil {
		return
	}

	log.Debug().Msg("chosen algorithm " + string(chosenAlgo))

	postSigner, chosenAlgo, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, postHeaders, httpsig.Signature, 3600)
	if err != nil {
		return
	}

	owner.Fragment = "main-key"
	transport = pub.NewHttpSigTransport(c.client, "gowiki", c, getSigner, postSigner, owner.String(), key)
	return
}

func (c *HttpClient) Now() time.Time {
	return time.Now()
}