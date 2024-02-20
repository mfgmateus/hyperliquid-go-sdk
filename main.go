package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/STFX-IO/hyperliquid-go-sdk/cryptoutil"
	"github.com/STFX-IO/hyperliquid-go-sdk/hyperliquid"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"log"
	"os"
)

const DefaultSlippage = 0.05

func main() {

	baseClient := hyperliquid.NewApiDefault(hyperliquid.MainnetUrl)
	info := hyperliquid.NewInfoApi(&baseClient)
	metaMap := make(map[string]hyperliquid.AssetInfo)

	for index, asset := range info.GetMeta().Universe {
		metaMap[asset.Name] = hyperliquid.AssetInfo{
			SzDecimals: asset.SzDecimals,
			AssetId:    index,
		}
	}

	address := os.Getenv("ADDRESS")
	pkey := os.Getenv("PKEY")

	manager := cryptoutil.NewPkey(pkey)
	exchangeApi := hyperliquid.NewExchange(manager, info, metaMap, address, &baseClient)

	var px float64
	var sz float64
	var cloid = GetRandomCloid()
	result := exchangeApi.MarketClose(
		"ARB",
		&sz,
		&px,
		DefaultSlippage,
		&cloid,
	)

	m, _ := json.Marshal(result)

	fmt.Printf("Result %s", m)

}

func GetRandomCloid() string {
	buf := make([]byte, 16)
	// then we can call rand.Read.
	_, err := rand.Read(buf)
	if err != nil {
		log.Fatalf("error while generating random string: %s", err)
	}

	return hexutil.Encode(buf)
}
