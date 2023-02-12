package arweave

import (
	"encoding/json"
	"fmt"
	"github.com/everFinance/arseeding/sdk"
	"github.com/everFinance/arseeding/sdk/schema"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goether"
	"github.com/nbd-wtf/go-nostr"
	"github.com/permadao/ArNostr-relayer"
	"github.com/permadao/ArNostr-relayer/utils"
	"log"
	"strconv"
	"time"
)

type ArweaveBackend struct {
	Owner         string
	PayUrl        string
	SeedUrl       string
	PrivateKey    string
	Currency      string
	GraphEndpoint string
}

type Node struct {
	Id string `json:"id"`
}
type PageInfo struct {
	HasNextPage bool `json:"hasNextPage"`
}
type Edge struct {
	Cursor string `json:"cursor"`
	Node   `json:"node"`
}
type Transaction struct {
	PageInfo `json:"pageInfo"`
	Edges    []Edge `json:"edges"`
}
type Transactions struct {
	Transaction `json:"transactions"`
}

func (b *ArweaveBackend) Init() error {
	return nil
}
func (b *ArweaveBackend) SaveEvent(evt *nostr.Event) error {

	eccSigner, err := goether.NewSigner(b.PrivateKey)
	if err != nil {
		panic(err)
	}
	sdk, err := sdk.NewSDK(b.SeedUrl, b.PayUrl, eccSigner)
	if err != nil {
		panic(err)
	}
	uploadTime := strconv.FormatInt(time.Now().UnixNano(), 10)
	eventTime := strconv.FormatInt(evt.CreatedAt.UnixNano(), 10)
	// fmt.Println(b.Owner)
	tags := []types.Tag{
		{
			Name:  "Content-Type",
			Value: "application/json",
		},
		{
			Name:  "App-Name",
			Value: "ArNostr",
		},
		{
			Name:  "App-Vesion",
			Value: "0.1",
		},
		{
			Name:  "Relay-Name",
			Value: b.Owner,
		},
		{
			Name:  "Pubkey",
			Value: evt.PubKey,
		},
		{
			Name:  "Create-Time",
			Value: uploadTime,
		},
		{
			Name:  "Event-Time",
			Value: eventTime,
		},
	}
	event, err := json.Marshal(evt)
	if err != nil {
		log.Println(err)
		return err
	}
	tx, itemId, err := sdk.SendDataAndPay(event, b.Currency, &schema.OptionItem{Tags: tags}, false) // your account must have enough balance in everpay
	fmt.Printf("itemId:%s", itemId)
	fmt.Printf("hash:%s", tx.HexHash())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (b ArweaveBackend) QueryEvents(filter *relayer.StorgeFilter) (events *relayer.QueryEvents, err error) {

	client := goar.NewClient(b.GraphEndpoint)
	after := ""
	if len(filter.Cursor) > 0 {
		after = fmt.Sprintf(`after:"%s"`, filter.Cursor)
	}
	querySql := fmt.Sprintf(`
	{
		transactions(
			first:%d,%s
			owners:["%s"]
			tags: [
					{
						name: "App-Name",
						values: "ArNostr"
					},
					{
						name:"Relay-Name",
						values:"%s",
					},
			]
			sort: HEIGHT_DESC
		) {
			pageInfo {
				hasNextPage
			  }
			edges {
				cursor
				node {
					id
				}
			}
		}
	}`, filter.PageNum, after, b.Owner, b.Owner)
	data, err := client.GraphQL(querySql)
	fmt.Printf("%s", data)
	loops := 10
	for i := 0; i < loops && err != nil; i++ {
		data, err = client.GraphQL(querySql)
		time.Sleep(time.Duration(2) * time.Second)
	}
	if err != nil {
		return nil, err
	} else {
		var transactions Transaction
		err = json.Unmarshal(data, &transactions)
		if err != nil {
			return nil, err
		}
		var queryEvents relayer.QueryEvents
		queryEvents.HasNextPage = transactions.HasNextPage
		edges := transactions.Edges

		var events []nostr.Event
		var content []byte
		for _, edge := range edges {
			id := edge.Node.Id
			if content, err = utils.DoGet(b.SeedUrl + "/" + id); err != nil {
				log.Println(err)
				return nil, err
			}
			var evt nostr.Event
			if err = json.Unmarshal(content, &evt); err != nil {
				log.Println(err)
				return nil, err
			}
			events = append(events, evt)
		}
		queryEvents.Events = events
		if len(edges) > 0 {
			queryEvents.Cursor = edges[len(edges)-1].Cursor
		}
		return &queryEvents, nil
	}
}
