package epay

import (
	"fmt"
	"io"
)

// Status is representing the response status which is returned
// back to the epay processor
type Status string

const (
	// BillReturned indicates that bill is returned successfully
	BillReturned Status = "00"
	// NoCurrentBill indicates that subscriber doesn't have current bill
	NoCurrentBill Status = "62"
	// UnknownSubscriber indicates that subscriber ID was broken
	UnknownSubscriber Status = "14"

	// PaymentProcessed indicates that payment was processed successfully
	PaymentProcessed Status = "00"
	// PaymentAlreadyProcessed indicates that payment was already processed
	PaymentAlreadyProcessed Status = "94"
	// CommonError indicates an error which was occurred during payment
	CommonError Status = "96"
)

// request is representing a single EPAY request
type request struct {
	Type          string
	CustomerID    string
	TransactionID string
	Amount        int
}

// response is representing an abstraction of responses which are returned
// to epay
type response interface {
	Write(w io.Writer) (int, error)
}

type billResponse struct {
	Amount int
	Status Status
}

func (b *billResponse) Write(w io.Writer) (int, error) {
	cmd := fmt.Sprintf("XTYPE=RBN\nXVALIDTO=%s\nAMOUNT=%d\nSTATUS=%s\n", "", b.Amount, string(b.Status))
	return w.Write([]byte(cmd))
}

type paymentResponse struct {
	Status Status
}

func (p *paymentResponse) Write(w io.Writer) (int, error) {
	cmd := fmt.Sprintf("XTYPE=RBC\nSTATUS=%s\n", string(p.Status))
	return w.Write([]byte(cmd))
}
