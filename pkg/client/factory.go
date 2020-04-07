package client

import (
	"context"
	"net/url"

	"github.com/clouway/go-epay/pkg/client/telcong"
	"github.com/clouway/go-epay/pkg/client/ucrm"
	"github.com/clouway/go-epay/pkg/epay"
	"golang.org/x/oauth2/google"

	"cloud.google.com/go/datastore"
)

// NewClientFactory creates a new Factory for Client creation.
func NewClientFactory(dClient *datastore.Client) epay.ClientFactory {
	return &clientFactory{dClient}
}

type clientFactory struct {
	dClient *datastore.Client
}

func (c *clientFactory) Create(ctx context.Context, env epay.Environment) epay.Client {
	billingURL, _ := url.Parse(env.BillingURL)

	// Environment is of type UCRM
	if env.Type == "ucrm" {
		methodID := env.Metadata["methodId"]
		providerName := env.Metadata["providerName"]
		providerPaymentID := env.Metadata["providerPaymentId"]
		providerPaymentTime := env.Metadata["providerPaymentTime"]

		return ucrm.NewClient(billingURL, env.BillingKey, c.dClient, ucrm.PaymentProvider{
			MethodID:    methodID,
			Name:        providerName,
			PaymentID:   providerPaymentID,
			PaymentTime: providerPaymentTime,
		})
	}

	conf, _ := google.JWTConfigFromJSON([]byte(env.BillingJWTKey))
	oauth2client := conf.Client(ctx)
	return telcong.NewClient(oauth2client, billingURL)

}
