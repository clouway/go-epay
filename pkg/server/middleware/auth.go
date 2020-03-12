package middleware

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/clouway/go-epay/pkg/epay"
	"github.com/clouway/go-epay/pkg/server"
	"github.com/clouway/go-epay/pkg/server/api"
	"github.com/clouway/go-epay/pkg/server/httputil"
)

// EpayAPIMiddleware is a middleware used to check the request using internal secret
// stored in the environment.
func EpayAPIMiddleware(envStore epay.EnvironmentStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextLogger := log.WithContext(r.Context())
			host := r.Header.Get("X-Appengine-Default-Version-Hostname")
			if host == "" {
				host = r.URL.Host
			}
			env, err := envStore.Get(r.Context(), host)

			if err != nil {
				contextLogger.Debugf("unable to read environment due: %v", err)
				http.Error(w, "unable to read env configuration", http.StatusInternalServerError)
				return
			}

			checksum := r.URL.Query().Get("CHECKSUM")

			if checksum != epay.Checksum(r.URL.Query(), env.EpaySecret) {
				httputil.RespondWithJSON(r.Context(), w, api.ErrBadChecksum)
				return
			}

			nextCtx := context.WithValue(r.Context(), server.EnvironmentKey, env)
			next.ServeHTTP(w, r.WithContext(nextCtx))
		})
	}
}
