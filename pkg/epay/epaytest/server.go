package epaytest

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// TestServer testing server that simulates Epay requestor
type TestServer struct {
	c *net.TCPConn
}

// NewServer creates a new testing epay server that tries to
// connect to the provided host
func NewServer(t *testing.T, host string) (*TestServer, func()) {
	c, err := net.DialTimeout("tcp4", host, time.Second)
	if err != nil {
		t.Fatalf("unable to connect to testing server due: %v", err)
	}
	tearDown := func() {
		c.Close()
	}
	return &TestServer{c.(*net.TCPConn)}, tearDown
}

// DummyRequest allows sending of dummy request to the
// target epay adapter
func (f *TestServer) DummyRequest(cmd string) string {
	return f.sendCommand(cmd)
}

// GetCurrentBill asks for duties of a given customer
func (f *TestServer) GetCurrentBill(customerID, transactionID string) string {
	cmd := fmt.Sprintf("XTYPE=QBN\nIDN=%s\nTID=%s\n", customerID, transactionID)
	return f.sendCommand(cmd)
}

// PayBill notifies trader for executed payment
func (f *TestServer) PayBill(customerID, transactionID string, amount int) string {
	cmd := fmt.Sprintf("XTYPE=QBC\nIDN=%s\nTID=%s\nAMOUNT=%d\n", customerID, transactionID, amount)
	return f.sendCommand(cmd)
}

func (f *TestServer) sendCommand(cmd string) string {
	f.c.Write([]byte(cmd))
	f.c.CloseWrite()
	return f.readResponse()
}

func (f *TestServer) readResponse() string {
	buf := make([]byte, 512)
	n, _ := f.c.Read(buf)
	return string(buf[0:n])
}
