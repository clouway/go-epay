package db

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/clouway/go-epay/pkg/epay"
)

// NewEnvironmentStore creates a new environment store that is using datastore as a backend layer.
func NewEnvironmentStore(client *datastore.Client) epay.EnvironmentStore {
	return &store{client}
}

type store struct {
	c *datastore.Client
}

func (s *store) Get(ctx context.Context, name string) (*epay.Environment, error) {

	// A default name should be used
	if strings.Contains(name, "appspot") || name == "" {
		name = "default"
	}

	k := datastore.NameKey("Environment", name, nil)

	e := &environmentEntity{}
	if err := s.c.Get(ctx, k, e); err != nil {
		return nil, fmt.Errorf("could not load the environment '%s' due: %v", name, err)
	}

	return &epay.Environment{
		Type:          e.Type,
		BillingJWTKey: e.BillingKey,
		BillingKey:    e.BillingKey,
		BillingURL:    e.BillingURL,
		EpaySecret:    e.EpaySecret,
		MerchantID:    e.MerchantID,
	}, nil
}

type environmentEntity struct {
	ID         *datastore.Key
	Type       string
	BillingKey string
	BillingURL string
	EpaySecret string
	MerchantID string
}
