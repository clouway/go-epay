package httputil

import (
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// RespondWithJSON is writing the provided value as JSON response with a proper content type.
func RespondWithJSON(ctx context.Context, rw http.ResponseWriter, v interface{}) {
	payload, err := json.Marshal(v)
	contextLogger := log.WithContext(ctx)

	if err != nil {
		contextLogger.Debugf("unable to build json response due: %v", err)
		http.Error(rw, "unable to serialize response", http.StatusInternalServerError)
		return
	}
	contextLogger.Debugf("response: %s", string(payload))

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(payload)
}
