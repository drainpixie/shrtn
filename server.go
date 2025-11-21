package main

import (
	"database/sql"
	"net/http"
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
