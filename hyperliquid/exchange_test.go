package hyperliquid

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/mfgmateus/hyperliquid-go-sdk/v2/cryptoutil"
	"strconv"
	"testing"
	"time"
)

const Address = "0x60Cc17b782e9c5f14806663f8F617921275b9720"
const PrivateKey = "35e02d3d3e6f65dcc37886ab779af1c4e01d4b915a06bdacbcdb4da09497996c"

var (
	keyManager  = NewKeyManager(PrivateKey)
	baseClient  = NewApiDefault(MainnetUrl, &DefaultLogger{})
	exchangeApi = NewExchange(&baseClient, &keyManager, &DefaultLogger{})
	infoApi     = NewInfoApi(&baseClient)
)

type SingleKeyManager struct {
	privKey *ecdsa.PrivateKey
}

func (m SingleKeyManager) GetKey(address string) *ecdsa.PrivateKey {
	return m.privKey
}

func NewKeyManager(privKey string) KeyManager {
	manager := cryptoutil.NewPkey(privKey)
	fmt.Printf("%s\n", manager.PublicAddress())
	return &SingleKeyManager{privKey: manager.PrivateECDSA()}
}

func TestMarketOpenAndClose(t *testing.T) {

	size := 7000.0
	cloid := GetRandomCloid()
	//
	const coin = "KPEPE"
	req := OpenRequest{
		Address: Address,
		Coin:    coin,
		Sz:      &size,
		Cloid:   &cloid,
		IsBuy:   true,
	}

	result := exchangeApi.MarketOpen(context.Background(), req)
	m, _ := json.Marshal(result)
	fmt.Printf("Open Result is %s\n", m)

	r2 := exchangeApi.FindOrder(context.Background(), Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)

	cloid = GetRandomCloid()

	closeReq := CloseRequest{
		Address: Address,
		Coin:    coin,
		Cloid:   &cloid,
	}

	////wait for 2 seconds?
	time.Sleep(time.Duration(time.Duration.Seconds(2)))

	result = exchangeApi.MarketClose(context.Background(), closeReq)
	fmt.Printf("%s\n", *result.GetAvgPrice())
	m, _ = json.Marshal(result)

	fmt.Printf("Close Result is %s\n", m)
}

