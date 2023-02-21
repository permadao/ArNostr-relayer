package arweave

import (
	"encoding/json"
	"fmt"
	seedSchema "github.com/everFinance/arseeding/schema"
	"github.com/everFinance/arseeding/sdk"
	"time"

	"github.com/everFinance/goar"
	"github.com/nbd-wtf/go-nostr"
	relayer "github.com/permadao/ArNostr-relayer"
)

type ArweaveBackend struct {
	Owner         string
	PayUrl        string
	SeedUrl       string
	PrivateKey    string
	Currency      string
	GraphEndpoint string
	ArseedSDK     *sdk.SDK
	ArseedOrderCh chan *seedSchema.RespOrder
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
	ord, err := UploadLoadEvent(b, evt)
	if err != nil {
		return "", err
	}
	b.ArseedOrderCh <- ord
	return ord.ItemId, nil
}

func (b *ArweaveBackend) ListenAndPayOrders() {
	ticker := time.NewTicker(5 * time.Second)
	ords := make([]*seedSchema.RespOrder, 0, 500)
	for {
		select {
		case ord := <-b.ArseedOrderCh:
			ords = append(ords, ord)
		case <-ticker.C:
			if len(ords) > 0 {
				_, err := b.ArseedSDK.BatchPayOrders(ords)
				if err != nil {
					fmt.Printf("b.ArseedSDK.BatchPayOrders(ords); err: %v \n", err)
					continue
				}
				// clear ords
				ords = make([]*seedSchema.RespOrder, 0, 500)
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

		var events []nostr.Event
		for _, edge := range edges {
			id := edge.Node.Id
			evt, err := DownLoadContentById(b, id)
			if err != nil {
				panic(err)
			}
			events = append(events, *evt)
		}
		queryEvents.Events = events
		if len(edges) > 0 {
			queryEvents.Cursor = edges[len(edges)-1].Cursor
		}
		// fmt.Printf("%v", queryEvents)
		return &queryEvents, nil
	}
}
