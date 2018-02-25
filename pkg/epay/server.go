package epay

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// Gateway is a generic gateway to the remote billing system
type Gateway interface {
	// GetCurrentBill returns the current bill of the provided customer.
	GetCurrentBill(customerID, transactionID string) (*BillResponse, error)

	// PayBill pays bill using the provided amount
	PayBill(customerID, transactionID string, Amount int) (*PaymentResponse, error)
}

// BillResponse is representing the response from the billing
type BillResponse struct {
	Successful        bool
	UnknownSubscriber bool
	Amount            int
}

// Status gets bill response status
func (br *BillResponse) Status() Status {
	if br.Successful {
		return BillReturned
	} else if br.UnknownSubscriber {
		return UnknownSubscriber
	} else {
		return CommonError
	}
}

// PaymentResponse is representing the response of payment in the billing system.
type PaymentResponse struct {
	AlreadyPaid bool
	Successful  bool
}

// Status gets payment response status
func (pr *PaymentResponse) Status() Status {
	if pr.Successful {
		return PaymentProcessed
	} else if pr.AlreadyPaid {
		return PaymentAlreadyProcessed
	} else {
		return CommonError
	}
}

// Server is representing an implementation of the epay server
type Server struct {
	listener net.Listener

	quit chan bool
}

// NewServer creates a new instance of the EpayServer
func NewServer() *Server {
	s := &Server{quit: make(chan bool)}
	return s
}

// Serve serves the incoming connections received from the provided listener
func (s *Server) Serve(l net.Listener, gateway Gateway) error {
	defer l.Close()
	s.listener = l

	for {
		tcpConn := l.(*net.TCPListener)
		tcpConn.SetDeadline(time.Now().Add(300 * time.Millisecond))

		c, err := l.Accept()
		//Check for the channel being closed
		select {
		case <-s.quit:
			fmt.Println("finishing task")
			fmt.Println("task done")
			s.quit <- true
			return nil
		default:
			//If the channel is still open, continue as normal
		}
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			}

		}

		go s.handle(c, gateway)

	}
}

func (s *Server) handle(c io.ReadWriteCloser, gateway Gateway) {
	defer c.Close()
	req, err := parseRequest(c)
	var resp response
	if err != nil {
		log.Printf("could not parse request due: %v", err)
		resp = &paymentResponse{CommonError}
		resp.Write(c)
		return
	}

	if req.IsForBillCheck() {
		if cb, err := gateway.GetCurrentBill(req.CustomerID, req.TransactionID); err != nil {
			resp = &billResponse{Status: CommonError}
		} else {
			log.Printf("unable to call billing due: %v", err)
			resp = &billResponse{Amount: cb.Amount, Status: cb.Status()}
		}
	} else if req.IsForPayment() {
		pr, err := gateway.PayBill(req.CustomerID, req.TransactionID, req.Amount)
		if err != nil {
			resp = &paymentResponse{CommonError}
		} else {
			resp = &paymentResponse{pr.Status()}
		}
	}

	resp.Write(c)
}

// Stop stops listening of incomming connections and terminates
// server instance
func (s *Server) Stop() {
	s.quit <- true
	<-s.quit
}
