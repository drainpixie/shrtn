package main

import (
	"database/sql"
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
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func setup() {
	host = getEnv("HOST", "http://localhost")
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
}

func handleGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if r.URL.Path == "/" {
		title := fmt.Sprintf("shrtn @ %s", host)
		underline := strings.Repeat("=", len(title))

		lines := []string{
			title,
			underline,
			"a small url shortener",
			"",
			"get      /         index        ex: this website",
			"get      /<id>     redirect     ex: some-slug -> https://google.com",
			"post     /<url>    shorten url  ex: https://google.com -> some-slug",
			"",
			"links    0",
			"clicks   0",
			"version  0",
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Join(lines, "\n")))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("TODO"))
}

func handlePOST(w http.ResponseWriter, r *http.Request) {
	// TODO:
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

	log.Info().Str("addr", addr).Msg("server running")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
