package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/datastore"
	"golang.org/x/oauth2"
)

// TokenSource creates a new token source which uses the provided one as source and
// uses memcahce service of GAE for caching of tokens.
func TokenSource(ctx context.Context, client *datastore.Client, ts oauth2.TokenSource, tenant string) oauth2.TokenSource {
	return &dbTokenSource{ctx, client, ts, tenant}
}

type dbTokenSource struct {
	ctx    context.Context
	c      *datastore.Client
	ts     oauth2.TokenSource
	tenant string
}

func (d *dbTokenSource) Token() (token *oauth2.Token, err error) {
	k := datastore.NameKey("Token", d.tenant, nil)
	e := &tokenEntity{}

	if err := d.c.Get(d.ctx, k, e); err != nil {
		return d.refreshAndUpdate(k)
	}

	err = json.NewDecoder(bytes.NewReader([]byte(e.Payload))).Decode(&token)

	if token.Valid() {
		return token, nil
	}

	return d.refreshAndUpdate(k)
}

func (d *dbTokenSource) refreshAndUpdate(k *datastore.Key) (token *oauth2.Token, err error) {
	token, err = d.ts.Token()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve token due: %v", err)
	}

	tokenValue, _ := json.Marshal(token)
	if _, err = d.c.Put(d.ctx, k, &tokenEntity{
		ID:      k,
		Payload: string(tokenValue),
	}); err != nil {
		return nil, fmt.Errorf("got datastore error: %v", err)
	}

	return token, nil
}

type tokenEntity struct {
	ID      *datastore.Key
	Payload string `datastore:"payload,noindex"`
}
