package api

import (
	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/clouway/go-epay/pkg/epay"
	"github.com/clouway/go-epay/pkg/server"
	"github.com/clouway/go-epay/pkg/server/httputil"
)

// CreatePaymentOrder creates a new handler for the creation of Payment Order.
func CreatePaymentOrder(cf epay.ClientFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		env := r.Context().Value(server.EnvironmentKey).(*epay.Environment)
		client := cf.Create(ctx, *env)
		idn := r.URL.Query().Get("IDN")
		transactionID := r.URL.Query().Get("TID")
		contextLogger.Printf("IDN: %s, TID: %s", idn, transactionID)

		res, err := client.CreatePaymentOrder(ctx, epay.CreatePaymentOrderRequest{SubscriberID: idn, TransactionID: transactionID, PaymentSource: EPAY})

		var response *DutyResponse
		if err == nil {
			coins := res.Amount.InCoins()
			response = successResponse(idn, res.CustomerName, res.Items, coins)
		} else if err == epay.ErrPaymentOrderAlreadyExists {
			response = &DutyResponse{Status: StatusNoDuties}
		} else if err == epay.ErrSubscriberNotFound {
			response = &DutyResponse{Status: StatusSubscriberNotFound}
		} else {
			contextLogger.Printf("got error: %v", err)
			response = &DutyResponse{Status: StatusCommonError}
		}

		httputil.RespondWithJSON(ctx, w, response)
	})
}
