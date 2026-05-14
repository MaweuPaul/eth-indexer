package main

import (
	"context"
	"log"

	"github.com/MaweuPaul/eth-indexer/config"
	"github.com/MaweuPaul/eth-indexer/internal/api"
	"github.com/MaweuPaul/eth-indexer/internal/indexer"
	"github.com/MaweuPaul/eth-indexer/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.LoadConfig()

	db, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to database!")

	contracts := []string{
		"0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT
	}

	idx, err := indexer.New(cfg.AlchemyKey, db, contracts)
	if err != nil {
		log.Fatal("Failed to create indexer:", err)
	}

	// Run API in background
	a := api.New(db)
	go func() {
		log.Println("API running on port", cfg.Port)
		if err := a.Start(cfg.Port); err != nil {
			log.Fatal("API error:", err)
		}
	}()

	log.Println("🚀 Starting indexer...")
	if err := idx.Start(context.Background()); err != nil {
		log.Fatal("Indexer error:", err)
	}
}
