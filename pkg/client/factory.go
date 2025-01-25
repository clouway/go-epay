package client

import (
	"context"
	"net/url"
	"strconv"
	"strings"

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

func (c *clientFactory) Create(ctx context.Context, env epay.Environment, idn string) epay.Client {
	billingURL, _ := url.Parse(env.BillingURL)

	if isTelcoNGContractCode(idn) && env.BillingJWTKey != "" && env.BillingURL != "" {
		conf, _ := google.JWTConfigFromJSON([]byte(env.BillingJWTKey))
		oauth2client := conf.Client(ctx)
		return telcong.NewClient(oauth2client, billingURL)
	}

	if billingURL, ok := env.Metadata["billingUrl"]; ok {
		billingURL, _ := url.Parse(billingURL)
		apiKey := env.Metadata["apiKey"]
		methodID := env.Metadata["methodId"]
		providerName := env.Metadata["providerName"]
		providerPaymentID := env.Metadata["providerPaymentId"]
		providerPaymentTime := env.Metadata["providerPaymentTime"]
		organizationID := env.Metadata["organizationId"]

		return ucrm.NewClient(billingURL, apiKey, c.dClient, ucrm.PaymentProvider{
			MethodID:       methodID,
			Name:           providerName,
			PaymentID:      providerPaymentID,
			PaymentTime:    providerPaymentTime,
			OrganizationID: organizationID,
		})

	}

	// Default to telcong client
	conf, _ := google.JWTConfigFromJSON([]byte(env.BillingJWTKey))
	oauth2client := conf.Client(ctx)
	return telcong.NewClient(oauth2client, billingURL)
}

// isTelcoNGContractCode validates the provided code using the checksum algorithm
func isTelcoNGContractCode(code string) bool {
	const length = 7

	// Check if the length of the input is valid
	if len(code) != length {
		return false
	}

	// Separate the base number (neid) and the check digit (ncrc)
	neid := code[:length-1]
	ncrc, err := strconv.Atoi(code[length-1:])
	if err != nil {
		return false
	}

	// Reverse the neid string
	reversedNeid := reverseString(neid)

	// Calculate the checksum
	sum := 0
	for i := 0; i < len(reversedNeid); i++ {
		digit, _ := strconv.Atoi(string(reversedNeid[i]))
		if i%2 == 0 { // Odd positions (in the reversed string)
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	// Compute the expected check digit
	crc := (10 - (sum % 10)) % 10

	// Check if the computed check digit matches the provided one
	return crc == ncrc
}

// reverseString reverses the input string
func reverseString(s string) string {
	var sb strings.Builder
	for i := len(s) - 1; i >= 0; i-- {
		sb.WriteByte(s[i])
	}
	return sb.String()
}
