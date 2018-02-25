package epay

import (
	"io"
	"net"
	"testing"

	"github.com/clouway/go-epay/pkg/epay/epaytest"
)

func TestGetCurrentBill(t *testing.T) {
	s := NewServer()
	defer s.Stop()
	l, _ := net.Listen("tcp", ":0")
	go s.Serve(l, &fakeGateway{billResponse: &BillResponse{Successful: true, Amount: 360}, err: nil})

	epayServer, tearDown := epaytest.NewServer(t, l.Addr().String())
	defer tearDown()
	response := epayServer.GetCurrentBill("123", "T1")
	if exp := "XTYPE=RBN\nXVALIDTO=\nAMOUNT=360\nSTATUS=00\n"; exp != response {
		t.Errorf("expected: %s", exp)
		t.Errorf("     got: %s", response)
	}
}

func TestGetCurrentBillFails(t *testing.T) {
	s := NewServer()
	defer s.Stop()
	l, _ := net.Listen("tcp", ":0")
	go s.Serve(l, &fakeGateway{err: io.ErrClosedPipe})

	epayServer, tearDown := epaytest.NewServer(t, l.Addr().String())
	defer tearDown()
	response := epayServer.GetCurrentBill("123", "T1")
	if exp := "XTYPE=RBN\nXVALIDTO=\nAMOUNT=0\nSTATUS=96\n"; exp != response {
		t.Errorf("expected: %s", exp)
		t.Errorf("     got: %s", response)
	}
}

func TestGatewaySendsBrokenRequest(t *testing.T) {
	s := NewServer()
	defer s.Stop()
	l, _ := net.Listen("tcp", ":0")
	go s.Serve(l, nil)

	epayServer, tearDown := epaytest.NewServer(t, l.Addr().String())
	defer tearDown()
	response := epayServer.DummyRequest("::broken::")
	if exp := "XTYPE=RBC\nSTATUS=96\n"; exp != response {
		t.Errorf("expected: %s", exp)
		t.Errorf("     got: %s", response)
	}
}

func TestPayBill(t *testing.T) {
	s := NewServer()
	defer s.Stop()
	l, _ := net.Listen("tcp", ":0")
	go s.Serve(l, &fakeGateway{paymentResponse: &PaymentResponse{Successful: true}, err: nil})

	epayServer, tearDown := epaytest.NewServer(t, l.Addr().String())
	defer tearDown()
	response := epayServer.PayBill("123", "T1", 10)
	if exp := "XTYPE=RBC\nSTATUS=00\n"; exp != response {
		t.Errorf("expected: %s", exp)
		t.Errorf("     got: %s", response)
	}
}

func TestPayBillWasNotSuccessful(t *testing.T) {
	s := NewServer()
	defer s.Stop()
	l, _ := net.Listen("tcp", ":0")
	go s.Serve(l, &fakeGateway{paymentResponse: &PaymentResponse{Successful: false}, err: nil})

	epayServer, tearDown := epaytest.NewServer(t, l.Addr().String())
	defer tearDown()
	response := epayServer.PayBill("123", "T1", 10)
	if exp := "XTYPE=RBC\nSTATUS=96\n"; exp != response {
		t.Errorf("expected: %s", exp)
		t.Errorf("     got: %s", response)
	}
}

func TestBillAlreadyPaid(t *testing.T) {
	s := NewServer()
	defer s.Stop()
	l, _ := net.Listen("tcp", ":0")
	go s.Serve(l, &fakeGateway{paymentResponse: &PaymentResponse{AlreadyPaid: true, Successful: false}, err: nil})

	epayServer, tearDown := epaytest.NewServer(t, l.Addr().String())
	defer tearDown()
	response := epayServer.PayBill("123", "T1", 10)
	if exp := "XTYPE=RBC\nSTATUS=94\n"; exp != response {
		t.Errorf("expected: %s", exp)
		t.Errorf("     got: %s", response)
	}
}

type fakeGateway struct {
	billResponse    *BillResponse
	paymentResponse *PaymentResponse
	err             error
}

func (f *fakeGateway) GetCurrentBill(CustomerID, TransactionID string) (*BillResponse, error) {
	return f.billResponse, f.err
}

func (f *fakeGateway) PayBill(CustomerID, TransactionID string, Amount int) (*PaymentResponse, error) {
	return f.paymentResponse, f.err
}
