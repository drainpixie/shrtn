package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
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

func jsonOk(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func jsonCreated(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": message,
	})
}

func jsonMessage(w http.ResponseWriter, message string) {
	jsonOk(w, map[string]string{
		"message": message,
	})
}
