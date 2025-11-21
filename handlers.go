package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

func (s *App) handleIndex(w http.ResponseWriter, _ *http.Request) {
	title := fmt.Sprintf("SHRTN @ %s", s.config.Host)
	separator := strings.Repeat("=", len(title))

	var urls, clicks int
	s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&urls)
	s.db.QueryRow("SELECT COALESCE(SUM(clicks), 0) FROM urls").Scan(&clicks)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s\n%s\nENDPOINTS\n", title, separator)
	fmt.Fprintf(w, "   GET  /       This page\n")
	fmt.Fprintf(w, "   GET  /<id>   Redirect to original URL\n")
	fmt.Fprintf(w, "   POST /<url>  Shorten URL [auth]\n")
	fmt.Fprintf(w, "   POST /       Shorten URL via JSON [auth]\n\n")
	fmt.Fprintf(w, "STATISTICS\n")
	fmt.Fprintf(w, "   URLS         %d\n", urls)
	fmt.Fprintf(w, "   CLICKS       %d\n", clicks)
	fmt.Fprintf(w, "   VERSION      %s\n", version)
}

func (s *App) handleRedirect(w http.ResponseWriter, r *http.Request, id string) {
	var targetURL string
	err := s.db.QueryRow("SELECT url FROM urls WHERE id = ?", id).Scan(&targetURL)
	if err == sql.ErrNoRows {
		http.Error(w, "url not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("database query failed")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	go s.incrementClicks(id)

	if decoded, err := url.QueryUnescape(targetURL); err == nil {
		targetURL = decoded
	}

	http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
}

func (s *App) handleShorten(w http.ResponseWriter, r *http.Request) {
	if !s.isAuthorized(r) {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}

	var targetURL string

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		targetURL = req.URL
	} else {
		targetURL = strings.TrimPrefix(r.URL.Path, "/")
		if decoded, err := url.PathUnescape(targetURL); err == nil {
			targetURL = decoded
		}
	}

	if targetURL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	if _, err := url.Parse(targetURL); err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	id, err := s.generateUniqueID()
	if err != nil {
		http.Error(w, "failed to generate id", http.StatusInternalServerError)
		return
	}

	encodedURL := url.QueryEscape(targetURL)
	if _, err := s.db.Exec("INSERT INTO urls(id, url) VALUES(?, ?)", id, encodedURL); err != nil {
		log.Error().Err(err).Msg("failed to store url")
		http.Error(w, "failed to store url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%s/%s\n", s.config.Host, id)
}

func (s *App) handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote", r.RemoteAddr).
		Msg("request")

	if r.URL.Path == "/" {
		switch r.Method {
		case http.MethodGet:
			s.handleIndex(w, r)
		case http.MethodPost:
			s.handleShorten(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	switch r.Method {
	case http.MethodGet:
		s.handleRedirect(w, r, id)
	case http.MethodPost:
		s.handleShorten(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
