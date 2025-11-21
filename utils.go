package main

import (
	"crypto/rand"
	"encoding/base64"
	"os"
)

const version = "1.0.0"

type Config struct {
	Host     string
	Port     string
	Database string
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type URLRecord struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Clicks    int    `json:"clicks"`
	CreatedAt string `json:"created_at"`
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func generateID(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}
