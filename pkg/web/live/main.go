package live

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
	"github.com/therealpaulgg/ssh-sync-common/pkg/wsutils"
	"github.com/therealpaulgg/ssh-sync-server/pkg/crypto"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/repository"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
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

type ChallengeSession struct {
	Username          string
	ChallengeAccepted chan bool
	ChallengerChannel chan *dto.PublicKeyDto
	ResponderChannel  chan []byte
	ResponderNode     string
}

type SafeChallengeResponseDict struct {
	mux  sync.Mutex
	dict map[string]ChallengeSession
}

// Utility method for safely writing to the dict
func (c *SafeChallengeResponseDict) WriteChallenge(challengePhrase string, data ChallengeSession) {
	c.mux.Lock()
	c.dict[challengePhrase] = data
	c.mux.Unlock()
}

// Utility method for safely reading from the dict
func (c *SafeChallengeResponseDict) ReadChallenge(challengePhrase string) (ChallengeSession, bool) {
	c.mux.Lock()
	data, exists := c.dict[challengePhrase]
	c.mux.Unlock()
	return data, exists
}

var ChallengeResponseChannel = make(chan ChallengeResponse)
var ChallengeResponseDict = SafeChallengeResponseDict{
	dict: make(map[string]ChallengeSession),
}

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
	bus := getChallengeBus()

	user, ok := r.Context().Value(context_keys.UserContextKey).(*models.User)
	if !ok {
		log.Warn().Msg("Could not get user from context")
		return
	}
	foo, err := wsutils.ReadClientMessage[dto.ChallengeResponseDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}
	chalChan, ok := ChallengeResponseDict.ReadChallenge(foo.Data.Challenge)
	if !ok {
		if handled := handleRemoteChallengeResponse(bus, user, foo.Data.Challenge, &conn); handled {
			return
		}
		log.Warn().Msg("Could not find challenge in dict")
		if err := wsutils.WriteServerError[dto.ChallengeSuccessEncryptedKeyDto](&conn, "Invalid challenge response."); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
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
	if key == nil {
		log.Debug().Msg("Response from challenger channel - key is nil. Exiting.")
		if err := wsutils.WriteServerError[dto.ChallengeSuccessEncryptedKeyDto](&conn, "Error responding to challenge - client abruptly closed connection."); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	}
	keys := dto.ChallengeSuccessEncryptedKeyDto{
		PublicKey:        key.PublicKey,
		EncapsulationKey: key.EncapsulationKey,
	}
	if err := wsutils.WriteServerMessage(&conn, keys); err != nil {
		log.Err(err).Msg("Error writing server message")
		return
	}
	encMasterKeyDto, err := wsutils.ReadClientMessage[dto.EncryptedMasterKeyDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}
	chalChan.ResponderChannel <- encMasterKeyDto.Data.EncryptedMasterKey
}

