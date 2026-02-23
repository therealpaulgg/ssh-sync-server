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
	pqc "github.com/therealpaulgg/ssh-sync-server/pkg/crypto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
)

func ConfigureAuth(i *do.Injector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			re := regexp.MustCompile(`Bearer (.*)`)
			submatches := re.FindStringSubmatch(authHeader)
			if len(submatches) < 2 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			tokenString := submatches[1]
			if tokenString == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Detect the JWT algorithm (ES256 for hybrid, ES512 for legacy)
			alg, err := pqc.DetectJWTAlgorithm(tokenString)
			if err != nil {
				log.Debug().Msg(fmt.Sprintf("Error detecting JWT algorithm: %s", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if alg != "ES256" && alg != "ES512" {
				log.Debug().Msg(fmt.Sprintf("Unsupported JWT algorithm: %s", alg))
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
			userRepo := do.MustInvoke[repository.UserRepository](i)
			user, err := userRepo.GetUserByUsername(username)
			if err != nil {
				log.Debug().Err(err).Msg("couldnt get user")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			machineRepo := do.MustInvoke[repository.MachineRepository](i)
			m, err := machineRepo.GetMachineByNameAndUser(machine, user.ID)
			if err != nil {
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
			if _, err := jwt.ParseRequest(r, jwt.WithKey(jwa.SignatureAlgorithm(alg), key)); err != nil {
				log.Debug().Msg(fmt.Sprintf("Error verifying JWT: %s", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), context_keys.UserContextKey, user)
			ctx = context.WithValue(ctx, context_keys.MachineContextKey, m)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
