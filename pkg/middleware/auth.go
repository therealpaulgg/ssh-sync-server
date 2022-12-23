package middleware

import (
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog/log"
)

// Auth middleware: parse a JWT signed with ES512 and verify it with the public key
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the token from the request
		// TODO do not use filesystem, this is just for testing
		u, err := user.Current()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data, err := os.ReadFile(path.Join(u.HomeDir, "/.ssh-sync/keypair.pub"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		key, err := jwk.ParseKey(data, jwk.WithPEM(true))
		if err != nil {
			log.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = jwt.ParseRequest(r, jwt.WithKey(jwa.ES512, key))
		if err != nil {
			log.Debug().Msg(fmt.Sprintf("Error parsing JWT: %s", err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// token is valid
		next.ServeHTTP(w, r)
	})
}
