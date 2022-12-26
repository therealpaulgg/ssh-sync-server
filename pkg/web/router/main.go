package router

import (
	"fmt"

	"net/http"
	"os"
	"os/user"
	"path"

	"github.com/go-chi/chi"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/router/routes"
)

func Router(i *do.Injector) chi.Router {
	baseRouter := chi.NewRouter()
	baseRouter.Use(middleware.Log)

	apiV1Router := chi.NewRouter()
	apiV1Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	apiV1Router.Mount("/users", routes.UserRoutes(i))
	apiV1Router.Mount("/setup", routes.SetupRoutes(i))
	apiV1Router.Mount("/machines", routes.MachineRoutes(i))
	apiV1Router.Mount("/data", routes.DataRoutes(i))

	apiV1Router.Get("/token", func(w http.ResponseWriter, r *http.Request) {
		// In order for a user to successfully authenticate themselves, we can use public key cryptography.
		// The user should be able to 'authenticate' themselves by signing a piece of data submitted by the server with their private key.
		// The server will then verify the signature with the public key.
		// The signature needs to be a unique signature each time.
		// the user will generate a JWT using their private key and the server will verify it with their public key.

		// the client can generate their own tokens with their private key, and these tokens can be verified by the server.
		// the claims on the token would be something like:
		/*
			username
			machine_name
		*/
		// the server will find the key corresponding to machine name which belongs to the user, and use that key to verify the payload
		// so any request can be made with these tokens and the token can be constructed on the client.
		// However this seems questionable for performance - the server will verify the token with the public key each time.

		// an alternate method of authentication - client generates a JWT which can be validated by the server, but the server will return a new JWT which is valid for a certain amount of time.
		// this JWT is made with HMAC, and the server will have a secret key which is used to sign the JWT.

		// pros of only using RSA:
		// - the server will not need to store any secrets
		// - server does not need to generate tokens
		// cons:
		// - the server will need to verify the token with the public key each time, which is slower

	})
	apiV1Router.Get("/upload", func(w http.ResponseWriter, r *http.Request) {
		// TODO do not use filesystem, this is just for testing
		u, err := user.Current()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		file, err := os.ReadFile(path.Join(u.HomeDir, "/.ssh-sync/keypair.pub"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		key, err := jwk.ParseKey(file, jwk.WithPEM(true))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		encrypted, err := jwe.Encrypt([]byte("hello world"), jwe.WithKey(jwa.ECDH_ES_A256KW, key))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, string(encrypted))
	})
	apiV1Router.Get("/download", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "TODO")
	})
	// first time setup: a user will establish a keypair and upload it to the server
	// this is how they can authenticate to the server.
	// the user will also need a username of some sort.
	// at this point, the user will be able to upload and download from the server easily.
	// If setting up on a new computer, there should be two methods of authentication:
	// 1. The user can use the CLI on their other machine to allow access to the new machine.
	// When wanting to set up on machine B, on machine A, the user will run a command to permit access to the machine. Some basic keyword will be generated. For example:
	// Machine B: 'Please enter the keyword 'red-flying-sausage' on your other machine'
	// Machine A: ssh-sync permit-access red-flying-sausage
	// ----
	// Machine B will then receive a token which then can be used to upload a new SSH key to the server, allowing access.
	// Each device will have their own keys to authenticate to the ssh-sync-server
	baseRouter.Mount("/api/v1", apiV1Router)
	return baseRouter
}
