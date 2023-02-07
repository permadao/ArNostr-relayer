package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/nbd-wtf/go-nostr"
	"github.com/permadao/ArNostr-relayer"
	"github.com/permadao/ArNostr-relayer/storage/postgresql"
)

type Relay struct {
	PostgresDatabase string `envconfig:"POSTGRESQL_DATABASE"`
	CLNNodeId        string `envconfig:"CLN_NODE_ID"`
	CLNHost          string `envconfig:"CLN_HOST"`
	CLNRune          string `envconfig:"CLN_RUNE"`
	TicketPriceSats  int64  `envconfig:"TICKET_PRICE_SATS"`

	storage *postgresql.PostgresBackend
}

var r = &Relay{}

func (r *Relay) Name() string {
	return "ExpensiveRelay"
}

func (r *Relay) Storage() relayer.Storage {
	return r.storage
}

func (r *Relay) Init() error {
	// every hour, delete all very old events
	go func() {
		db := r.Storage().(*postgresql.PostgresBackend)

		for {
			time.Sleep(60 * time.Minute)
			db.DB.Exec(`DELETE FROM event WHERE created_at < $1`, time.Now().AddDate(0, -3, 0).Unix()) // 6 months
		}
	}()

	return nil
}

func (r *Relay) OnInitialized(s *relayer.Server) {
	// special handlers
	s.Router().Path("/").HandlerFunc(handleWebpage)
	s.Router().Path("/invoice").HandlerFunc(func(w http.ResponseWriter, rq *http.Request) {
		handleInvoice(w, rq, r)
	})
}

func (r *Relay) AcceptEvent(evt *nostr.Event) bool {
	// only accept they have a good preimage for a paid invoice for their public key
	if !checkInvoicePaidOk(evt.PubKey) {
		return false
	}

	// block events that are too large
	jsonb, _ := json.Marshal(evt)
	if len(jsonb) > 100000 {
		return false
	}

	return true
}

func (r *Relay) BeforeSave(evt *nostr.Event) {
	// do nothing
}

func (r *Relay) AfterSave(evt *nostr.Event) {
	// delete all but the 1000 most recent ones for each key
	r.Storage().(*postgresql.PostgresBackend).DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = $2 AND created_at < (
      SELECT created_at FROM event WHERE pubkey = $1
      ORDER BY created_at DESC OFFSET 1000 LIMIT 1
    )`, evt.PubKey, evt.Kind)
}

func main() {
	r := Relay{}
	if err := envconfig.Process("", &r); err != nil {
		log.Fatalf("failed to read from env: %v", err)
		return
	}
	r.PostgresDatabase = "postgres://postgres:123456@127.0.0.1:5432/nostr_relay?sslmode=disable"
	r.storage = &postgresql.PostgresBackend{DatabaseURL: r.PostgresDatabase}
	if err := relayer.Start(&r); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
