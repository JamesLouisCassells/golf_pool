package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/api"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/auth"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/config"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/golf"
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
	golfProvider := golf.NewProvider(golf.ProviderConfig{
		Provider: cfg.GolfProvider,
		BaseURL:  cfg.GolfAPIBaseURL,
		APIKey:   cfg.GolfAPIKey,
		APIHost:  cfg.GolfAPIHost,
	})
	authMiddleware := auth.NewMiddleware(store, auth.Config{
		MockEnabled:       cfg.MockAuthEnabled,
		MockClerkID:       cfg.MockAuthClerkID,
		MockEmail:         cfg.MockAuthEmail,
		MockName:          cfg.MockAuthName,
		MockAdmin:         cfg.MockAuthAdmin,
		JWKSURL:           cfg.ClerkJWKSURL,
		Issuer:            cfg.ClerkIssuer,
		Audience:          cfg.ClerkAudience,
		SecretKey:         cfg.ClerkSecretKey,
		AuthorizedParties: cfg.ClerkAuthorizedParties,
		EmailClaim:        cfg.ClerkEmailClaim,
		NameClaim:         cfg.ClerkNameClaim,
		AdminClaim:        cfg.AdminClaim,
		AdminValue:        cfg.AdminValue,
	})
	router := api.NewRouter(store, authMiddleware, golfProvider)

	log.Printf("connected to postgres")
	if cfg.MockAuthEnabled {
		log.Printf("mock auth enabled for local development as %s", cfg.MockAuthEmail)
	}
	if cfg.ClerkJWKSURL == "" {
		log.Printf("auth middleware is unconfigured: protected routes will return 503 until CLERK_JWKS_URL is set")
	}
	if golfProvider == nil {
		log.Printf("golf provider refresh is unconfigured: /api/admin/refresh will only accept manual results payloads")
	}
	log.Printf("starting api on %s", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Fatal(err)
	}
}
