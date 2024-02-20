package hyperliquid

import (
	"encoding/json"
	"fmt"
	"github.com/STFX-IO/hyperliquid-go-sdk/cryptoutil"
	"testing"
	"time"
)

const Address = "0x60Cc17b782e9c5f14806663f8F617921275b9720"
const PrivateKey = "16ec09be5213c662256b65ed5d6059d3dbd5c65ab6f21e7d7878eac291ca0eb1"

var (
	baseClient  = NewApiDefault(TestnetUrl)
	info        = NewInfoApi(&baseClient)
	manager     = cryptoutil.NewPkey(PrivateKey)
	metaMap     = BuildMetaMap(info)
	exchangeApi = NewExchange(manager, info, metaMap, Address, &baseClient)
)

func TestMarketOpenAndClose(t *testing.T) {

	size := 10.0
	cloid := GetRandomCloid()

	req := OpenRequest{
		Coin:  "ARB",
		Sz:    &size,
		Cloid: &cloid,
	}

	result := exchangeApi.MarketOpen(req)
	m, _ := json.Marshal(result)

	fmt.Printf("Open Result is %s", m)

	closeReq := CloseRequest{
		Coin:  "ARB",
		Cloid: &cloid,
	}

	//wait for 2 seconds?
	time.Sleep(time.Duration(time.Duration.Seconds(2)))

	result = exchangeApi.MarketClose(closeReq)
	m, _ = json.Marshal(result)

	fmt.Printf("Close Result is %s", m)

}

func TestMarketClose(t *testing.T) {

	cloid := GetRandomCloid()

	req := CloseRequest{
		Coin:  "ARB",
		Cloid: &cloid,
	}

	result := exchangeApi.MarketClose(req)
	m, _ := json.Marshal(result)

	fmt.Printf("Result is %s", m)

}
