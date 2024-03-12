package hyperliquid

import (
	"encoding/json"
	"fmt"
	"github.com/mfgmateus/hyperliquid-go-sdk/cryptoutil"
	"strconv"
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

	size := 10.0
	cloid := GetRandomCloid()

	const coin = "ARB"
	req := OpenRequest{
		Coin:  coin,
		Sz:    &size,
		Cloid: &cloid,
	}

	result := exchangeApi.MarketOpen(req)
	m, _ := json.Marshal(result)
	fmt.Printf("Open Result is %s\n", m)

	r2 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)

	cloid = GetRandomCloid()

	closeReq := CloseRequest{
		Coin:  coin,
		Cloid: &cloid,
	}

	//wait for 2 seconds?
	time.Sleep(time.Duration(time.Duration.Seconds(2)))

	//place a take profit order

	result = exchangeApi.MarketClose(closeReq)
	fmt.Printf("%s\n", *result.GetAvgPrice())
	m, _ = json.Marshal(result)

	fmt.Printf("Close Result is %s\n", m)
	r2 = exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)
}

func TestMarketClose(t *testing.T) {

	cloid := GetRandomCloid()

	req := CloseRequest{
		Coin:  "ARB",
		Cloid: &cloid,
	}

	result := exchangeApi.MarketClose(req)
	fmt.Printf("%s\n", *result.GetAvgPrice())
	m, _ := json.Marshal(result)
	fmt.Printf("Result is %s\n", m)

	r2 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
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

func TestTrigger(t *testing.T) {

	triggerPrice := 2.10
	decimals := 4
	slippage := float64(0)
	price := float64(0)
	cloid := GetRandomCloid()

	req := TriggerRequest{
		Coin:     "ARB",
		Px:       &price,
		Slippage: &slippage,
		Trigger: TriggerOrderType{
			TriggerPx: FloatToWire(triggerPrice, &decimals),
			TpSl:      TriggerTp,
			IsMarket:  true,
		},
		Cloid: &cloid,
	}

	result := exchangeApi.Trigger(req)
	m, _ := json.Marshal(result)
	fmt.Printf("Trigger Result is %s\n", m)

	r2 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s", m)

}

func TestCancel(t *testing.T) {

	//cloid := GetRandomCloid()
	mkPrice := exchangeApi.GetMktPx("ARB")
	mkPrice = mkPrice * 1.05
	var (
		cloid  string
		result *PlaceOrderResponse
		m      any
		order  OrderResponse
	)

	cloid = GetRandomCloid()

	var req = OrderRequest{
		Coin:       "ARB",
		IsBuy:      true,
		LimitPx:    mkPrice,
		Sz:         10,
		OrderType:  OrderType{Limit: &LimitOrderType{Tif: "Gtc"}},
		Cloid:      &cloid,
		ReduceOnly: false,
	}

	result = exchangeApi.Order(req, "na")
	m, _ = json.Marshal(result)
	fmt.Printf("Order Result is %s\n", m)

	triggerPrice := mkPrice * 1.5
	decimals := 4
	slippage := float64(0)
	price := float64(0)
	cloid = GetRandomCloid()

	req2 := TriggerRequest{
		Coin:     "ARB",
		Px:       &price,
		Slippage: &slippage,
		Trigger: TriggerOrderType{
			TriggerPx: FloatToWire(triggerPrice, &decimals),
			TpSl:      TriggerTp,
			IsMarket:  true,
		},
		Cloid: &cloid,
	}

	result = exchangeApi.Trigger(req2)
	m, _ = json.Marshal(result)
	fmt.Printf("Trigger Result is %s\n", m)

	order = exchangeApi.FindOrder(Address, cloid)

	r2 := exchangeApi.CancelOrder("ARB", cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)
	fmt.Printf("Result is %s\n", strconv.FormatBool(r2.IsCancelled()))

	r2 = exchangeApi.CancelOrderByOid("ARB", int(order.Order.Order.Oid))
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)
	fmt.Printf("Result is %s\n", strconv.FormatBool(r2.IsCancelled()))
	//
	r3 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r3)
	fmt.Printf("Result is %s\n", m)

}
