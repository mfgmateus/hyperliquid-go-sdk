package hyperliquid

import (
	"encoding/json"
	"fmt"
	"github.com/mfgmateus/hyperliquid-go-sdk/cryptoutil"
	"testing"
	"time"
)

const Address = "0x60Cc17b782e9c5f14806663f8F617921275b9720"
const PrivateKey = "16ec09be5213c662256b65ed5d6059d3dbd5c65ab6f21e7d7878eac291ca0eb1"

var (
	baseClient  = NewApiDefault(TestnetUrl)
	manager     = cryptoutil.NewPkey(PrivateKey)
	exchangeApi = NewExchange(manager, Address, &baseClient)
)

func TestMarketOpenAndClose(t *testing.T) {

	size := 2.0

	req := OpenRequest{
		Coin: "ARB",
		Sz:   &size,
	}

	result := exchangeApi.MarketOpen(req)
	m, _ := json.Marshal(result)

	fmt.Printf("Open Result is %s", m)

	closeReq := CloseRequest{
		Coin: "ARB",
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

func TestUpdateLeverage(t *testing.T) {

	req := UpdateLeverageRequest{
		Coin:     "ARB",
		Leverage: 5,
		IsCross:  true,
	}

	result := exchangeApi.UpdateLeverage(req)
	m, _ := json.Marshal(result)

	fmt.Printf("Result is %s", m)

}
