package live

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
	"github.com/therealpaulgg/ssh-sync/pkg/dto"
	"github.com/therealpaulgg/ssh-sync/pkg/utils"
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
	foo, err := utils.ReadClientMessage[dto.ChallengeResponseDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}
	chalChan, ok := ChallengeResponseDict[foo.Data.Challenge]
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
	err = utils.WriteServerMessage(&conn, keys)
	if err != nil {
		log.Err(err).Msg("Error writing server message")
		return
	}
	encMasterKeyDto, err := utils.ReadClientMessage[dto.EncryptedMasterKeyDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}
	chalChan.ResponderChannel <- encMasterKeyDto.Data.EncryptedMasterKey
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
	userMachine, err := utils.ReadClientMessage[dto.UserMachineDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}
	user := models.User{}
	user.Username = userMachine.Data.Username
	err = user.GetUserByUsername(i)

	if errors.Is(err, sql.ErrNoRows) {
		err = utils.WriteServerError[dto.MessageDto](&conn, "User not found")
		if err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	}
	if err != nil {
		log.Err(err).Msg("Error getting user by username")
		return
	}
	machine := models.Machine{}
	machine.Name = userMachine.Data.MachineName
	machine.UserID = user.ID
	err = machine.GetMachineByNameAndUser(i)
	// if the machine already exists, reject
	if err == nil && machine.ID != uuid.Nil {
		err = utils.WriteServerError[dto.MessageDto](&conn, "Machine already exists")
		if err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Err(err).Msg("Error getting machine by name and user")
		return
	}
	fmt.Println("here")
	// We are in an acceptable state, generate a challenge
	// The server will generate a phrase, sending it back to Computer B. The user will need to type this phrase into Computer A.
	words, err := diceware.GenerateWithWordList(3, diceware.WordListEffLarge())
	if err != nil {
		log.Err(err).Msg("Error generating diceware")
		return
	}
	challengePhrase := strings.Join(words, "-")
	err = utils.WriteServerMessage(&conn, dto.MessageDto{Message: challengePhrase})
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
	err = utils.WriteServerMessage(&conn, dto.MessageDto{Message: "Challenge accepted!"})
	if err != nil {
		log.Err(err).Msg("Error writing challenge accepted")
		return
	}
	pubkey, err := utils.ReadClientMessage[dto.PublicKeyDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}
	ChallengeResponseDict[challengePhrase].ChallengerChannel <- pubkey.Data.PublicKey
	dat := <-ChallengeResponseDict[challengePhrase].ResponderChannel
	machine.PublicKey = pubkey.Data.PublicKey
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
	err = utils.WriteServerMessage(&conn, dto.MessageDto{Message: "Everything is done, you can now use ssh-sync"})
	if err != nil {
		log.Err(err).Msg("Error writing final message")
		return
	}
}
