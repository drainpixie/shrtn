package main

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO: Configuration through environment variables, use templating
func handleGET(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		title := "shrtn @ http://localhost:8080"
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

		w.Write([]byte(strings.Join(lines, "\n")))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("TODO"))
	}
}

func handlePOST(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read POST body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Str("body", string(body)).Msg("POST data received")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("POST received!\n"))
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
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	http.HandleFunc("/", requestHandler)
	log.Info().Msg("server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
