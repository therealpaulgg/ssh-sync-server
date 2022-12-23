package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/therealpaulgg/ssh-sync-server/pkg/router"
)

func main() {
	r := router.Router()
	log.Info().Msg("Server started on port 3000")
	http.ListenAndServe(":3000", r)
}
