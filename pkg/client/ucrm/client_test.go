package ucrm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/clouway/go-epay/pkg/epay"
)

func TestSusbscriberNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1.0/clients", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("[]"))
	}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(baseURL, "testing-key", nil, PaymentProvider{})
	_, err := client.GetSubscriberDuties(context.Background(), "::subscriber id::")
	if err != epay.ErrSubscriberNotFound {
		t.Fatalf("expected subscriber not found but got: %v", err)
	}
}

func TestResidentialSubscriberDuties(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1.0/clients", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := `[
			{
			   "id": 708,           			   
			   "firstName": "John",
			   "lastName": "Smith",
			   "companyName": ""
			}
		 ]`
		w.Write([]byte(content))
	}))
	mux.HandleFunc("/api/v1.0/invoices", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := `[
			{
			   "id": 101,           
			   "total":20.0,
			   "amountPaid": 0.0,
			   "clientFirstName": "John",
			   "clientLastName": "Smith",
			   "items":[
					{
						"label":"service 1 04/2020"
					}
				]
			}
		 ]`
		w.Write([]byte(content))
	}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(baseURL, "testing-key", nil, PaymentProvider{})
	resp, err := client.GetSubscriberDuties(context.Background(), "::subscriber id::")
	if err != nil {
		t.Fatalf("unable to retrieve subscriber duties due: %v", err)
	}

	serverResponse := &epay.SubscriberDuties{
		CustomerName: "John Smith",
		CustomerRef:  "708",
		DutyAmount:   epay.Amount{Value: "20.00"},
		Items:        []epay.Item{epay.Item{Name: "service 1 04/2020"}},
		DocumentIDs:  []string{"101"},
	}
	if !reflect.DeepEqual(resp, serverResponse) {
		t.Errorf("expected response to be: %v", serverResponse)
		t.Errorf("	              got: %v", resp)
	}
}

func TestGetDutiesWhenMultipleInvoices(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1.0/clients", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := `[
			{
			   "id": 708,           			   
			   "firstName": "John",
			   "lastName": "Smith",
			   "companyName": ""
			}
		 ]`
		w.Write([]byte(content))
	}))
	mux.HandleFunc("/api/v1.0/invoices", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := `[
			{
			   "id": 101,           
			   "total":20.0,
			   "amountPaid": 10.0,
			   "clientFirstName": "John",
			   "clientLastName": "Smith",
			   "items":[
					{
						"label":"service 1 04/2020"
					}
				]
			},
			{
				"id": 102,           
				"total":12.0,
				"amountPaid": 0.0,
				"clientFirstName": "John",
				"clientLastName": "Smith",
				"items":[
					{
						"label":"service 1 05/2020"
					}
				]
			 },
			 {
				"id": 103,           
				"total":13.43,
				"amountPaid": 0.0,
				"clientFirstName": "John",
				"clientLastName": "Smith",
				"items":[
					{
						"label":"service 1 06/2020"
					}
				]
			 }
		 ]`
		w.Write([]byte(content))
	}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(baseURL, "testing-key", nil, PaymentProvider{})
	resp, err := client.GetSubscriberDuties(context.Background(), "::subscriber id::")
	if err != nil {
		t.Fatalf("unable to retrieve subscriber duties due: %v", err)
	}

	serverResponse := &epay.SubscriberDuties{
		CustomerName: "John Smith",
		CustomerRef:  "708",
		DutyAmount:   epay.Amount{Value: "35.43"},
		DocumentIDs:  []string{"101", "102", "103"},
		Items: []epay.Item{
			epay.Item{Name: "service 1 04/2020"},
			epay.Item{Name: "service 1 05/2020"},
			epay.Item{Name: "service 1 06/2020"},
		},
	}

	if !reflect.DeepEqual(resp, serverResponse) {
		t.Errorf("expected response to be: %v", serverResponse)
		t.Errorf("	              got: %v", resp)
	}
}

func TestGetDutiesWhenNoInvoices(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1.0/clients", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := `[
			{
			   "id": 708,           			   
			   "firstName": "John",
			   "lastName": "Smith",
			   "companyName": "My Company"
			}
		 ]`
		w.Write([]byte(content))
	}))
	mux.HandleFunc("/api/v1.0/invoices", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := "[]"
		w.Write([]byte(content))
	}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	baseURL, _ := url.Parse(ts.URL)

	client := NewClient(baseURL, "testing-key", nil, PaymentProvider{})
	resp, err := client.GetSubscriberDuties(context.Background(), "::subscriber id::")
	if err != nil {
		t.Fatalf("unable to retrieve subscriber duties due: %v", err)
	}

	serverResponse := &epay.SubscriberDuties{
		CustomerName: "My Company",
		CustomerRef:  "708",
		DutyAmount:   epay.Amount{Value: "0.00"},
		DocumentIDs:  []string{},
		Items:        []epay.Item{},
	}
	if !reflect.DeepEqual(resp, serverResponse) {
		t.Errorf("expected response to be: %v", serverResponse)
		t.Errorf("	              got: %v", resp)
	}
}

func jsonReply(w http.ResponseWriter, v interface{}) {
	jsonVal, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonVal)
}
