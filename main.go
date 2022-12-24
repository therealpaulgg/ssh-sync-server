package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/router"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
	err = database.DataAccessorInstance.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to database")
	}
	r := router.Router()
	log.Info().Msg("Server started on port 3000")
	http.ListenAndServe(":3000", r)
}
