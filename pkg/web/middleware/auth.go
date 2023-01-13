package middleware

import (
	"fmt"
	"net/http"

	"regexp"

	"context"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
)

type UserKey string
type MachineKey string

var UserContextKey = UserKey("user")
var MachineContextKey = MachineKey("machine")

func ConfigureAuth(i *do.Injector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			re := regexp.MustCompile(`Bearer (.*)`)
			tokenString := re.FindStringSubmatch(authHeader)[1]
			if tokenString == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token, err := jwt.ParseString(tokenString, jwt.WithVerify(false))
			if err != nil {
				log.Debug().Msg(fmt.Sprintf("Error parsing JWT: %s", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			username, ok := token.PrivateClaims()["username"].(string)
			if username == "" || !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			machine, ok := token.PrivateClaims()["machine"].(string)
			if machine == "" || !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			user := &models.User{}
			user.Username = username
			if err := user.GetUserByUsername(i); err != nil {
				log.Debug().Err(err).Msg("couldnt get user")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			m := &models.Machine{}
			m.UserID = user.ID
			m.Name = machine
			if err := m.GetMachineByNameAndUser(i); err != nil {
				log.Debug().Err(err).Msg("couldnt get machine")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			key, err := jwk.ParseKey(m.PublicKey, jwk.WithPEM(true))
			if err != nil {
				log.Error().Msg(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if _, err := jwt.ParseRequest(r, jwt.WithKey(jwa.ES512, key)); err != nil {
				log.Debug().Msg(fmt.Sprintf("Error parsing JWT: %s", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, MachineContextKey, m)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Auth middleware: parse a JWT signed with ES512 and verify it with the public key
