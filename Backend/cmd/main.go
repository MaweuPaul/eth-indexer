package main

import (
	"context"
	"log"

	"github.com/MaweuPaul/eth-indexer/config"
	"github.com/MaweuPaul/eth-indexer/internal/api"
	"github.com/MaweuPaul/eth-indexer/internal/hub"
	"github.com/MaweuPaul/eth-indexer/internal/indexer"
	"github.com/MaweuPaul/eth-indexer/internal/store"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found")
	}

	cfg := config.LoadConfig()

	db, err := store.New(cfg.DatabaseURL)

	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	log.Println("connected to database")

	// initialize websocket hub
	h := hub.NewHub()

	// start websocket broadcaster
	go h.Run()

	// contracts to index
	contracts := []string{
		"0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT
	}

	// initialize indexer
	idx, err := indexer.New(
		cfg.AlchemyKey,
		db,
		h,
		contracts,
	)

	if err != nil {
		log.Fatal("failed to create indexer:", err)
	}

	// initialize API
	a := api.New(db, h)

	// run API in background
	go func() {

		log.Println("API running on port", cfg.Port)

		err := a.Start(cfg.Port)

		if err != nil {
			log.Fatal("API error:", err)
		}
	}()

	log.Println("starting websocket hub")
	log.Println("starting indexer")

	// start indexer
	err = idx.Start(context.Background())

	if err != nil {
		log.Fatal("indexer error:", err)
	}
}
