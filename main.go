package main

import (
	"log"
	"os"

	"github.com/StevenSopilidis/kvs/core"
	"github.com/StevenSopilidis/kvs/frontend"
	"github.com/StevenSopilidis/kvs/persistance"
)

func main() {
	logger, err := persistance.NewTransactionLogger(os.Getenv("TLOG_TYPE"))
	if err != nil {
		log.Fatal("could not create file_transactional_logger: ", err)
	}

	core, err := core.NewKeyValueStore(logger)
	if err != nil {
		log.Fatal(err)
	}

	f, err := frontend.NewFrontend(os.Getenv("FRONTEND_TYPE"))
	if err != nil {
		log.Fatal(err)
	}

	err = f.Start(core)
	if err != nil {
		log.Fatal(err)
	}
}
