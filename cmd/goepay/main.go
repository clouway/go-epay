package main

import (
	"context"
	"net/http"
	"os"

	"github.com/andyfusniak/stackdriver-gae-logrus-plugin"
	lmiddleware "github.com/andyfusniak/stackdriver-gae-logrus-plugin/middleware"
	"github.com/clouway/go-epay/pkg/client"
	"github.com/clouway/go-epay/pkg/server/api"
	"github.com/clouway/go-epay/pkg/server/db"
	"github.com/clouway/go-epay/pkg/server/middleware"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	// Log as JSON Stackdriver with entry threading
	// instead of the default ASCII formatter.
	formatter := stackdriver.GAEStandardFormatter(
		stackdriver.WithProjectID(projectID),
	)
	log.SetFormatter(formatter)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Log the debug severity or above.
	log.SetLevel(log.DebugLevel)

	dClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	envStore := db.NewEnvironmentStore(dClient)
	cf := client.NewClientFactory(dClient)

	r := mux.NewRouter()

	skipChecks := middleware.Skip("IDN", map[string]interface{}{
		"1111111111": true,
	})
	epayAPI := middleware.EpayAPIMiddleware(envStore)

	r.Handle("/v1/pay/init", skipChecks(epayAPI(api.CheckBill(cf)))).Queries("TYPE", "CHECK")
	r.Handle("/v1/pay/init", epayAPI(api.CreatePaymentOrder(cf))).Queries("TYPE", "BILLING")
	r.Handle("/v1/pay/confirm", epayAPI(api.ConfirmPaymentOrder(cf))).Queries("TYPE", "BILLING")

	http.Handle("/", lmiddleware.XCloudTraceContext(r))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
