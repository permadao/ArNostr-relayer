package main

import (
	"encoding/json"
	"fmt"
	schema2 "github.com/everFinance/arseeding/schema"
	"github.com/everFinance/arseeding/sdk"
	"log"
	"os"
	"strings"

	"crypto/sha256"

	"github.com/everFinance/goar/utils"
	"github.com/everFinance/goether"
	"github.com/nbd-wtf/go-nostr"
	relayer "github.com/permadao/ArNostr-relayer"
	"github.com/permadao/ArNostr-relayer/storage/arweave"
	"github.com/permadao/ArNostr-relayer/storage/postgresql"
	"github.com/spf13/viper"
)

type Relay struct {
	PostgresDatabase string
	Whitelist        []string
	Version          string
	arweaveStorge    *arweave.ArweaveBackend
	storage          *postgresql.PostgresBackend
	filterKeywords   []string
}

func (r *Relay) Name() string {
	name := viper.GetString("appname")
	pk := viper.GetString("arweave.pk")
	if len(pk) > 0 {
		s, _ := goether.NewSigner(pk)
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
	go r.arweaveStorge.ListenAndPayOrders()
	return nil
}

func (r *Relay) AcceptEvent(evt *nostr.Event) bool {
	for _, kw := range r.filterKeywords {
		if strings.Contains(evt.Content, kw) {
			return false
		}
	}
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

func (r *Relay) loadFilterKeywords(filterKeywordsFile string) (err error) {
	keywords := []string{}
	dat, err := os.ReadFile(filterKeywordsFile)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(dat), "\n") {
		word := strings.TrimSpace(line)
		if word != "" {
			keywords = append(keywords, word)
		}
	}
	r.filterKeywords = keywords
	return
}

// merge
func main() {
	// Read configs
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("read config failed: %s", err.Error()))
	}

	// relay
	r := Relay{
		PostgresDatabase: viper.GetString("postgresql_db.url"),
		Version:          viper.GetString("service.version"),

		filterKeywords: []string{},
	}
	filterKeywordsFile := viper.GetString("anti_spam.filter_keywords_file")
	if filterKeywordsFile != "" {
		err := r.loadFilterKeywords(filterKeywordsFile)
		if err != nil {
			panic(fmt.Sprintf("load filter keywords failed: %s", err.Error()))
		}
	}
	r.storage = &postgresql.PostgresBackend{DatabaseURL: r.PostgresDatabase}
	r.arweaveStorge = &arweave.ArweaveBackend{
		Owner:         viper.GetString("appname"),
		PayUrl:        viper.GetString("arweave.everpay_url"),
		SeedUrl:       viper.GetString("arweave.arseed_url"),
		PrivateKey:    viper.GetString("arweave.pk"),
		Currency:      viper.GetString("arweave.pay_currency"),
		GraphEndpoint: viper.GetString("arweave.graph_endpint"),
		ArseedOrderCh: make(chan *schema2.RespOrder, 500),
	}
	eccSigner, err := goether.NewSigner(viper.GetString("arweave.pk"))
	if err != nil {
		panic(err)
	}
	arseedSdk, err := sdk.NewSDK(viper.GetString("arweave.arseed_url"), viper.GetString("arweave.everpay_url"), eccSigner)
	if err != nil {
		panic(err)
	}
	r.arweaveStorge.ArseedSDK = arseedSdk
	if err := relayer.Start(&r); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
