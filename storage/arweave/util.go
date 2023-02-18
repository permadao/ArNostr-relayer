package arweave

import (
	"encoding/json"
	"fmt"
	"github.com/everFinance/arseeding/sdk"
	"github.com/everFinance/arseeding/sdk/schema"
	paySchema "github.com/everFinance/everpay-go/pay/schema"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goether"
	"github.com/nbd-wtf/go-nostr"
	"log"
	"strconv"
	"time"
)

func initSdk(b *ArweaveBackend) (*sdk.SDK, error) {
	eccSigner, err := goether.NewSigner(b.PrivateKey)
	if err != nil {
		panic(err)
	}
	sdk, err := sdk.NewSDK(b.SeedUrl, b.PayUrl, eccSigner)
	if err != nil {
		panic(err)
	}
	return sdk, nil
}
func DownLoadContentById(b *ArweaveBackend, id string) (*nostr.Event, error) {
	// sdk, err := initSdk(b)
	// if err != nil {
	// 	panic(err)
	// }
	arNode := b.GraphEndpoint
	client := goar.NewClient(arNode)
	// fmt.Printf("u:%s", id)
	data, err := client.GetTransactionDataByGateway(id)
	if err != nil {
		fmt.Printf("u:%v", err)
		return nil, err
	}
	var evt nostr.Event
	// fmt.Printf("u:%s", data)
	if err = json.Unmarshal([]byte(data), &evt); err != nil {
		log.Println(err)
		return nil, err
	}
	return &evt, nil
}
func UploadLoadEvent(b *ArweaveBackend, evt *nostr.Event) (*paySchema.Transaction, string, error) {
	sdk, err := initSdk(b)
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
		return nil, "", err
	}
	tx, itemId, err := sdk.SendDataAndPay(event, b.Currency, &schema.OptionItem{Tags: tags}, false) // your account must have enough balance in everpay
	fmt.Printf("itemId:%s", itemId)
	fmt.Printf("hash:%s", tx.HexHash())
	return tx, itemId, err
}
