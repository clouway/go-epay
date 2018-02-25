package telcong

import "time"

// PaymentSource is representing the source of the payment
type PaymentSource string

// CreatePaymentOrderRequest represents the request for the creation
// of a new payment order
type CreatePaymentOrderRequest struct {
	SubscriberID  string        `json:"subscriberId"`
	PaymentSource PaymentSource `json:"paymentSource"`
	TransactionID string        `json:"transactionId"`
}

// CreatePaymentOrderResponse represents the response of the creation
type CreatePaymentOrderResponse struct {
	ID            string             `json:"id"`
	TransactionID string             `json:"transactionId"`
	Amount        Amount             `json:"amount"`
	Created       time.Time          `json:"created"`
	Items         []PaymentOrderItem `json:"items"`
}

type PayPaymentOrderResponse struct {
	ID            string             `json:"id"`
	TransactionID string             `json:"transactionId"`
	Amount        Amount             `json:"amount"`
	Created       time.Time          `json:"created"`
	PaidOn        time.Time          `json:"paidOn"`
	Items         []PaymentOrderItem `json:"items"`
}

// PaymentOrderItem is a single item in the PaymentOrder.
type PaymentOrderItem struct {
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
