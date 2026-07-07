package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fanuelson/wishlist-api/internal/wishlist"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	MaxItems        int
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		MaxItems:        wishlist.DefaultMaxItemsPerWishlist,
		ShutdownTimeout: 10 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if raw := os.Getenv("MAX_WISHLIST_ITEMS"); raw != "" {
		max, err := strconv.Atoi(raw)
		if err != nil || max <= 0 {
			return Config{}, fmt.Errorf("MAX_WISHLIST_ITEMS must be a positive integer, got %q", raw)
		}
		cfg.MaxItems = max
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
