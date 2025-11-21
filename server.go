package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	db     *sql.DB
	config Config
}

func (s *App) init() error {
	s.config = Config{
		Host:     env("HOST", "http://localhost:3002"),
		Port:     env("PORT", "3002"),
		Database: env("DATABASE", "./shrtn.db"),
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	return s.initDatabase()
}

func (s *App) serve() error {
	http.HandleFunc("/", s.handleRequest)

	addr := ":" + s.config.Port
	log.Info().Str("addr", s.config.Host).Msg("server starting")

	return http.ListenAndServe(addr, nil)
}

func (s *App) listURLs() error {
	rows, err := s.db.Query("SELECT id, url, clicks, created_at FROM urls ORDER BY created_at DESC")
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var record URLRecord
		if err := rows.Scan(&record.ID, &record.URL, &record.Clicks, &record.CreatedAt); err != nil {
			continue
		}

		if decoded, err := url.QueryUnescape(record.URL); err == nil {
			record.URL = decoded
		}

		fmt.Printf("%s %s %d %s\n", record.ID, record.URL, record.Clicks, record.CreatedAt)
	}

	return rows.Err()
}

func (s *App) getURL(id string) error {
	var targetURL string
	err := s.db.QueryRow("SELECT url FROM urls WHERE id = ?", id).Scan(&targetURL)
	if err == sql.ErrNoRows {
		fmt.Println("url not found")
		return nil
	}
	if err != nil {
		return err
	}

	if decoded, err := url.QueryUnescape(targetURL); err == nil {
		targetURL = decoded
	}

	fmt.Println(targetURL)
	return nil
}

func (s *App) deleteURL(id string) error {
	result, err := s.db.Exec("DELETE FROM urls WHERE id = ?", id)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		fmt.Println("url not found")
	} else {
		fmt.Printf("deleted %s\n", id)
	}
	return nil
}

func (s *App) shortenURL(rawURL string) error {
	if _, err := url.Parse(rawURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	id, err := s.generateUniqueID()
	if err != nil {
		return err
	}

	encodedURL := url.QueryEscape(rawURL)
	if _, err := s.db.Exec("INSERT INTO urls(id, url) VALUES(?, ?)", id, encodedURL); err != nil {
		return err
	}

	fmt.Printf("%s/%s\n", s.config.Host, id)
	return nil
}
