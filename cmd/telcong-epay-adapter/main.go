package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/clouway/go-epay/pkg/client/telcong"
	"github.com/clouway/go-epay/pkg/epay"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

var (
	listenAddr     = flag.String("listenAddr", ":5555", "the tcp listenAddr of the epay-adapter server")
	billingKeyFile = flag.String("billing-key-file", "app.key", "the path to the billing API keyfile")
	billingURL     = flag.String("billing-url", "https://cloud.telcong.com", "the url of the billing server")
)

const (
	// EPAY payment source
	EPAY epay.PaymentSource = "EPAY"
)

func main() {
	flag.Parse()

	conf, err := loadConf(*billingKeyFile)
	if err != nil {
		log.Fatalf("could not load billing-key-file '%s' due: %v", *billingKeyFile, err)
	}

	telcongURL, err := url.Parse(*billingURL)
	if err != nil {
		log.Fatalf("billing-url is not a valid url address")
	}

	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("unable to listen on: %s", *listenAddr)
	}

	oauth2client := conf.Client(context.Background())
	client := telcong.NewClient(oauth2client, telcongURL)

	server := epay.NewServer()
	go server.Serve(l, &telcongEpayGateway{client})

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Listening on %s", *listenAddr)
	log.Printf("Billing URL: %s", *billingURL)
	log.Printf("Billing Key File: %s", *billingKeyFile)

	log.Println("ePay adapter started successfully.")

	go func() {
		sig := <-sigs
		log.Printf("got: %v\n", sig)
		server.Stop()
		done <- true
	}()
	<-done

	log.Println("ePay adapter terminated successfully")
}

type telcongEpayGateway struct {
	client epay.Client
}

// GetCurrentBill returns the current bill of the provided customer.
func (t *telcongEpayGateway) GetCurrentBill(customerID, transactionID string) (*epay.BillResponse, error) {
	res, err := t.client.CreatePaymentOrder(context.Background(), epay.CreatePaymentOrderRequest{SubscriberID: customerID, TransactionID: transactionID, PaymentSource: EPAY})
	if err != nil {
		if err == epay.ErrPaymentOrderAlreadyExists {
			return nil, fmt.Errorf("bill with transactionId '%s' was already processed", transactionID)
		}

		if err == epay.ErrSubscriberNotFound {
			return &epay.BillResponse{Successful: false, UnknownSubscriber: true}, nil
		}

		return nil, err
	}
	// Convert currency to be applicable with the epay adapter as it
	// receives only value in coins.
	amount, _ := strconv.ParseFloat(res.Amount.Value, 64)
	inCoins := int(amount * 100)

	return &epay.BillResponse{Successful: true, Amount: inCoins}, nil
}

// PayBill pays bill using the provided amount
func (t *telcongEpayGateway) PayBill(customerID, transactionID string, Amount int) (*epay.PaymentResponse, error) {
	paymentOrder, err := t.client.GetPaymentOrder(context.Background(), transactionID)

	if err != nil {
		if err == epay.ErrPaymentOrderNotFound {
			return &epay.PaymentResponse{Successful: false}, nil
		}
		return nil, fmt.Errorf("could not retrieve payment order due: %v", err)
	}

	_, err = t.client.PayPaymentOrder(context.Background(), paymentOrder.ID)
	if err != nil {
		if err == epay.ErrPaymentOrderAlreadyPaid {
			return &epay.PaymentResponse{Successful: false, AlreadyPaid: true}, nil
		}
		return nil, err
	}
	return &epay.PaymentResponse{Successful: true}, nil
}

func loadConf(file string) (*jwt.Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("could not open configuration file '%s' with error: %v", file, err)
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("could not read configuration file '%s' with error: %v", file, err)
	}

	conf, err := google.JWTConfigFromJSON(c)
	if err != nil {
		return nil, fmt.Errorf("configuration file is not well formed: %v", err)
	}

	return conf, nil
}
