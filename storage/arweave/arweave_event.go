package arweave
import(
	"github.com/everFinance/goar/types"
	"time"
	"github.com/everFinance/goether"
	"github.com/everFinance/arseeding/sdk"
	"github.com/everFinance/arseeding/sdk/schema"
	"fmt"
)

type ArweaveEvent struct{
	ParentOrderId   string
	tags  []types.Tag
	UploadTime time.Time
	SeqNo uint64
	Content []byte
}

func StoreEvent(event []byte){
	priKey := "a74e92149be3560f7559ee6fae8362d4eb804cff247dea7bc14079c2fc50abee"
	eccSigner, err := goether.NewSigner(priKey)
	if err != nil {
		panic(err)
	}
	payUrl := "https://api.everpay.io"
	seedUrl := "https://arseed.web3infra.dev"
	sdk, err := sdk.NewSDK(seedUrl, payUrl, eccSigner)
	if err != nil {
		panic(err)
	}
	tags := []types.Tag{
		{
			Name:"Content-Type", 
			Value:"text",
		},
		{
         Name:"relay_name",
		 Value:"3333",
		},
	}
	tx,itemId , err := sdk.SendDataAndPay(event, "usdc", &schema.OptionItem{Tags: tags}, false) // your account must have enough balance in everpay
 
	fmt.Println("itemId:%s", itemId)
	fmt.Println("hash:%s",tx.HexHash())
}
