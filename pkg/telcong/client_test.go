package telcong

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestCreatePaymentOrder(t *testing.T) {
	serverResponse := &CreatePaymentOrderResponse{ID: "1"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, serverResponse)
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	resp, err := client.CreatePaymentOrder(CreatePaymentOrderRequest{SubscriberID: "::sub::", TransactionID: "TID1"})
	if err != nil {
		t.Fatalf("unable to create payment order due: %v", err)
	}

	if !reflect.DeepEqual(resp, serverResponse) {
		t.Errorf("expected response to be: %v", serverResponse)
		t.Errorf("	              got: %v", resp)
	}
}

func TestCreatePaymentOrderWithServerFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "could not handle request", http.StatusInternalServerError)
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	_, err := client.CreatePaymentOrder(CreatePaymentOrderRequest{SubscriberID: "::sub::", TransactionID: "TID1"})
	if err == nil {
		t.Fatal("expected error due the error of the server, but got nothing")
	}
}

func TestTryToCreateOrderForNotExistingSubscriber(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unknown subscriber", http.StatusNotFound)
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	_, err := client.CreatePaymentOrder(CreatePaymentOrderRequest{SubscriberID: "::unknown::", TransactionID: "TID2"})

	if err != ErrSubscriberNotFound {
		t.Errorf("	expected: %v", ErrSubscriberNotFound)
		t.Errorf("	     got: %v", err)
	}
}

func TestPaymentOrderAlreadyExists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unknown subscriber", http.StatusBadRequest)
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	_, err := client.CreatePaymentOrder(CreatePaymentOrderRequest{SubscriberID: "::unknown::", TransactionID: "TID2"})

	if err != ErrPaymentOrderAlreadyExists {
		t.Errorf("	expected: %v", ErrPaymentOrderAlreadyExists)
		t.Errorf("	     got: %v", err)
	}
}

func TestPayPaymentOrder(t *testing.T) {
	serverResponse := &PayPaymentOrderResponse{ID: "::any order id::", TransactionID: "TID2"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, serverResponse)
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	resp, err := client.PayPaymentOrder("::any order id::")

	if err != nil {
		t.Fatal("error should not be returned for successful payment")
	}

	if !reflect.DeepEqual(resp, serverResponse) {
		t.Errorf("	expected: %v", serverResponse)
		t.Errorf("	     got: %v", resp)
	}
}

func TestPayUnknownPaymentOrder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unknown", http.StatusNotFound)
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	_, err := client.PayPaymentOrder("::unknown order id::")

	if err != ErrPaymentOrderNotFound {
		t.Errorf("	expected: %v", ErrPaymentOrderNotFound)
		t.Errorf("	     got: %v", err)
	}

}

func TestPayAlreadyPaidPaymentOrder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		jsonReply(w, &errorResponse{Message: "Payment order is already paid."})
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	_, err := client.PayPaymentOrder("::paid order id::")

	if err != ErrPaymentOrderAlreadyPaid {
		t.Errorf("	expected: %v", ErrPaymentOrderAlreadyPaid)
		t.Errorf("	     got: %v", err)
	}

}

func TestPaymentFailsWithServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal", http.StatusInternalServerError)
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(nil, baseURL)

	_, err := client.PayPaymentOrder("::order id::")

	if err != ErrUnknown {
		t.Errorf("	expected: %v", ErrUnknown)
		t.Errorf("	     got: %v", err)
	}

}

func jsonReply(w http.ResponseWriter, v interface{}) {
	jsonVal, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonVal)
}
