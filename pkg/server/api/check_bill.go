package api

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/clouway/go-epay/pkg/epay"
	"github.com/clouway/go-epay/pkg/server"
	"github.com/clouway/go-epay/pkg/server/httputil"
)

const (
	shortDescMaxLen = 40
	longDescMaxLen  = 4000
)

// CheckBill checks bill of customer
func CheckBill(cf epay.ClientFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)

		env := r.Context().Value(server.EnvironmentKey).(*epay.Environment)
		idn := r.URL.Query().Get("IDN")
		client := cf.Create(ctx, *env, idn)

		var response *DutyResponse

		res, err := client.GetSubscriberDuties(r.Context(), idn)
		if err == nil {

			coins := res.DutyAmount.InCoins()
			contextLogger.Printf("got duty amount: %d", coins)
			if coins == 0 {
				response = &DutyResponse{Status: StatusNoDuties}
			} else {
				contextLogger.Printf("checking bill of idn: %v", res.Items)
				response = successResponse(idn, res.CustomerName, res.Items, coins)
			}
		} else if err == epay.ErrSubscriberNotFound {
			contextLogger.Printf("subscriber '%s' was not found", idn)
			// not valid idn number
			response = &DutyResponse{Status: StatusSubscriberNotFound}
		} else {
			contextLogger.Printf("got unknown error: %v", err)
			// temporary not available
			response = &DutyResponse{Status: StatusTemporaryNotAvailable}
		}

		httputil.RespondWithJSON(ctx, w, response)
	})
}

func successResponse(subscriberID, customerName string, items []epay.Item, coins int) *DutyResponse {
	shortDesc := "Абонатен номер: " + subscriberID
	return &DutyResponse{IDN: subscriberID, Status: "00", ShortDesc: shortDesc, LongDesc: buildLongDesc(customerName, subscriberID, items), Amount: coins}
}

func buildLongDesc(customerName string, subscriberID string, items []epay.Item) string {
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

	longDesc := fmt.Sprintf("Клиент: %s, Абонатен Номер: %s, Детайли: %s", customerName, subscriberID, strings.Join(lines, ","))
	if len(longDesc) > longDescMaxLen {
		longDesc = longDesc[0:longDescMaxLen]
	}
	return longDesc
}
