package ucrm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/clouway/go-epay/pkg/epay"

	"cloud.google.com/go/datastore"
)

const poKind = "PaymentOrder"

// NewClient creates a new client that uses the provided app key and baseURL.
func NewClient(baseURL *url.URL, appKey string, dClient *datastore.Client) epay.Client {
	return &client{BaseURL: baseURL, AppKey: appKey, dClient: dClient}
}

type client struct {
	BaseURL *url.URL
	AppKey  string
	dClient *datastore.Client
}

// GetSubscriberDuties gets current subscriber duties.
func (c *client) GetSubscriberDuties(ctx context.Context, subscriberID string) (*epay.SubscriberDuties, error) {
	clientRef, err := c.findClientID(ctx, subscriberID)
	if err != nil {
		return nil, epay.ErrSubscriberNotFound
	}

	clientID := strconv.Itoa(clientRef.ID)
	customerName := clientRef.FirstName + " " + clientRef.LastName
	if clientRef.CompanyName != "" {
		customerName = clientRef.CompanyName
	}

	params := url.Values{}
	params.Add("clientId", clientID)
	params.Add("statuses[0]", "1")
	params.Add("statuses[1]", "2")
	params.Add("limit", "100")

	req, err := c.newRequest(ctx, "GET", "/api/v1.0/invoices", params)
	if err != nil {
		return nil, fmt.Errorf("could not create request due: %v", err)
	}
	var duties []invoice
	resp, err := c.do(req, &duties)
	if err != nil {
		return nil, fmt.Errorf("could not process get subscriber duties request due: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		dutyAmount := 0.0
		documentIDs := make([]string, 0)
		items := make([]epay.Item, 0)
		for _, duty := range duties {
			dutyAmount += duty.Total - duty.AmountPaid
			documentID := strconv.Itoa(duty.ID)
			documentIDs = append(documentIDs, documentID)

			for _, item := range duty.Items {
				items = append(items, epay.Item{Name: item.Label})
			}
		}

		amount := fmt.Sprintf("%.2f", dutyAmount)
		return &epay.SubscriberDuties{
			CustomerName: customerName,
			CustomerRef:  clientID,
			DutyAmount:   epay.Amount{Value: amount},
			DocumentIDs:  documentIDs,
			Items:        items,
		}, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, epay.ErrSubscriberNotFound
	}

	return nil, fmt.Errorf("unknown response during retrieving of subscriber duties: %v", err)
}

func (c *client) CreatePaymentOrder(ctx context.Context, createReq epay.CreatePaymentOrderRequest) (*epay.PaymentOrder, error) {
	contextLogger := log.WithContext(ctx)

	duties, err := c.GetSubscriberDuties(ctx, createReq.SubscriberID)
	if err != nil {
		return nil, err
	}

	k := datastore.NameKey(poKind, createReq.TransactionID, nil)

	po := &paymentOrder{
		CustomerName:  duties.CustomerName,
		ClientID:      duties.CustomerRef,
		TransactionID: createReq.TransactionID,
		SubscriberID:  createReq.SubscriberID,
		Amount:        duties.DutyAmount.Value,
		CreatedAt:     time.Now(),
		InvoiceIDs:    duties.DocumentIDs,
	}

	if _, err := c.dClient.Put(ctx, k, po); err != nil {
		contextLogger.Printf("got error: %v", err)
		return nil, epay.ErrUnknown
	}

	return &epay.PaymentOrder{
		ID:            k.Name,
		CustomerName:  po.CustomerName,
		TransactionID: po.TransactionID,
		Amount:        epay.Amount{Value: po.Amount},
		Created:       po.CreatedAt,
		Items:         duties.Items,
	}, nil
}

func (c *client) GetPaymentOrder(ctx context.Context, orderKey string) (*epay.PaymentOrder, error) {
	k := datastore.NameKey(poKind, orderKey, nil)

	po := &paymentOrder{}
	if err := c.dClient.Get(ctx, k, po); err != nil {
		return nil, epay.ErrPaymentOrderNotFound
	}

	return &epay.PaymentOrder{
		ID:            fmt.Sprintf("%d", k.ID),
		CustomerName:  po.CustomerName,
		TransactionID: po.TransactionID,
		Amount:        epay.Amount{Value: po.Amount},
		Created:       po.CreatedAt,
	}, nil
}

func (c *client) PayPaymentOrder(ctx context.Context, orderID string) (*epay.PayPaymentOrderResponse, error) {
	k := datastore.NameKey(poKind, orderID, nil)

	po := &paymentOrder{}
	if err := c.dClient.Get(ctx, k, po); err != nil {
		return nil, epay.ErrPaymentOrderNotFound
	}

	var invoiceIds = []int{}
	for _, i := range po.InvoiceIDs {
		invoiceID, _ := strconv.Atoi(i)
		invoiceIds = append(invoiceIds, invoiceID)
	}

	clientID, _ := strconv.Atoi(po.ClientID)
	amount, _ := strconv.ParseFloat(po.Amount, 64)
	paymentReq := &paymentRequest{
		ClientID:                     clientID,
		MethodID:                     "d8c1eae9-d41d-479f-aeaf-38497975d7b3",
		CheckNumber:                  "",
		CreatedDate:                  jsonDate{time.Now()},
		Amount:                       amount,
		CurrencyCode:                 "BGN",
		Note:                         "Paid in coins",
		InvoiceIDs:                   invoiceIds,
		ProviderName:                 "Worldpay",
		ProviderPaymentID:            "WP451837",
		ProviderPaymentTime:          "2016-09-12T00:00:00+0000",
		ApplyToInvoicesAutomatically: false,
		UserID:                       1,
	}

	req, err := c.newRequest(ctx, "POST", "/api/v1.0/payments", paymentReq)
	if err != nil {
		return nil, fmt.Errorf("could not create request due: %v", err)
	}

	r := &paymentResponse{}
	resp, err := c.do(req, &r)
	if err != nil {
		return nil, fmt.Errorf("could not process payment request due: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Printf("got response status: %s", resp.Status)
		return nil, epay.ErrUnknown
	}

	po.ProcessedOn = time.Now()
	if _, err := c.dClient.Put(ctx, k, po); err != nil {
		log.Printf("got error: %v", err)
		return nil, epay.ErrUnknown
	}

	return &epay.PayPaymentOrderResponse{
		ID:            orderID,
		TransactionID: po.TransactionID,
		Amount:        epay.Amount{Value: po.Amount},
		Created:       po.CreatedAt,
		PaidOn:        po.ProcessedOn,
	}, nil
}

func (c *client) findClientID(ctx context.Context, subscriberID string) (*clientRef, error) {
	params := url.Values{}
	params.Add("userIdent", subscriberID)

	req, err := c.newRequest(ctx, "GET", "/api/v1.0/clients", params)

	if err != nil {
		return nil, fmt.Errorf("could not create request due: %v", err)
	}

	var clients []clientRef
	resp, err := c.do(req, &clients)

	if err != nil {
		return nil, fmt.Errorf("could not process get subscriber duties request due: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got bad response due: %v", err)
	}
	if len(clients) == 0 {
		return nil, epay.ErrSubscriberNotFound
	}

	return &clients[0], nil
}

func (c *client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		switch v := body.(type) {
		case url.Values:
			u.RawQuery = v.Encode()
		default:
			buf = new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(body)
			if err != nil {
				return nil, err
			}
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
	req.Header.Set("X-Auth-App-Key", c.AppKey)

	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
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

type clientRef struct {
	ID          int    `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	CompanyName string `json:"companyName"`
}

type invoice struct {
	ID                int           `json:"id"`
	Total             float64       `json:"total"`
	AmountPaid        float64       `json:"amountPaid"`
	ClientFirstName   string        `json:"clientFirstName"`
	ClientLastName    string        `json:"clientLastName"`
	ClientCompanyName string        `json:"clientCompanyName"`
	Items             []invoiceItem `json:"items"`
}

type invoiceItem struct {
	Label string `json:"label"`
}

type paymentOrder struct {
	SubscriberID  string    `datastore:"subscriberId,noindex"`
	CustomerName  string    `datastore:"customerName,noindex"`
	ClientID      string    `datastore:"clientID,noindex"`
	TransactionID string    `datastore:"transactionId,noindex"`
	Amount        string    `datastore:"amount,noindex"`
	CreatedAt     time.Time `datastore:"createdOn,noindex"`
	ProcessedOn   time.Time `datastore:"processedOn,omitempty"`
	InvoiceIDs    []string  `datastore:"invoiceIds,noindex"`
}

type paymentRequest struct {
	ClientID                     int      `json:"clientId"`
	MethodID                     string   `json:"methodId"`
	CheckNumber                  string   `json:"checkNumber"`
	CreatedDate                  jsonDate `json:"createdDate"`
	Amount                       float64  `json:"amount"`
	CurrencyCode                 string   `json:"currencyCode"`
	Note                         string   `json:"note"`
	InvoiceIDs                   []int    `json:"invoiceIds"`
	ProviderName                 string   `json:"providerName"`
	ProviderPaymentID            string   `json:"providerPaymentId"`
	ProviderPaymentTime          string   `json:"providerPaymentTime"`
	ApplyToInvoicesAutomatically bool     `json:"applyToInvoicesAutomatically"`
	UserID                       int      `json:"userId"`
}

type paymentResponse struct {
	ID int `json:"id"`
}
