package main

import (
	"encoding/json"
	"fmt"
	"github.com/STFX-IO/hyperliquid-go-sdk/cryptoutil"
	"github.com/STFX-IO/hyperliquid-go-sdk/hyperliquid"
	"os"
)

func main() {

	var (
		baseClient = hyperliquid.NewApiDefault(hyperliquid.MainnetUrl)
		info       = hyperliquid.NewInfoApi(&baseClient)
		address    = os.Getenv("ADDRESS")
		pkey       = os.Getenv("PKEY")
		manager    = cryptoutil.NewPkey(pkey)
	)

	metaMap := hyperliquid.BuildMetaMap(info)
	exchangeApi := hyperliquid.NewExchange(manager, info, metaMap, address, &baseClient)

	var cloid = hyperliquid.GetRandomCloid()

	req := hyperliquid.Request{
		Coin:  "ARB",
		Cloid: &cloid,
	}

	result := exchangeApi.MarketClose(req)

	m, _ := json.Marshal(result)

	fmt.Printf("Result %s", m)

}
