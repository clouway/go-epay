package telcong

import "time"

// SubscriberDuties represents duties of the subscriber
type SubscriberDuties struct {
	CustomerName string `json:"customerName"`
	Address      string `json:"address"`
	DutyAmount   Amount `json:"dutyAmount"`
	Items        []Item `json:"items"`
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
