package client

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/httpsig"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

var prefs = []httpsig.Algorithm{httpsig.RSA_SHA256}
var getHeaders = []string{httpsig.RequestTarget, "date"}
var postHeaders = []string{httpsig.RequestTarget, "date", "digest"}
var mainKey, _ = url.Parse("#main-key")

// HttpClient is a client to be used by the wiki instance actor itself, such as when announcing edits.
// When it is needed to dereference or deliver objects on behalf of a particular actor, the client
// can create a new Transport for the given actor.
type HttpClient struct {
	db              db.DB
	client          *http.Client
	key             crypto.PrivateKey
	pubKeyId        *url.URL
	getSigner       httpsig.Signer
	getSignerMutex  sync.Mutex
	postSigner      httpsig.Signer
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
		db:         db,
		client:     client,
		key:        key,
		pubKeyId:   keyId,
		getSigner:  getSigner,
		postSigner: postSigner,
	}, nil
}

func (c *HttpClient) Get(ctx context.Context, iri *url.URL) (obj vocab.Type, err error) {
	res, err := c.Dereference(ctx, iri)
	if err != nil {
		log.Error().Err(err).Msg("failed to do request")
		return
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var props map[string]any
	err = decoder.Decode(&props)
	if err != nil {
		log.Error().Err(err).Msg("response body unmarshaling error")
		return
	}

	obj, err = streams.ToType(ctx, props)
	return
}

func (c *HttpClient) Dereference(ctx context.Context, iri *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iri.String(), nil)
	if err != nil {
		return nil, err
	}

	c.getSignerMutex.Lock()
	defer c.getSignerMutex.Unlock()
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("Accept", "application/activity+json")
	err = c.getSigner.SignRequest(c.key, c.pubKeyId.String(), req, nil)
	if err != nil {
		log.Error().Err(err).Msg("error while signing request")
		return nil, err
	}

	res, err := c.client.Do(req)
	if res.StatusCode >= 400 {
		event := log.Error().Str("status", res.Status)
		var content []byte
		content, err = io.ReadAll(res.Body)

		if err != nil {
			event.Err(err)
		}
		event.Bytes("response", content).Msg("fetch error")
		res.Body.Close()
		err = fmt.Errorf("%d %s: %s", res.StatusCode, res.Status, content)
	}

	return res, err
}

func (c *HttpClient) DeliverAs(ctx context.Context, obj map[string]any, to *url.URL, from *url.URL) error {
	if path := from.Path; path == "" || path == "/" {
		return c.Deliver(ctx, obj, to)
	}

	key, err := c.db.GetUserPrivateKeyByURI(ctx, from)
	if err != nil {
		log.Error().Err(err).Msg("user's private key not found")
		return err
	}

	signer, _, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, postHeaders, httpsig.Signature, 3600)
	if err != nil {
		log.Error().Err(err).Msg("failed to construct signer")
		return err
	}

	pubKeyid := from.ResolveReference(mainKey)
	log.Info().Str("public key", pubKeyid.String()).Send()
	transport := pub.NewHttpSigTransport(c.client, "gowiki", c, nil, signer, pubKeyid.String(), key)
	return transport.Deliver(ctx, obj, to)
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
	req.Header.Add("Content-Type", "application/activity+json")

	c.postSignerMutex.Lock()
	defer c.postSignerMutex.Unlock()
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	err = c.postSigner.SignRequest(c.key, c.pubKeyId.String(), req, body)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(res.Body)
		log.Error().Int("code", res.StatusCode).Bytes("response body", body).Msg("delivery error")
		return fmt.Errorf("error %d: %s", res.StatusCode, res.Status)
	}
	return nil
}

func (c *HttpClient) BatchDeliver(ctx context.Context, obj map[string]interface{}, recipients []*url.URL) error {
	return errors.New("not yet implemented")
}

func (c *HttpClient) NewTransport(ctx context.Context, prefs []httpsig.Algorithm, id *url.URL) (transport pub.Transport, err error) {
	key, err := c.db.GetUserPrivateKeyByURI(ctx, id)
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

	id.Fragment = "main-key"
	transport = pub.NewHttpSigTransport(c.client, "gowiki", c, getSigner, postSigner, id.String(), key)
	return
}

func (c *HttpClient) Now() time.Time {
	return time.Now()
}
