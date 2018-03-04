package memcache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// TokenSource creates a new token source which uses the provided one as source and
// uses memcahce service of GAE for caching of tokens.
func TokenSource(ctx context.Context, ts oauth2.TokenSource) oauth2.TokenSource {
	return &memcacheTokenSource{ctx, ts}
}

type memcacheTokenSource struct {
	ctx context.Context

	ts oauth2.TokenSource
}

func (m *memcacheTokenSource) Token() (token *oauth2.Token, err error) {
	if item, merr := memcache.Get(m.ctx, "token"); merr == nil {
		log.Debugf(m.ctx, "retrieved token from cache")
		err = json.NewDecoder(bytes.NewReader(item.Value)).Decode(&token)
	} else if merr == memcache.ErrCacheMiss {
		log.Debugf(m.ctx, "token was not found and looking for new one")
		// not found in cache
		token, err = m.ts.Token()
		if err != nil {
			return nil, fmt.Errorf("could not retrieve token due: %v", err)
		}
		tokenValue, _ := json.Marshal(token)
		memcache.Set(m.ctx, &memcache.Item{Key: "token", Value: tokenValue, Expiration: 4 * time.Hour})
	} else {
		// got error
		log.Debugf(m.ctx, "got memcache error: %v", merr)
		return nil, fmt.Errorf("got memcache error: %v", err)
	}

	if token.Valid() {
		return token, nil
	}

	token, err = m.ts.Token()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve token due: %v", err)
	}

	tokenValue, _ := json.Marshal(token)
	memcache.Set(m.ctx, &memcache.Item{Key: "token", Value: tokenValue})

	return token, nil
}
