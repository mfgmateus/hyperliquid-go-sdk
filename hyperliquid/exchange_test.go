package hyperliquid

import (
	"encoding/json"
	"fmt"
	"github.com/STFX-IO/hyperliquid-go-sdk/cryptoutil"
	"testing"
	"time"
)

const Address = ""
const PrivateKey = ""

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
