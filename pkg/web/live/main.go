package live

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
)

// Computer A creates a live connection.
// Computer A sends challenge response to the server.
// This challenge is sent to the challenge response channel.
// Main worker thread looks at the challenge response channel.
// Gets the challenge, looks in dict to find a matching connection
// If a matching connection is found, notify that connection that a challenge has been accepted.
// So each connection will need its own channel.

type ChallengeResponse struct {
	Challenge       string
	ResponseChannel chan bool
}

type Something struct {
	Username          string
	ChallengeAccepted chan bool
	ChallengerChannel chan []byte
	ResponderChannel  chan []byte
}

var ChallengeResponseChannel = make(chan ChallengeResponse)
var ChallengeResponseDict = make(map[string]Something)

func MachineChallengeResponse(i *do.Injector, r *http.Request, w http.ResponseWriter) error {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}
	go MachineChallengeResponseHandler(i, r, w, &conn)
	return nil
}

func MachineChallengeResponseHandler(i *do.Injector, r *http.Request, w http.ResponseWriter, c *net.Conn) {
	conn := *c
	defer conn.Close()
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok {
		log.Warn().Msg("Could not get user from context")
		return
	}
	machine, ok := r.Context().Value(middleware.MachineContextKey).(*models.Machine)
	if !ok {
		log.Warn().Msg("Could not get machine from context")
		return
	}
	var foo dto.ChallengeResponseDto
	msg, err := wsutil.ReadClientBinary(conn)
	if err != nil {
		log.Err(err).Msg("Error reading client binary")
		return
	}
	reader := bytes.NewReader(msg)
	err = json.NewDecoder(reader).Decode(&foo)
	if err != nil {
		log.Err(err).Msg("Error decoding json")
		return
	}
	chalChan, ok := ChallengeResponseDict[foo.Challenge]
	if !ok {
		log.Warn().Msg("Could not find challenge in dict")
		return
	}
	if user.Username != chalChan.Username {
		log.Warn().Msg("Usernames do not match. Uh oh...")
		return
	}
	chalChan.ChallengeAccepted <- true

	// need a channel that both threads can access so that machine B can send public key to machine A
	// and machine A can send back encrypted master key
	key := <-chalChan.ChallengerChannel
	masterKey := models.MasterKey{
		UserID:    user.ID,
		MachineID: machine.ID,
	}
	err = masterKey.GetMasterKey(i)
	if err != nil {
		log.Err(err).Msg("Error getting master key")
		return
	}
	keys := dto.ChallengeSuccessEncryptedKeyDto{
		PublicKey:          key,
		EncryptedMasterKey: masterKey.Data,
	}
	b, err := json.Marshal(keys)
	if err != nil {
		log.Err(err).Msg("Error marshaling JSON")
		return
	}
	err = wsutil.WriteServerBinary(conn, b)
	if err != nil {
		log.Err(err).Msg("Error writing server binary")
		return
	}
	encMasterKey, err := wsutil.ReadClientBinary(conn)
	if err != nil {
		log.Err(err).Msg("Error reading client binary")
		return
	}
	chalChan.ResponderChannel <- encMasterKey
}

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
	var userMachine dto.UserMachineDto
	msg, err := wsutil.ReadClientBinary(conn)
	if err != nil {
		log.Err(err).Msg("Error reading client binary")
		return
	}
	reader := bytes.NewReader(msg)
	err = json.NewDecoder(reader).Decode(&userMachine)
	if err != nil {
		log.Err(err).Msg("Error decoding JSON")
		return
	}
	user := models.User{}
	user.Username = userMachine.Username
	err = user.GetUserByUsername(i)
	if errors.Is(err, sql.ErrNoRows) {
		b, err := json.Marshal(dto.ServerMessageDto{
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
	machine.Name = userMachine.MachineName
	machine.UserID = user.ID
	err = machine.GetMachineByNameAndUser(i)
	// if the machine already exists, reject
	if err == nil && machine.ID != uuid.Nil {
		b, err := json.Marshal(dto.ServerMessageDto{
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
	words, err := diceware.GenerateWithWordList(3, diceware.WordListEffLarge())
	if err != nil {
		log.Err(err).Msg("Error generating diceware")
		return
	}
	challengePhrase := strings.Join(words, "-")
	err = wsutil.WriteServerBinary(conn, []byte(challengePhrase))
	if err != nil {
		log.Err(err).Msg("Error writing challenge phrase")
		return
	}
	// The server will save this current WS connection into a map corresponding to this challenge phrase
	// Computer A starts its own connection (auth & jwt required to start this one)
	// It will send the challenge phrase to the server. Assuming it is valid, the server will send a message to Computer B to continue.
	// Computer B will then generate a pub/priv keypair, sending the public key to the server.
	// Computer A will receive the public key, decrypt the master key, encrypt the master key with the public key, and send it back to the server.
	// At this point Computer B will be able to communicate freely.

	ChallengeResponseDict[challengePhrase] = Something{
		Username:          user.Username,
		ChallengeAccepted: make(chan bool),
		ChallengerChannel: make(chan []byte),
		ResponderChannel:  make(chan []byte),
	}
	defer func() {
		close(ChallengeResponseDict[challengePhrase].ChallengeAccepted)
		close(ChallengeResponseDict[challengePhrase].ChallengerChannel)
		close(ChallengeResponseDict[challengePhrase].ResponderChannel)
		delete(ChallengeResponseDict, challengePhrase)
	}()
	timer := time.NewTimer(30 * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				ChallengeResponseDict[challengePhrase].ChallengeAccepted <- false
				return
			case chalWon := <-ChallengeResponseDict[challengePhrase].ChallengeAccepted:
				if chalWon {
					timer.Stop()
				}
				return
			}
		}
	}()
	challengeResult := <-ChallengeResponseDict[challengePhrase].ChallengeAccepted

	if !challengeResult {
		// TODO better error message - need to ensure client can receive it too
		return
	}
	err = wsutil.WriteServerBinary(conn, []byte("challenge-accepted"))
	if err != nil {
		log.Err(err).Msg("Error writing challenge accepted")
		return
	}
	pubkey, err := wsutil.ReadClientBinary(conn)
	if err != nil {
		log.Err(err).Msg("Error reading client binary")
		return
	}
	ChallengeResponseDict[challengePhrase].ChallengerChannel <- pubkey
	dat := <-ChallengeResponseDict[challengePhrase].ResponderChannel
	machine.PublicKey = pubkey
	err = machine.CreateMachine(i)
	if err != nil {
		log.Err(err).Msg("Error creating machine")
		return
	}
	masterKey := models.MasterKey{
		UserID:    user.ID,
		MachineID: machine.ID,
		Data:      dat,
	}
	err = masterKey.CreateMasterKey(i)
	if err != nil {
		log.Err(err).Msg("Error creating master key")
		return
	}
	wsutil.WriteServerBinary(conn, []byte("Everything is done, you can now use ssh-sync"))
}
