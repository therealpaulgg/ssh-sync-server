package live

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/dto"
)

func NewMachineChallenge(i *do.Injector, r *http.Request, w http.ResponseWriter) error {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}
	go NewMachineChallengeHandler(i, r, w, &conn)
	return nil
}

func NewMachineChallengeHandler(i *do.Injector, r *http.Request, w http.ResponseWriter, c *net.Conn) {
	conn := *c
	defer conn.Close()
	// first message sent should be JSON payload
	var dto dto.UserMachineDto
	msg, err := wsutil.ReadClientBinary(conn)
	if err != nil {
		log.Err(err).Msg("Error reading client binary")
		return
	}
	reader := bytes.NewReader(msg)
	err = json.NewDecoder(reader).Decode(&dto)
	if err != nil {
		log.Err(err).Msg("Error decoding JSON")
		return
	}
	user := models.User{}
	user.Username = dto.Username
	err = user.GetUserByUsername(i)
	if errors.Is(err, sql.ErrNoRows) {
		b, err := json.Marshal(ServerMessage{
			Message: "User not found",
			Error:   true,
		})
		if err != nil {
			log.Err(err).Msg("Error marshaling JSON")
			return
		}
		wsutil.WriteServerBinary(conn, b)
		return
	}
	if err != nil {
		log.Err(err).Msg("Error getting user by username")
		return
	}
	machine := models.Machine{}
	machine.Name = dto.MachineName
	machine.UserID = user.ID
	err = machine.GetMachineByNameAndUser(i)
	// if the machine already exists, reject
	if err == nil && machine.ID != uuid.Nil {
		b, err := json.Marshal(ServerMessage{
			Message: "Machine already exists",
			Error:   true,
		})
		if err != nil {
			log.Err(err).Msg("Error marshaling JSON")
			return
		}
		wsutil.WriteServerBinary(conn, b)
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Err(err).Msg("Error getting machine by name and user")
		return
	}
	// We are in an acceptable state, generate a challenge
	// The server will generate a phrase, sending it back to Computer B. The user will need to type this phrase into Computer A.
	// The server will save this current WS connection into a map corresponding to this challenge phrase
	// Computer A starts its own connection (auth & jwt required to start this one)
	// It will send the challenge phrase to the server. Assuming it is valid, the server will send a message to Computer B to continue.
	// Computer B will then generate a pub/priv keypair, sending the public key to the server.
	// Computer A will receive the public key, decrypt the master key, encrypt the master key with the public key, and send it back to the server.
	// At this point Computer B will be able to communicate freely.
	fmt.Println(dto)
}
