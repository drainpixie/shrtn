package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

func (s *App) handleIndex(w http.ResponseWriter, _ *http.Request) {
	var urls, clicks int

	s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&urls)
	s.db.QueryRow("SELECT COALESCE(SUM(clicks), 0) FROM urls").Scan(&clicks)

	jsonOk(w, map[string]any{
		"urls":    urls,
		"clicks":  clicks,
		"version": version,
	})
}

func (s *App) handleRedirect(w http.ResponseWriter, r *http.Request, id string) {
	var target string
	err := s.db.QueryRow("SELECT url FROM urls WHERE id = ?", id).Scan(&target)

	if err == sql.ErrNoRows {
		jsonError(w, http.StatusNotFound, "url not found")
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("database query failed")
		jsonError(w, http.StatusInternalServerError, "database query failed")
		return
	}

	go s.incrementClicks(id)

	if decoded, err := url.QueryUnescape(target); err == nil {
		target = decoded
	}

	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

func (s *App) handleShorten(w http.ResponseWriter, r *http.Request) {
	if !s.isAuthorized(r) {
		jsonError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	var target string

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req ShortenRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, http.StatusBadRequest, "invalid json")
			return
		}

		target = req.URL
	} else {
		target = strings.TrimPrefix(r.URL.Path, "/")

		if decoded, err := url.PathUnescape(target); err == nil {
			target = decoded
		}
	}

	if target == "" {
		jsonError(w, http.StatusBadRequest, "url required")
		return
	}

	if _, err := url.Parse(target); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid url")
		return
	}

	id, err := s.generateUniqueID()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to generate id")
		return
	}

	encodedURL := url.QueryEscape(target)
	if _, err := s.db.Exec("INSERT INTO urls(id, url) VALUES(?, ?)", id, encodedURL); err != nil {
		log.Error().Err(err).Msg("failed to store url")
		jsonError(w, http.StatusInternalServerError, "failed to store url")
		return
	}

	jsonOk(w, map[string]string{
		"id":  id,
		"url": fmt.Sprintf("%s/%s", s.config.Host, id),
	})
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
