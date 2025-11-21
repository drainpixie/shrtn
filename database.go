package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

func (s *App) initDatabase() error {
	var err error
	s.db, err = sql.Open("sqlite3", s.config.Database+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS urls (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			clicks INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS auth_keys (
			key TEXT PRIMARY KEY,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at)`,
	}

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	}

	return nil
}

func (s *App) generateUniqueID() (string, error) {
	for range make([]struct{}, 100) {
		id := generateID(6)
		if id == "" {
			continue
		}

		var exists string
		err := s.db.QueryRow("SELECT id FROM urls WHERE id = ?", id).Scan(&exists)

		if err == sql.ErrNoRows {
			return id, nil
		}

		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("failed to generate unique id after 100 attempts")
}

func (s *App) incrementClicks(id string) {
	if _, err := s.db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE id = ?", id); err != nil {
		log.Error().Err(err).Str("id", id).Msg("failed to increment clicks")
	}
}
