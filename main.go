package main

import (
	"log"
	"net/http"

	"github.com/crypblorm/bitcoin/api-server/routers"
	"github.com/go-chi/chi"
)

func main() {
	router := routers.Routes()

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		// Walk and print out all routes
		log.Printf("Walk Callbacks: %s %s\n", method, route)
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		// panic if there is an error
		log.Panicf("Logging err: %s\n", err.Error())
	}

	// NOTE: the port usually come from the environment.
	// log.Fatal(http.ListenAndServe("0.0.0.0:28332", router))
	log.Fatal(http.ListenAndServeTLS(
		"0.0.0.0:28332",
		"/home/hyper/Workspace/bitcoin/iwallet/ssl/som32.crypblorm.com.crt",
		"/home/hyper/Workspace/bitcoin/iwallet/ssl/som32.crypblorm.com.key",
		router,
	))
}
