package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/therealpaulgg/ssh-sync-common/pkg/dto"
)

const (
	eventTypeAccepted           = "accepted"
	eventTypeChallengerKey      = "challenger_key"
	eventTypeEncryptedMasterKey = "encrypted_master_key"
)

type challengeEvent struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge"`
	Target    string          `json:"target_node"`
	Source    string          `json:"source_node"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type challengeMetadata struct {
	Username string `json:"username"`
	Owner    string `json:"owner_node"`
}

type acceptedPayload struct {
	Username string `json:"username"`
}

type challengerKeyPayload struct {
	PublicKey        []byte `json:"public_key"`
	EncapsulationKey []byte `json:"encapsulation_key"`
}

type encryptedKeyPayload struct {
	EncryptedMasterKey []byte `json:"encrypted_master_key"`
}

type remoteWait struct {
	challengerKey chan *dto.PublicKeyDto
	encryptedKey  chan []byte
}

type ChallengeBus struct {
	client *redis.Client
	nodeID string

	ctx    context.Context
	cancel context.CancelFunc

	remoteWaits map[string]*remoteWait
	mux         sync.Mutex
}

var (
	challengeBusOnce sync.Once
	challengeBus     *ChallengeBus
)

func getChallengeBus() *ChallengeBus {
	challengeBusOnce.Do(func() {
		challengeBus = newChallengeBus()
	})
	return challengeBus
}

func newChallengeBus() *ChallengeBus {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil
	}

	opts := &redis.Options{Addr: addr}
	if pwd := os.Getenv("REDIS_PASSWORD"); pwd != "" {
		opts.Password = pwd
	}
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			opts.DB = db
		}
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Warn().Err(err).Msg("Redis unavailable; falling back to in-memory challenge coordination")
		return nil
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		if host, err := os.Hostname(); err == nil {
			nodeID = host
		} else {
			nodeID = "ssh-sync-server"
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	bus := &ChallengeBus{
		client:      client,
		nodeID:      nodeID,
		ctx:         ctx,
		cancel:      cancel,
		remoteWaits: make(map[string]*remoteWait),
	}
	bus.startListener()
	return bus
}

func (c *ChallengeBus) NodeID() string {
	if c == nil {
		return ""
	}
	return c.nodeID
}

func (c *ChallengeBus) challengeKey(challenge string) string {
	return fmt.Sprintf("challenge:session:%s", challenge)
}

func (c *ChallengeBus) channelForNode(node string) string {
	return fmt.Sprintf("challenge:events:%s", node)
}

func (c *ChallengeBus) registerChallenge(ctx context.Context, challenge, username string, ttl time.Duration) error {
	if c == nil {
		return nil
	}
	meta := challengeMetadata{
		Username: username,
		Owner:    c.nodeID,
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.challengeKey(challenge), payload, ttl).Err()
}

func (c *ChallengeBus) removeChallenge(ctx context.Context, challenge string) {
	if c == nil {
		return
	}
	if err := c.client.Del(ctx, c.challengeKey(challenge)).Err(); err != nil {
		log.Warn().Err(err).Msg("Failed to remove challenge metadata from redis")
	}
}

func (c *ChallengeBus) getMetadata(ctx context.Context, challenge string) (*challengeMetadata, error) {
	if c == nil {
		return nil, nil
	}
	data, err := c.client.Get(ctx, c.challengeKey(challenge)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	meta := &challengeMetadata{}
	if err := json.Unmarshal([]byte(data), meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (c *ChallengeBus) publishEvent(target string, event challengeEvent) error {
	if c == nil {
		return nil
	}
	event.Source = c.nodeID
	event.Target = target
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return c.client.Publish(c.ctx, c.channelForNode(target), payload).Err()
}

func (c *ChallengeBus) publishAccepted(target, challenge, username string) error {
	if c == nil {
		return nil
	}
	payload, err := json.Marshal(acceptedPayload{Username: username})
	if err != nil {
		return err
	}
	return c.publishEvent(target, challengeEvent{
		Type:      eventTypeAccepted,
		Challenge: challenge,
		Payload:   payload,
	})
}

func (c *ChallengeBus) publishChallengerKey(target, challenge string, key dto.PublicKeyDto) error {
	if c == nil {
		return nil
	}
	payload, err := json.Marshal(challengerKeyPayload{
		PublicKey:        key.PublicKey,
		EncapsulationKey: key.EncapsulationKey,
	})
	if err != nil {
		return err
	}
	return c.publishEvent(target, challengeEvent{
		Type:      eventTypeChallengerKey,
		Challenge: challenge,
		Payload:   payload,
	})
}

func (c *ChallengeBus) publishEncryptedKey(target, challenge string, key []byte) error {
	if c == nil {
		return nil
	}
	payload, err := json.Marshal(encryptedKeyPayload{EncryptedMasterKey: key})
	if err != nil {
		return err
	}
	return c.publishEvent(target, challengeEvent{
		Type:      eventTypeEncryptedMasterKey,
		Challenge: challenge,
		Payload:   payload,
	})
}

func (c *ChallengeBus) registerRemoteWait(challenge string) (*remoteWait, bool) {
	if c == nil {
		return nil, false
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	if wait, exists := c.remoteWaits[challenge]; exists {
		return wait, true
	}
	wait := &remoteWait{
		challengerKey: make(chan *dto.PublicKeyDto, 1),
		encryptedKey:  make(chan []byte, 1),
	}
	c.remoteWaits[challenge] = wait
	return wait, true
}

func (c *ChallengeBus) removeRemoteWait(challenge string) {
	if c == nil {
		return
	}
	c.mux.Lock()
	wait, exists := c.remoteWaits[challenge]
	if exists {
		delete(c.remoteWaits, challenge)
	}
	c.mux.Unlock()
	if !exists {
		return
	}
	close(wait.challengerKey)
	close(wait.encryptedKey)
}

func (c *ChallengeBus) startListener() {
	if c == nil {
		return
	}
	sub := c.client.Subscribe(c.ctx, c.channelForNode(c.nodeID))
	go func() {
		for msg := range sub.Channel() {
			event := challengeEvent{}
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Warn().Err(err).Msg("Failed to unmarshal challenge event")
				continue
			}
			c.handleEvent(event)
		}
	}()
}

func (c *ChallengeBus) handleEvent(event challengeEvent) {
	switch event.Type {
	case eventTypeAccepted:
		payload := acceptedPayload{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			log.Warn().Err(err).Msg("Failed to parse accepted payload")
			return
		}
		ChallengeResponseDict.mux.Lock()
		session, exists := ChallengeResponseDict.dict[event.Challenge]
		if exists {
			session.ResponderNode = event.Source
			ChallengeResponseDict.dict[event.Challenge] = session
		}
		ChallengeResponseDict.mux.Unlock()
		if exists {
			sendBool(session.ChallengeAccepted, true)
		}
	case eventTypeChallengerKey:
		payload := challengerKeyPayload{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			log.Warn().Err(err).Msg("Failed to parse challenger key payload")
			return
		}
		key := &dto.PublicKeyDto{PublicKey: payload.PublicKey, EncapsulationKey: payload.EncapsulationKey}

		ChallengeResponseDict.mux.Lock()
		session, exists := ChallengeResponseDict.dict[event.Challenge]
		ChallengeResponseDict.mux.Unlock()
		if exists {
			sendKey(session.ChallengerChannel, key)
			return
		}

		c.mux.Lock()
		wait, ok := c.remoteWaits[event.Challenge]
		c.mux.Unlock()
		if ok {
			sendKey(wait.challengerKey, key)
		}
	case eventTypeEncryptedMasterKey:
		payload := encryptedKeyPayload{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			log.Warn().Err(err).Msg("Failed to parse encrypted key payload")
			return
		}

		ChallengeResponseDict.mux.Lock()
		session, exists := ChallengeResponseDict.dict[event.Challenge]
		ChallengeResponseDict.mux.Unlock()
		if exists {
			sendBytes(session.ResponderChannel, payload.EncryptedMasterKey)
			return
		}

		c.mux.Lock()
		wait, ok := c.remoteWaits[event.Challenge]
		c.mux.Unlock()
		if ok {
			sendBytes(wait.encryptedKey, payload.EncryptedMasterKey)
		}
	default:
		log.Warn().Str("type", event.Type).Msg("Received unknown challenge event type")
	}
}

func sendBool(ch chan bool, val bool) {
	if ch == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	ch <- val
}

func sendKey(ch chan *dto.PublicKeyDto, key *dto.PublicKeyDto) {
	if ch == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	ch <- key
}

func sendBytes(ch chan []byte, data []byte) {
	if ch == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	ch <- data
}
