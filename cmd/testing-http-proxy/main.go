package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	listenAddr = flag.String("listenAddr", ":16700", "http listening address, the default value is ':16700'")
	targetURL  = flag.String("targetURL", "http://127.0.0.1:8080", "the URL of the target HTTP server")
)

func main() {
	flag.Parse()

	backend, _ := url.Parse(*targetURL)
	proxy := httputil.NewSingleHostReverseProxy(backend)

	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		director(r)
		r.Host = r.URL.Host
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(rw, r)
	})

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
