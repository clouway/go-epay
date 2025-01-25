package api

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/clouway/go-epay/pkg/epay"
	"github.com/clouway/go-epay/pkg/server"
	"github.com/clouway/go-epay/pkg/server/httputil"
)

// ConfirmPaymentOrder creates a new handler for the creation of Payment Order.
func ConfirmPaymentOrder(cf epay.ClientFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)

		env := ctx.Value(server.EnvironmentKey).(*epay.Environment)
		idn := r.URL.Query().Get("IDN")
		client := cf.Create(r.Context(), *env, idn)

		transactionID := r.URL.Query().Get("TID")

		contextLogger.Printf("Confirming payment order with transaction: %s", transactionID)

		_, err := client.PayPaymentOrder(ctx, transactionID)

		var response *DutyResponse
		if err == nil {
			response = &DutyResponse{Status: StatusSuccess}
		} else if err == epay.ErrPaymentOrderAlreadyPaid {
			response = &DutyResponse{Status: StatusAlreadyPaid}
		} else {
			contextLogger.Printf("could not confirm order due: %v", err)
			response = &DutyResponse{Status: StatusCommonError}
		}

		httputil.RespondWithJSON(ctx, w, response)
	})
}
