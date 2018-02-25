package telcong

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	userAgent = "telcong-golang-epay/20180125"
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

// Client is an HTTP client which uses API endpoints provided by the platform
// for dealing with it remotely
type Client struct {
	BaseURL *url.URL

	httpClient *http.Client
}

// NewClient creates a new HTTP client by using the provided baseURL.
func NewClient(httpClient *http.Client, baseURL *url.URL) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{httpClient: httpClient, BaseURL: baseURL}
	return c
}

// GetSubscriberDuties gets current subscriber duties.
func (c *Client) GetSubscriberDuties(ctx context.Context, subscriberID string) (*SubscriberDuties, error) {

	req, err := c.newRequest(ctx, "GET", "/v1/billing/"+subscriberID+"/duties", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request due: %v", err)
	}
	var duties SubscriberDuties
	resp, err := c.do(req, &duties)
	if err != nil {
		return nil, fmt.Errorf("could not process get subscriber duties request due: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		return &duties, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSubscriberNotFound
	}

	return nil, fmt.Errorf("unknown repospone during retrieving of subscriber duties: %v", err)
}

// CreatePaymentOrder creates a new PaymentOrder in the target system using the provided request.
func (c *Client) CreatePaymentOrder(ctx context.Context, createReq CreatePaymentOrderRequest) (*PaymentOrder, error) {
	req, err := c.newRequest(ctx, "POST", "/v1/paymentorders", createReq)
	if err != nil {
		return nil, fmt.Errorf("could not create request due: %v", err)
	}

	var paymentOrder PaymentOrder
	resp, err := c.do(req, &paymentOrder)
	if err != nil {
		return nil, fmt.Errorf("could not process create order request due: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		return &paymentOrder, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSubscriberNotFound
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, ErrPaymentOrderAlreadyExists
	}

	return nil, ErrUnknown
}

// GetPaymentOrder gets the PaymentOrder which is associated with the provided orderKey
func (c *Client) GetPaymentOrder(ctx context.Context, orderKey string) (*PaymentOrder, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/v1/paymentorders/%s", orderKey), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request due: %v", err)
	}
	var paymentOrder PaymentOrder
	resp, err := c.do(req, &paymentOrder)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve payment order due: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPaymentOrderNotFound
	}

	return &paymentOrder, nil
}

// PayPaymentOrder performs payment of the the order associated with the providing
// the ID of the order or the transactionID associated with it.
func (c *Client) PayPaymentOrder(ctx context.Context, orderID string) (*PayPaymentOrderResponse, error) {
	req, err := c.newRequest(context.Background(), "POST", fmt.Sprintf("/v1/paymentorders/%s/pay", orderID), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request due: %v", err)
	}

	var paymentResponse PayPaymentOrderResponse
	resp, err := c.do(req, &paymentResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to process payment request due: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		return &paymentResponse, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPaymentOrderNotFound
	}

	if resp.StatusCode == http.StatusBadRequest {
		er, derr := decodeResponse(resp.Body)
		if derr == nil && er.Message == "Payment order is already paid." {
			return nil, ErrPaymentOrderAlreadyPaid
		}
	}

	return nil, ErrUnknown
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return resp, nil
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

func decodeResponse(r io.ReadCloser) (*errorResponse, error) {
	defer r.Close()
	resp := errorResponse{}
	err := json.NewDecoder(r).Decode(&resp)
	return &resp, err
}

type errorResponse struct {
	Message string `json:"message"`
}
