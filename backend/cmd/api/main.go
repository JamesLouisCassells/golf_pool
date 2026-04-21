package main

import (
	"log"
	"net/http"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/api"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/config"
)

func main() {
	cfg := config.Load()
	router := api.NewRouter()

	log.Printf("starting api on %s", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Fatal(err)
	}
}