func handleRemoteChallengeResponse(bus *ChallengeBus, user *models.User, challengePhrase string, conn *net.Conn) bool {
	if bus == nil {
		return false
	}
	meta, err := bus.getMetadata(context.Background(), challengePhrase)
	if err != nil {
		log.Err(err).Msg("Error looking up challenge metadata in redis")
		if err := wsutils.WriteServerError[dto.ChallengeSuccessEncryptedKeyDto](conn, "Error validating challenge."); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return true
	}
	if meta == nil {
		return false
	}
	if meta.Username != user.Username {
		log.Warn().Msg("Usernames do not match for remote challenge")
		return true
	}

	wait, ok := bus.registerRemoteWait(challengePhrase)
	if !ok {
		return false
	}
	defer bus.removeRemoteWait(challengePhrase)

	if err := bus.publishAccepted(meta.Owner, challengePhrase, user.Username); err != nil {
		log.Err(err).Msg("Error publishing accepted challenge event")
		if err := wsutils.WriteServerError[dto.ChallengeSuccessEncryptedKeyDto](conn, "Error validating challenge."); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return true
	}

	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	var challengerKey *dto.PublicKeyDto
	select {
	case challengerKey = <-wait.challengerKey:
		if challengerKey == nil {
			log.Warn().Msg("Received nil challenger key for remote challenge")
			if err := wsutils.WriteServerError[dto.ChallengeSuccessEncryptedKeyDto](conn, "Error responding to challenge - client abruptly closed connection."); err != nil {
				log.Err(err).Msg("Error writing server error")
			}
			return true
		}
	case <-timer.C:
		if err := wsutils.WriteServerError[dto.ChallengeSuccessEncryptedKeyDto](conn, "Challenge timed out"); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return true
	}

	keys := dto.ChallengeSuccessEncryptedKeyDto{
		PublicKey:        challengerKey.PublicKey,
		EncapsulationKey: challengerKey.EncapsulationKey,
	}
	if err := wsutils.WriteServerMessage(conn, keys); err != nil {
		log.Err(err).Msg("Error writing server message")
		return true
	}
	encMasterKeyDto, err := wsutils.ReadClientMessage[dto.EncryptedMasterKeyDto](conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return true
	}
	if err := bus.publishEncryptedKey(meta.Owner, challengePhrase, encMasterKeyDto.Data.EncryptedMasterKey); err != nil {
		log.Err(err).Msg("Error publishing encrypted master key")
	}
	return true
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
	bus := getChallengeBus()
	// first message sent should be JSON payload
	userMachine, err := wsutils.ReadClientMessage[dto.UserMachineDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}
	userRepo := do.MustInvoke[repository.UserRepository](i)
	user, err := userRepo.GetUserByUsername(userMachine.Data.Username)
	if errors.Is(err, sql.ErrNoRows) || user == nil {
		if err := wsutils.WriteServerError[dto.MessageDto](&conn, "User not found"); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	}
	if err != nil {
		log.Err(err).Msg("Error getting user by username")
		return
	}
	machineRepo := do.MustInvoke[repository.MachineRepository](i)
	machine, err := machineRepo.GetMachineByNameAndUser(userMachine.Data.MachineName, user.ID)
	// if the machine already exists, reject
	if err == nil && machine.ID != uuid.Nil {
		if err = wsutils.WriteServerError[dto.MessageDto](&conn, "Machine already exists"); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Err(err).Msg("Error getting machine by name and user")
		return
	}
	machine = &models.Machine{}
	machine.Name = userMachine.Data.MachineName
	machine.UserID = user.ID
	// We are in an acceptable state, generate a challenge
	// The server will generate a phrase, sending it back to Computer B. The user will need to type this phrase into Computer A.
	words, err := diceware.GenerateWithWordList(3, diceware.WordListEffLarge())
	if err != nil {
		log.Err(err).Msg("Error generating diceware")
		if err := wsutils.WriteServerError[dto.MessageDto](&conn, "Error generating diceware"); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	}
	challengePhrase := strings.Join(words, "-")
	if err := wsutils.WriteServerMessage(&conn, dto.MessageDto{Message: challengePhrase}); err != nil {
		log.Err(err).Msg("Error writing challenge phrase")
		return
	}
	// The server will save this current WS connection into a map corresponding to this challenge phrase
	// Computer A starts its own connection (auth & jwt required to start this one)
	// It will send the challenge phrase to the server. Assuming it is valid, the server will send a message to Computer B to continue.
	// Computer B will then generate a pub/priv keypair, sending the public key to the server.
	// Computer A will receive the public key, decrypt the master key, encrypt the master key with the public key, and send it back to the server.
	// At this point Computer B will be able to communicate freely.

	ChallengeResponseDict.WriteChallenge(challengePhrase, ChallengeSession{
		Username:          user.Username,
		ChallengeAccepted: make(chan bool),
		ChallengerChannel: make(chan *dto.PublicKeyDto),
		ResponderChannel:  make(chan []byte),
	})
	defer func() {
		if bus != nil {
			bus.removeChallenge(context.Background(), challengePhrase)
		}
		ChallengeResponseDict.mux.Lock()
		defer ChallengeResponseDict.mux.Unlock()
		item, exists := ChallengeResponseDict.dict[challengePhrase]
		if exists {
			close(item.ChallengeAccepted)
			close(item.ChallengerChannel)
			close(item.ResponderChannel)
			delete(ChallengeResponseDict.dict, challengePhrase)
		}
	}()

	if bus != nil {
		if err := bus.registerChallenge(r.Context(), challengePhrase, user.Username, time.Minute); err != nil {
			log.Warn().Err(err).Msg("Failed to register challenge in redis")
		}
	}
	timer := time.NewTimer(30 * time.Second)
	challengeResponse := make(chan bool)
	go func() {
		var challengeAcceptedChan chan bool

		// Lock before accessing ChallengeResponseDict
		ChallengeResponseDict.mux.Lock()
		if item, exists := ChallengeResponseDict.dict[challengePhrase]; exists {
			challengeAcceptedChan = item.ChallengeAccepted
		}
		ChallengeResponseDict.mux.Unlock()

		// If the channel does not exist, return to avoid a nil channel operation
		if challengeAcceptedChan == nil {
			return
		}

		for {
			select {
			case <-timer.C:
				challengeResponse <- false
				return
			case chalWon := <-challengeAcceptedChan:
				log.Debug().Msg("Gorountine received challenge response")
				if chalWon {
					timer.Stop()
				}
				challengeResponse <- chalWon
				return
			}
		}
	}()
	challengeResult := <-challengeResponse

	if !challengeResult {
		if err := wsutils.WriteServerError[dto.MessageDto](&conn, "Challenge timed out"); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	}
	cha, ok := ChallengeResponseDict.ReadChallenge(challengePhrase)
	if !ok {
		log.Err(err).Msg("Error getting challenge from dict")
		return
	}
	if err := wsutils.WriteServerMessage(&conn, dto.MessageDto{Message: "Challenge accepted!"}); err != nil {
		log.Err(err).Msg("Error writing challenge accepted")
		return
	}
	pubkey, err := wsutils.ReadClientMessage[dto.PublicKeyDto](&conn)
	if err != nil {
		log.Err(err).Msg("Error reading client message")
		return
	}

	if _, err := crypto.ValidatePublicKey(pubkey.Data.PublicKey); err != nil {
		log.Err(err).Msg("Invalid public key format in challenge flow")
		if err := wsutils.WriteServerError[dto.MessageDto](&conn, "Invalid public key format"); err != nil {
			log.Err(err).Msg("Error writing server error")
		}
		return
	}
	if bus != nil && cha.ResponderNode != "" && cha.ResponderNode != bus.NodeID() {
		if err := bus.publishChallengerKey(cha.ResponderNode, challengePhrase, pubkey.Data); err != nil {
			log.Err(err).Msg("Error publishing challenger key to responder node")
			return
		}
	} else {
		cha.ChallengerChannel <- &pubkey.Data
	}
	encryptedMasterKey := <-cha.ResponderChannel
	machine.PublicKey = pubkey.Data.PublicKey
	if _, err = machineRepo.CreateMachine(machine); err != nil {
		log.Err(err).Msg("Error creating machine")
		return
	}
	if err := wsutils.WriteServerMessage(&conn, dto.EncryptedMasterKeyDto{EncryptedMasterKey: encryptedMasterKey}); err != nil {
		log.Err(err).Msg("Error writing encrypted master key")
		return
	}
	if err := wsutils.WriteServerMessage(&conn, dto.MessageDto{Message: "Everything is done, you can now use ssh-sync"}); err != nil {
		log.Err(err).Msg("Error writing final message")
		return
	}
}
