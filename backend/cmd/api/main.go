package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/api"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/config"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Bound startup work so a missing database does not hang forever.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	store := db.NewStore(dbPool)
	router := api.NewRouter(store)

	log.Printf("connected to postgres")
	log.Printf("starting api on %s", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Fatal(err)
	}
}
