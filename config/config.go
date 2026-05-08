package config

import "os"

type Config struct {
	AlchemyKey  string
	DatabaseURL string
	Port        string
}

func LoadConfig() *Config {
	return &Config{
		AlchemyKey:  os.Getenv("ALCHEMY_URL"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
