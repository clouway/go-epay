package api

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"net/http"
	"strings"

	"github.com/clouway/go-epay/pkg/epay"
	"github.com/clouway/go-epay/pkg/server"
	"github.com/clouway/go-epay/pkg/server/httputil"
)

// CheckBill checks bill of customer
func CheckBill(cf epay.ClientFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)

		env := r.Context().Value(server.EnvironmentKey).(*epay.Environment)
		client := cf.Create(ctx, *env)

		var response *DutyResponse

		idn := r.URL.Query().Get("IDN")
		contextLogger.Debugf("checking bill of idn: %s", idn)
		res, err := client.GetSubscriberDuties(r.Context(), idn)
		if err == nil {

			coins := res.DutyAmount.InCoins()
			contextLogger.Debugf("got duty amount: %d", coins)
			if coins == 0 {
				response = &DutyResponse{Status: StatusNoDuties}
			} else {
				response = successResponse(idn, res.CustomerName, res.Items, coins)
			}
		} else if err == epay.ErrSubscriberNotFound {
			contextLogger.Debugf("subscriber '%s' was not found", idn)
			// not valid idn number
			response = &DutyResponse{Status: StatusSubscriberNotFound}
		} else {
			contextLogger.Debugf("got unknown error: %v", err)
			// temporary not available
			response = &DutyResponse{Status: StatusTemporaryNotAvailable}
		}

		httputil.RespondWithJSON(ctx, w, response)
	})
}

func successResponse(subscriberID, customerName string, items []epay.Item, coins int) *DutyResponse {
	return &DutyResponse{IDN: subscriberID, Status: "00", ShortDesc: "Клиент: " + customerName, LongDesc: "", Amount: coins}
}

func buildLongDesc(subscriberID string, items []epay.Item) string {
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
