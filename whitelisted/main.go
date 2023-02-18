package main

import (
	"encoding/json"
	"log"

	"crypto/sha256"
	"github.com/everFinance/goar/utils"
	"github.com/everFinance/goether"
	"github.com/kelseyhightower/envconfig"
	"github.com/nbd-wtf/go-nostr"
	"github.com/permadao/ArNostr-relayer"
	"github.com/permadao/ArNostr-relayer/storage/arweave"
	"github.com/permadao/ArNostr-relayer/storage/postgresql"
)

type Relay struct {
	PostgresDatabase string   `envconfig:"POSTGRESQL_DATABASE"`
	Whitelist        []string `envconfig:"WHITELIST"`
	ArPrivateKey     string   `envconfig:"ARPRIVATEKEY"`
	arweaveStorge    *arweave.ArweaveBackend
	// IsEnableArstorge  bool
	storage     *postgresql.PostgresBackend
	relayConfig *relayer.RelayConfig
}

func (r *Relay) Name() string {
	// data := []byte(r.ArPrivateKey)
	// hashValue := md5.Sum(data)
	// name := fmt.Sprintf("%x", hashValue)
	name := "ArNostr"
	if len(r.ArPrivateKey) > 0 {
		s, _ := goether.NewSigner(r.ArPrivateKey)
		addr := sha256.Sum256(s.GetPublicKey())
		name = utils.Base64Encode(addr[:])
	}
	return name
}

func (r *Relay) OnInitialized(*relayer.Server) {}

func (r *Relay) Storage() relayer.Storage {
	return r.storage
}

func (r *Relay) BackupStorage() relayer.BackupStorage {
	return r.arweaveStorge
}

func (r *Relay) Init() error {
	return nil
}

func (r *Relay) AcceptEvent(evt *nostr.Event) bool {

	// block events that are too large
	jsonb, _ := json.Marshal(evt)
	if len(jsonb) > 100000 {
		return false
	}

	if len(r.Whitelist) == 0 {
		return true
	}
	// disallow anything from non-authorized pubkeys
	found := false
	for _, pubkey := range r.Whitelist {
		if pubkey == evt.PubKey {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	return true
}
func (r *Relay) RelayConfig() *relayer.RelayConfig {
	return r.relayConfig
}

func main() {
	r := Relay{}
	if err := envconfig.Process("", &r); err != nil {
		log.Fatalf("failed to read from env: %v", err)
		return
	}
	config, err := relayer.NewConfig()
	if err != nil {
		log.Fatalf("failed to read from env: %v", err)
		return
	}
	r.relayConfig = config
	r.storage = &postgresql.PostgresBackend{DatabaseURL: r.PostgresDatabase}
	r.arweaveStorge = &arweave.ArweaveBackend{
		Owner:         r.Name(),
		PayUrl:        "https://api.everpay.io",
		SeedUrl:       "https://arseed.web3infra.dev",
		PrivateKey:    r.ArPrivateKey,
		Currency:      "usdc",
		GraphEndpoint: "https://arweave.net",
	}
	if err := relayer.Start(&r); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
