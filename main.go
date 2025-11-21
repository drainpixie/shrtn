package main

import (
	"github.com/rs/zerolog/log"
)

func main() {
	app := &App{}

	if err := app.init(); err != nil {
		log.Fatal().Err(err).Msg("initialization failed")
	}

	defer app.db.Close()

	if err := app.serve(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
