package config

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/appengine/datastore"
)

// Environment is representing a single environment in the context of the application.
type Environment struct {
	// The Billing JWT Key as string value. This key is issued
	// from iam.telcong.com and is available for everyone that has clouway account
	BillingJWTKey string

	// BillingURL is the billing is the URL API which to be used for retrieving of the
	// billing information
	BillingURL string

	// EpaySecret is the secret that is provided from ePay for verification
	// of the Checksum using HMAC SHA1 encoded as HEX
	EpaySecret string

	// MerchantID is the identifier of the merchant which was issued by ePay
	// provider
	MerchantID string
}

// GetEnv gets the configuration environment associated with the provided name. The GetEnv
// tries to return the default environment if it detects that target hostname is for GAE.
func GetEnv(ctx context.Context, name string) (*Environment, error) {

	// A default name should be used
	if strings.Contains(name, "appspot") {
		name = "default"
	}

	key := datastore.NewKey(ctx, "Environment", name, 0, nil)
	e := new(environmentEntity)
	if err := datastore.Get(ctx, key, e); err != nil {
		return nil, fmt.Errorf("could not load the environment '%s' due: %v", name, err)
	}
	return &Environment{BillingJWTKey: e.BillingKey, BillingURL: e.BillingURL, EpaySecret: e.EpaySecret, MerchantID: e.MerchantID}, nil
}

type environmentEntity struct {
	BillingKey string `datastore:"billingKey"`
	BillingURL string `datastore:"billingURL"`
	EpaySecret string `datastore:"epaySecret"`
	MerchantID string `datastore:"merchantId"`
}
