package main

import (
	"ci-recipe-finder-bot/config"
	"ci-recipe-finder-bot/handlers"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	config.Init()

	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/receivesms", handlers.ReceiveSMSHandler)
	log.Printf("About to listen on %s. Go to https://127.0.0.1%s/", listenAddr, listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}