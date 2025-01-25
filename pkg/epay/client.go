package epay

import "context"

// ClientFactory creates a client for particular environment.
type ClientFactory interface {
	// Create creates a new client for the provided environment.
	Create(ctx context.Context, env Environment, idn string) Client
}

// Client is representing a client to billing.
type Client interface {
	// GetSubscriberDuties gets duties of subscriber.
	GetSubscriberDuties(ctx context.Context, subscriberID string) (*SubscriberDuties, error)

	// CreatePaymentOrder creates a new PaymentOrder in the target system using the provided request.
	CreatePaymentOrder(ctx context.Context, createReq CreatePaymentOrderRequest) (*PaymentOrder, error)

	// GetPaymentOrder gets the PaymentOrder which is associated with the provided orderKey.
	GetPaymentOrder(ctx context.Context, orderKey string) (*PaymentOrder, error)

	// PayPaymentOrder performs payment of the the order associated with the providing
	// the ID of the order or the transactionID associated with it.
	PayPaymentOrder(ctx context.Context, orderID string) (*PayPaymentOrderResponse, error)
}
