package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/internal/setup"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/router"
)

func main() {
	injector := do.New()
	setup.SetupServices(injector)
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
	r := router.Router(injector)
	log.Info().Msg("Server started on port 3000")
	http.ListenAndServe(":3000", r)
}
