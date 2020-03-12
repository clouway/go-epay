package middleware

import (
	"net/http"

	"github.com/clouway/go-epay/pkg/server/api"
	"github.com/clouway/go-epay/pkg/server/httputil"
	log "github.com/sirupsen/logrus"
)

// Skip skips the next request processing if the queryParam is contained in the passed
// entry map.
func Skip(queryParam string, entry map[string]interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			contextLogger := log.WithContext(ctx)

			v := r.URL.Query().Get(queryParam)
			_, ok := entry[v]

			if !ok {
				next.ServeHTTP(w, r)
			} else {
				contextLogger.Debugf("Skipping processing of '%s' as it is ignored", v)
				httputil.RespondWithJSON(ctx, w, api.DutyResponse{Status: "14"})
			}
		})
	}

}
