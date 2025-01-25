package epay

import (
	"errors"
	"strconv"
	"time"

	"github.com/clouway/go-epay/pkg/number"
)

var (
	// ErrPaymentOrderAlreadyExists is the error used for registration
	// when PaymentOrder was already registered
	ErrPaymentOrderAlreadyExists = errors.New("duplication of PaymentOrders is not allowed")

	// ErrPaymentOrderNotFound is the error used during payment
	// when PaymentOrder was not found
	ErrPaymentOrderNotFound = errors.New("payment order was not found")

	// ErrPaymentOrderAlreadyPaid is the error used during paymeny
	// when PaymentOrder was already paid
	ErrPaymentOrderAlreadyPaid = errors.New("payment order was already paid")

	// ErrSubscriberNotFound is the error used for indication when
	// subscriber was not found
	ErrSubscriberNotFound = errors.New("the requested subscriber was not found")

	// ErrUnknown is the error which is return when no known cases
	// are recognized by the code
	ErrUnknown = errors.New("unknown error")
)

// Environment is representing a single environment in the context of the application.
type Environment struct {
	// The Billing JWT Key as string value. This key is issued
	// from iam.telcong.com and is available for everyone that has clouway account
	BillingJWTKey string

	// BillingKey is the billing API key for authentication
	BillingKey string

	// BillingURL is an URL API endpoint which to be used for retrieving of the
	// billing information
	BillingURL string

	// EpaySecret is the secret that is provided from ePay for verification
	// of the Checksum using HMAC SHA1 encoded as HEX
	EpaySecret string

	// MerchantID is the identifier of the merchant which was issued by ePay
	// provider
	MerchantID string

	// Metadata is a set of key-value pairs keeping for keeping of internal metadata attributes
	Metadata map[string]string
}

// SubscriberDuties represents duties of the subscriber
type SubscriberDuties struct {
	CustomerName string   `json:"customerName"`
	CustomerRef  string   `json:"customerRef"`
	Address      string   `json:"address"`
	DutyAmount   Amount   `json:"dutyAmount"`
	Items        []Item   `json:"items"`
	DocumentIDs  []string `json:"documents"`
}

// PaymentSource is representing the source of the payment
type PaymentSource string

// CreatePaymentOrderRequest represents the request for the creation
// of a new payment order
type CreatePaymentOrderRequest struct {
	SubscriberID  string        `json:"subscriberId"`
	PaymentSource PaymentSource `json:"paymentSource"`
	TransactionID string        `json:"transactionId"`
}

// PayPaymentOrderResponse is representing the respons which is returned when payment
// order is paid
type PayPaymentOrderResponse struct {
	ID            string    `json:"id"`
	CustomerName  string    `json:"customerName"`
	TransactionID string    `json:"transactionId"`
	Amount        Amount    `json:"amount"`
	Created       time.Time `json:"created"`
	PaidOn        time.Time `json:"paidOn"`
	Items         []Item    `json:"items"`
}

// PaymentOrder is a single PaymentOrder
type PaymentOrder struct {
	ID            string    `json:"id"`
	CustomerName  string    `json:"customerName"`
	TransactionID string    `json:"transactionId"`
	Amount        Amount    `json:"amount"`
	Created       time.Time `json:"created"`
	Items         []Item    `json:"items"`
}

// Item is a single item line.
type Item struct {
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Amount    Amount    `json:"amount"`
	Vat       float64   `json:"vat"`
	Price     string    `json:"price"`
	Quantity  int       `json:"quantity"`
}

// Amount represents Order amount
type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// InCoins gets the amount value in coins.
func (a Amount) InCoins() int {
	amount, _ := strconv.ParseFloat(a.Value, 64)
	return int(number.Round(amount*100.00, 2))
}
