package ucrm

import (
	"context"
	"math/rand"

	"cloud.google.com/go/datastore"
	"github.com/googleapis/google-cloud-go-testing/datastore/dsiface"
)

type fakeClient struct {
	dsiface.Client
	m map[string]interface{}
}

func newFakeClient() dsiface.Client {
	return &fakeClient{
		m: make(map[string]interface{}),
	}
}

func (c *fakeClient) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	if key.Name == "" && key.ID == 0 {
		key.ID = rand.Int63()
	}

	c.m[key.String()] = *(src.(*paymentOrder))
	return key, nil
}

func (c *fakeClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	val, ok := c.m[key.String()]
	if !ok {
		return datastore.ErrNoSuchEntity
	}

	sd := dst.(*paymentOrder)
	sv := val.(paymentOrder)
	*sd = sv
	return nil
}

func (c *fakeClient) Delete(ctx context.Context, key *datastore.Key) error {
	delete(c.m, key.String())
	return nil
}
