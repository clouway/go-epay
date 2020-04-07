package db

import (
	"context"
	"encoding/json"
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
		Metadata:      e.Metadata,
	}, nil
}

type environmentEntity struct {
	ID         *datastore.Key
	Type       string
	BillingKey string
	BillingURL string
	EpaySecret string
	MerchantID string
	Metadata   map[string]string `datastore:"-"`
}

func (e *environmentEntity) Load(ps []datastore.Property) error {
	// Stored fields could not be loaded when struct is not having the same field for safety. This check
	// ensures that entity will be loaded with it's metadata field.
	if err := datastore.LoadStruct(e, ps); err != nil && err != err.(*datastore.ErrFieldMismatch) {
		return err
	}

	for _, p := range ps {
		if p.Name == "metadata" {
			json.Unmarshal([]byte(p.Value.(string)), &e.Metadata)
		}
	}
	return nil
}

func (e *environmentEntity) Save() ([]datastore.Property, error) {
	props, err := datastore.SaveStruct(e)
	if err != nil {
		return nil, err
	}
	metadata, err := json.Marshal(e.Metadata)
	if err != nil {
		return nil, err
	}
	props = append(props, datastore.Property{
		Name:    "metadata",
		Value:   string(metadata),
		NoIndex: true,
	})

	return props, nil
}