func TestMarketClose(t *testing.T) {

	cloid := GetRandomCloid()

	req := CloseRequest{
		Coin:  "ARB",
		Cloid: &cloid,
	}

	result := exchangeApi.MarketClose(context.Background(), req)
	fmt.Printf("%s\n", *result.GetAvgPrice())
	m, _ := json.Marshal(result)
	fmt.Printf("Result is %s\n", m)

	r2 := exchangeApi.FindOrder(context.Background(), Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s", m)

}

func TestAccountInfo(t *testing.T) {

	state := infoApi.GetUserState(context.Background(), Address)

	m, _ := json.Marshal(state)
	fmt.Printf("Result is %s\n", m)

	amountF, _ := strconv.ParseFloat(state.Withdrawable, 64)
	amount := ConvertTo2Decimals(amountF)
	szDecimals := 2
	wired := FloatToWire(amount, &szDecimals)
	fmt.Printf("Wired is %s\n", wired)
	fmt.Printf("Amount is %.6f\n", amountF)
}

func TestUpdateLeverage(t *testing.T) {

	req := UpdateLeverageRequest{
		Coin:     "ARB",
		Leverage: 5,
		IsCross:  false,
	}

	result := exchangeApi.UpdateLeverage(context.Background(), req)
	m, _ := json.Marshal(result)

	fmt.Printf("Result is %s\n", m)

}

func TestGetUserFills(t *testing.T) {

	fills := exchangeApi.GetUserFills(context.Background(), Address)
	m, _ := json.Marshal(fills)
	fmt.Printf("Result is %s\n", m)

	m0, _ := json.Marshal(fills[0])
	fmt.Printf("Result is %s\n", m0)

}

func TestTrigger(t *testing.T) {

	triggerPrice := 3090.10
	decimals := 4
	slippage := float64(0)
	price := float64(0)
	cloid := GetRandomCloid()

	req := TriggerRequest{
		Address:  Address,
		Coin:     "ETH",
		Px:       &price,
		Slippage: &slippage,
		Trigger: TriggerOrderType{
			TriggerPx: FloatToWire(triggerPrice, &decimals),
			TpSl:      TriggerTp,
			IsMarket:  true,
		},
		Cloid: &cloid,
	}

	result := exchangeApi.Trigger(context.Background(), req)
	m, _ := json.Marshal(result)
	fmt.Printf("Trigger Result is %s\n", m)

	r2 := exchangeApi.FindOrder(context.Background(), Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s", m)

}

func TestWithdraw(t *testing.T) {

	state := infoApi.GetUserState(context.Background(), Address)

	m, _ := json.Marshal(state)
	fmt.Printf("Result is %s\n", m)

	req := WithdrawRequest{
		Address:     Address,
		Destination: Address,
		Amount:      2,
	}

	res := exchangeApi.Withdraw(context.Background(), req)
	m, _ = json.Marshal(res)
	fmt.Printf("Result is %s\n", m)

}

func TestCancel(t *testing.T) {

	//cloid := GetRandomCloid()
	//mkPrice := exchangeApi.GetMktPx("ARB")
	//mkPrice = mkPrice * 1.05
	//var (
	//	cloid  string
	//	result *PlaceOrderResponse
	//	m      any
	//	order  OrderResponse
	//)
	//
	//cloid = GetRandomCloid()
	//
	//var req = OrderRequest{
	//	Coin:       "ARB",
	//	IsBuy:      true,
	//	LimitPx:    mkPrice,
	//	Sz:         10,
	//	OrderType:  OrderType{Limit: &LimitOrderType{Tif: "Gtc"}},
	//	Cloid:      &cloid,
	//	ReduceOnly: false,
	//}
	//
	//result = exchangeApi.Order(Address, req, "na")
	//m, _ = json.Marshal(result)
	//fmt.Printf("Order Result is %s\n", m)
	//
	//triggerPrice := mkPrice * 1.5
	//decimals := 4
	//slippage := float64(0)
	//price := float64(0)
	//cloid = GetRandomCloid()
	//
	//req2 := TriggerRequest{
	//	Coin:     "ARB",
	//	Px:       &price,
	//	Slippage: &slippage,
	//	Trigger: TriggerOrderType{
	//		TriggerPx: FloatToWire(triggerPrice, &decimals),
	//		TpSl:      TriggerTp,
	//		IsMarket:  true,
	//	},
	//	Cloid: &cloid,
	//}
	//
	//result = exchangeApi.Trigger(req2)
	//m, _ = json.Marshal(result)
	//fmt.Printf("Trigger Result is %s\n", m)

	//order := exchangeApi.FindOrder(Address, cloid)

	//r2 := exchangeApi.CancelOrder(Address, "POPCAT", cloid)
	//m, _ := json.Marshal(r2)
	//fmt.Printf("Result is %s\n", m)
	//fmt.Printf("Result is %s\n", strconv.FormatBool(r2.IsCancelled()))

	//r2 = exchangeApi.CancelOrderByOid(Address, "ARB", int(order.Order.Order.Oid))
	//m, _ = json.Marshal(r2)
	//fmt.Printf("Result is %s\n", m)
	//fmt.Printf("Result is %s\n", strconv.FormatBool(r2.IsCancelled()))
	////
	//r3 := exchangeApi.FindOrder(Address, cloid)
	//m, _ = json.Marshal(r3)
	//fmt.Printf("Result is %s\n", m)

}

func TestFloatToWire(t *testing.T) {

	f := 95.23567
	s := FloatToWire2(f, 5)

	fmt.Printf("%s\n", s)

}
