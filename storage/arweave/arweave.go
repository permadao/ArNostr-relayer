package arweave

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/everFinance/arseeding/sdk"
	"github.com/everFinance/arseeding/sdk/schema"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/nbd-wtf/go-nostr"
	relayer "github.com/permadao/ArNostr-relayer"
)

type ArweaveBackend struct {
	Owner            string
	PayUrl           string
	SeedUrl          string
	PrivateKey       string
	Currency         string
	GraphEndpoint    string
	ArseedSDK        *sdk.SDK
	EventBunleItemCh chan types.BundleItem
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

type Transactions struct {
	PageInfo `json:"pageInfo"`
	Edges    []Edge `json:"edges"`
}

type QueryTransactions struct {
	Transactions `json:"transactions"`
}

func (b *ArweaveBackend) Init() error {
	return nil
}

func (b *ArweaveBackend) SaveEvent(evt *nostr.Event) (itemid string, err error) {
	i, err := UploadLoadEvent(b, evt)
	if err != nil {
		return "", err
	}
	b.EventBunleItemCh <- *i
	return i.Id, nil
}

func (b *ArweaveBackend) ListenAndUpload() {
	ticker := time.NewTicker(10 * time.Second)
	events := make([]types.BundleItem, 0, 2000)
	for {
		select {
		case i := <-b.EventBunleItemCh:
			events = append(events, i)
		case <-ticker.C:
			if len(events) > 0 && len(events) > 500 {
				bundleTags := []types.Tag{
					{Name: "Bundle-Format", Value: "binary"},
					{Name: "Bundle-Version", Value: "2.0.0"},
				}
				bundle, err := utils.NewBundle(events...)
				if err != nil {
					fmt.Printf("failed to create nested bundle; err: %v \n", err)
					continue
				}

				b.ArseedSDK.SendDataAndPay(bundle.BundleBinary, b.Currency, &schema.OptionItem{Tags: bundleTags}, false)
				// clear events
				events = make([]types.BundleItem, 0, 2000)
			}
		}
	}
}

func (b *ArweaveBackend) QueryEvents(filter *relayer.StorgeFilter) (events *relayer.QueryEvents, err error) {

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
	// fmt.Printf("%s", data)
	loops := 10
	for i := 0; i < loops && err != nil; i++ {
		data, err = client.GraphQL(querySql)
		time.Sleep(time.Duration(2) * time.Second)
	}
	if err != nil {
		return nil, err
	} else {
		var transactions QueryTransactions
		err = json.Unmarshal(data, &transactions)
		if err != nil {
			return nil, err
		}
		// fmt.Printf("data:%s", data)
		// fmt.Printf("transactions:%v", transactions)
		var queryEvents relayer.QueryEvents
		queryEvents.HasNextPage = transactions.Transactions.HasNextPage
		edges := transactions.Transactions.Edges

		var events []relayer.ArEvent
		for _, edge := range edges {
			id := edge.Node.Id
			fmt.Printf("id:%s", id)
			evt, err := DownLoadContentById(b, id)
			if err != nil {
				panic(err)
			}
			arEvent := relayer.ArEvent{
				Event:  *evt,
				ItemId: id,
			}
			events = append(events, arEvent)
		}
		queryEvents.Events = events
		if len(edges) > 0 {
			queryEvents.Cursor = edges[len(edges)-1].Cursor
		}
		// fmt.Printf("%v", queryEvents)
		return &queryEvents, nil
	}
}
