package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	pqc "github.com/therealpaulgg/ssh-sync-server/pkg/crypto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
)

type authClaims struct {
	Username string `json:"username"`
	Machine  string `json:"machine"`
}

func extractAuthClaims(tokenString, alg string) (username, machine string, err error) {
	switch alg {
	case "ES256", "ES512":
		token, err := jwt.ParseString(tokenString, jwt.WithVerify(false))
		if err != nil {
			return "", "", err
		}
		var ok bool
		username, ok = token.PrivateClaims()["username"].(string)
		if !ok || username == "" {
			return "", "", errors.New("missing username claim")
		}
		machine, ok = token.PrivateClaims()["machine"].(string)
		if !ok || machine == "" {
			return "", "", errors.New("missing machine claim")
		}
	case "MLDSA":
		parts := strings.SplitN(tokenString, ".", 3)
		if len(parts) != 3 {
			return "", "", errors.New("invalid JWT format")
		}
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return "", "", err
		}
		var claims authClaims
		if err := json.Unmarshal(payloadBytes, &claims); err != nil {
			return "", "", err
		}
		username, machine = claims.Username, claims.Machine
		if username == "" || machine == "" {
			return "", "", errors.New("missing username or machine claim")
		}
	default:
		return "", "", errors.New("unsupported algorithm")
	}
	return
}

func ConfigureAuth(i *do.Injector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			submatches := regexp.MustCompile(`Bearer (.*)`).FindStringSubmatch(authHeader)
			if len(submatches) < 2 || submatches[1] == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			tokenString := submatches[1]

			alg, err := pqc.DetectJWTAlgorithm(tokenString)
			if err != nil {
				log.Debug().Err(err).Msg("failed to detect JWT algorithm")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			username, machine, err := extractAuthClaims(tokenString, alg)
			if err != nil {
				log.Debug().Err(err).Msg("failed to extract JWT claims")
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

			if err := pqc.VerifyJWT(tokenString, alg, m.PublicKey); err != nil {
				log.Debug().Err(err).Msg("JWT verification failed")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), context_keys.UserContextKey, user)
			ctx = context.WithValue(ctx, context_keys.MachineContextKey, m)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
