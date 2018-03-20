package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/clouway/go-epay/pkg/gae/config"
	"github.com/clouway/go-epay/pkg/gae/memcache"
	"github.com/clouway/go-epay/pkg/telcong"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type checkType string

const (
	check   checkType = "CHECK"
	billing checkType = "BILLING"

	StatusSuccess               string = "00"
	StatusSubscriberNotFound    string = "14"
	StatusNoDuties              string = "62"
	StatusTemporaryNotAvailable string = "80"
	StatusBadChecksum           string = "93"
	StatusAlreadyPaid           string = "94"
	StatusCommonError           string = "96"

	// EPAY payment source
	EPAY telcong.PaymentSource = "EPAY"
)

func init() {
	gateway := &epayGateway{}
	http.Handle("/v1/pay/", gateway)
}

type epayGateway struct {
	mu   sync.RWMutex //protects the following
	conf *jwt.Config
	env  *config.Environment
}

func (e *epayGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log.Debugf(ctx, "Looking for environment of: %s", r.URL.Host)
	env, err := config.GetEnv(ctx, r.URL.Host)
	if err != nil {
		log.Debugf(ctx, "unable to read environment due: %v", err)
		http.Error(w, "unable to read env configuration", http.StatusInternalServerError)
		return
	}
	conf, err := google.JWTConfigFromJSON([]byte(env.BillingJWTKey))
	if err != nil {
		http.Error(w, "configuration error", http.StatusInternalServerError)
		return
	}
	e.mu.Lock()
	e.conf = conf
	e.env = env
	e.mu.Unlock()

	idn := r.URL.Query().Get("IDN")
	checksum := r.URL.Query().Get("CHECKSUM")
	merchant := r.URL.Query().Get("MERCHANTID")
	transactionID := r.URL.Query().Get("TID")
	total := r.URL.Query().Get("TOTAL")
	operationType := checkType(r.URL.Query().Get("TYPE"))
	log.Debugf(ctx, "IDN=%s, checksum=%s, merchant=%s,transactionID=%s, operationType=%s, total=%s", idn, checksum, merchant, transactionID, operationType, total)

	if checksum != computeHmacSha1(r.URL.Query(), e.env.EpaySecret) {
		respondWithJSON(w, dutyResponse{Status: StatusBadChecksum})
		return
	}

	if strings.HasSuffix(r.URL.Path, "init") && (operationType == check || operationType == "") {
		e.checkBill(ctx, w, r)
	} else if strings.HasSuffix(r.URL.Path, "init") && operationType == billing {
		e.createPaymentOrder(ctx, w, r)
	} else if strings.HasSuffix(r.URL.Path, "confirm") {
		e.confirmPaymentOrder(ctx, w, r)
	} else {
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (e *epayGateway) checkBill(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var response *dutyResponse

	client := e.createClient(ctx)
	idn := r.URL.Query().Get("IDN")

	res, err := client.GetSubscriberDuties(ctx, idn)
	if err == nil {
		coins := inCoins(res.DutyAmount.Value)
		if coins == 0 {
			response = &dutyResponse{Status: StatusNoDuties}
		} else {
			response = successResponse(idn, res.CustomerName, res.Items, coins)
		}
	} else if err == telcong.ErrSubscriberNotFound {
		// not valid idn number
		response = &dutyResponse{Status: StatusSubscriberNotFound}
	} else {
		log.Debugf(ctx, "got unknown error: %v", err)
		// temporary not available
		response = &dutyResponse{Status: StatusTemporaryNotAvailable}
	}

	respondWithJSON(rw, response)
}

func (e *epayGateway) createPaymentOrder(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	log.Debugf(ctx, "Creating new payment order")

	var response *dutyResponse
	client := e.createClient(ctx)
	idn := r.URL.Query().Get("IDN")
	transactionID := r.URL.Query().Get("TID")

	res, err := client.CreatePaymentOrder(ctx, telcong.CreatePaymentOrderRequest{SubscriberID: idn, TransactionID: transactionID, PaymentSource: EPAY})
	if err == nil {
		coins := inCoins(res.Amount.Value)
		response = successResponse(idn, res.CustomerName, res.Items, coins)
	} else if err == telcong.ErrPaymentOrderAlreadyExists {
		response = &dutyResponse{Status: StatusNoDuties}
	} else if err == telcong.ErrSubscriberNotFound {
		response = &dutyResponse{Status: StatusSubscriberNotFound}
	} else {
		response = &dutyResponse{Status: StatusCommonError}
	}

	respondWithJSON(rw, response)
}

func (e *epayGateway) confirmPaymentOrder(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var response *dutyResponse

	transactionID := r.URL.Query().Get("TID")
	client := e.createClient(ctx)
	log.Debugf(ctx, "Confirming payment order with transaction: %s", transactionID)

	paymentOrder, err := client.GetPaymentOrder(ctx, transactionID)
	if err == nil {
		_, err = client.PayPaymentOrder(ctx, paymentOrder.ID)
		if err == nil {
			response = &dutyResponse{Status: StatusSuccess}
		} else if err == telcong.ErrPaymentOrderAlreadyPaid {
			// already paid notification
			response = &dutyResponse{Status: StatusAlreadyPaid}

		} else {
			log.Debugf(ctx, "could not confirm order due: %v", err)
			response = &dutyResponse{Status: StatusCommonError}
		}
	} else {
		log.Debugf(ctx, "got unknown error: %v", err)
		response = &dutyResponse{Status: StatusCommonError}
	}

	respondWithJSON(rw, response)
}

func successResponse(subscriberID, customerName string, items []telcong.Item, coins int) *dutyResponse {
	return &dutyResponse{IDN: subscriberID, Status: "00", ShortDesc: "Клиент: " + customerName, LongDesc: buildLongDesc(subscriberID, items), Amount: coins}
}

func (e *epayGateway) createClient(ctx context.Context) *telcong.Client {
	oauth2client := &http.Client{
		Transport: &oauth2.Transport{
			Source: memcache.TokenSource(ctx, e.conf.TokenSource(ctx)),
			Base: &urlfetch.Transport{
				Context: ctx,
			},
		},
	}
	billingURL, _ := url.Parse(e.env.BillingURL)
	return telcong.NewClient(oauth2client, billingURL)
}

func respondWithJSON(rw http.ResponseWriter, v interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(v); err != nil {
		http.Error(rw, "unable to serialize response", http.StatusInternalServerError)
	}
}

// inCoins converts the provided amount value in coins
func inCoins(value string) int {
	amount, _ := strconv.ParseFloat(value, 64)
	return int(amount * 100)
}

func buildLongDesc(subscriberID string, items []telcong.Item) string {
	lines := []string{}
	dup := make(map[string]string)
	for _, item := range items {
		_, ok := dup[item.Name]
		if ok {
			continue
		}
		dup[item.Name] = item.Name
		lines = append(lines, item.Name)
	}
	endDate := items[len(items)-1].EndDate

	return fmt.Sprintf("Клиентски Номер: %s,Задължения за периода до: %s, Детайли: %s", subscriberID, endDate.Format("02/01/2006"), strings.Join(lines, ","))
}

func computeHmacSha1(q url.Values, secret string) string {
	keys := make([]string, 0, len(q))
	for k := range q {
		if k == "CHECKSUM" {
			continue
		}
		keys = append(keys, k)
	}

	sort.Strings(keys)
	message := ""
	for _, k := range keys {
		message += fmt.Sprintf("%s%s\n", k, q[k][0])
	}

	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

type dutyResponse struct {
	Status    string `json:"STATUS,omitempty"`
	IDN       string `json:"IDN,omitempty"`
	ShortDesc string `json:"SHORTDESC,omitempty"`
	LongDesc  string `json:"LONGDESC,omitempty"`
	Amount    int    `json:"AMOUNT,omitempty"`
	ValidTo   string `json:"VALIDTO,omitempty"`
}
