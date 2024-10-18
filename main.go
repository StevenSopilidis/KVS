package main

import (
	"log"
	"net/http"

	"github.com/StevenSopilidis/kvs/db"
	"github.com/StevenSopilidis/kvs/persistance"
	"github.com/gorilla/mux"
)

func main() {
	// create file transactional logger
	logger, err := persistance.NewFileTransactionLogger("test_log.log")
	if err != nil {
		log.Fatal("could not create file_transactional_logger: ", err)
	}

	server, err := db.NewServer(logger)
	if err != nil {
		log.Fatal("could not create server: ", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/{key}", server.KeyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", server.KeyValueGetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", server.KeyValueDeleteHandler).Methods("DELETE")

	addr := "127.0.0.1:8080"

	log.Printf("Server starting listening at at address: %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
