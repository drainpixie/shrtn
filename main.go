// go build -ldflags "-s -w -X main.version=$(git describe --tags --always)"
package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB

	host     string
	port     string
	database string

	version string = "dev"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func generateRandomID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func setup() {
	host = getEnv("HOST", "http://localhost:3002")
	port = getEnv("PORT", "3002")
	database = getEnv("DATABASE", "./shrtn.db")

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var err error

	db, err = sql.Open("sqlite3", database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	const create = `
CREATE TABLE IF NOT EXISTS urls (
	id TEXT PRIMARY KEY,
	url TEXT NOT NULL,
	clicks INTEGER DEFAULT 0
);`

	if _, err := db.Exec(create); err != nil {
		log.Fatal().Err(err).Msg("couldn't create database")
	}
}

func handleGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if r.URL.Path == "/" {
		title := fmt.Sprintf("shrtn @ %s", host)
		underline := strings.Repeat("=", len(title))

		var links, clicks int
		db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&links)
		db.QueryRow("SELECT SUM(clicks) FROM urls").Scan(&clicks)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s\n%s\n%s\n\n", title, underline, "a small url shortener")
		fmt.Fprintf(w, "get      /         index        ex: this website\n")
		fmt.Fprintf(w, "get      /<id>     redirect     ex: some-slug -> https://google.com\n")
		fmt.Fprintf(w, "post     /<url>    shorten url  ex: https://google.com -> some-slug\n\n")
		fmt.Fprintf(w, "links    %d\nclicks   %d\nversion  %s\n", links, clicks, version)

		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")

	var target string
	err := db.QueryRow("SELECT url FROM urls WHERE id = ?", id).Scan(&target)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "url not found\n")
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "database error\n")
		return
	}

	db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE id = ?", id)
	http.Redirect(w, r, target, http.StatusFound)
}

func handlePOST(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimPrefix(r.URL.Path, "/")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "missing URL\n")
		return
	}

	var id string
	for {
		id = generateRandomID(6)

		var exists string
		if err := db.QueryRow("SELECT id FROM urls WHERE id = ?", id).Scan(&exists); err == sql.ErrNoRows {
			break
		}
	}

	_, err := db.Exec("INSERT INTO urls(id, url) VALUES(?, ?)", id, url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to store URL\n")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s/%s\n", host, id)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("incoming request")

	switch r.Method {
	case http.MethodGet:
		handleGET(w, r)
	case http.MethodPost:
		handlePOST(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	setup()

	addr := fmt.Sprintf(":%s", port)
	http.HandleFunc("/", requestHandler)

	log.Info().Str("addr", host).Msg("server running")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
