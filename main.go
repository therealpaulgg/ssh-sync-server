package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/router"
)

func main() {
	injector := do.New()
	do.Provide(injector, database.NewDataAccessorService)
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
	r := router.Router(injector)
	log.Info().Msg("Server started on port 3000")
	http.ListenAndServe(":3000", r)
}
