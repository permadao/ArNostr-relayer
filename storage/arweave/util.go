package arweave

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/nbd-wtf/go-nostr"
	"github.com/spf13/viper"
)

func DownLoadContentById(b *ArweaveBackend, id string) (*nostr.Event, error) {
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
func UploadLoadEvent(b *ArweaveBackend, evt *nostr.Event) (*types.BundleItem, error) {
	uploadTime := strconv.FormatInt(time.Now().UnixNano(), 10)
	eventTime := strconv.FormatInt(evt.CreatedAt.UnixNano(), 10)
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
			Value: viper.GetString("version"),
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
	eventJs, err := json.Marshal(evt)
	if err != nil {
		return nil, err
	}
	bundleItem, err := b.ArseedSDK.ItemSigner.CreateAndSignItem(eventJs, "", "", tags)
	return &bundleItem, err
}
