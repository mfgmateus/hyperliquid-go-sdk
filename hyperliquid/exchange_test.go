package hyperliquid

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/mfgmateus/hyperliquid-go-sdk/v2/cryptoutil"
	"github.com/stretchr/testify/require"
)

const Address = "0x60Cc17b782e9c5f14806663f8F617921275b9720"
const PrivateKey = "35e02d3d3e6f65dcc37886ab779af1c4e01d4b915a06bdacbcdb4da09497996c"

var (
	logger      = &DefaultLogger{}
	keyManager  = NewKeyManager(PrivateKey)
	baseClient  = NewApiDefault(TestnetUrl, logger)
	exchangeApi = NewExchange(&baseClient, &keyManager, logger)
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
	wired := SizeToWire(amount, szDecimals)
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
			TriggerPx: PriceToWire(triggerPrice, decimals),
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

func TestEditOrder(t *testing.T) {
	ctx := context.Background()
	coin := "KPEPE"

	cloid1 := GetRandomCloid()
	cloid2 := GetRandomCloid()
	cloid3 := GetRandomCloid()

	// create order
	exchangeApi.Order(ctx, Address, OrderRequest{
		Coin:       coin,
		IsBuy:      true,
		Sz:         3000,
		LimitPx:    0.005,
		OrderType:  OrderType{Limit: &LimitOrderType{Tif: "Alo"}},
		ReduceOnly: false,
		Cloid:      &cloid1,
	}, GroupingNa)

	order := exchangeApi.FindOrder(ctx, Address, cloid1)
	require.Equal(t, "order", order.Status)
	require.Equal(t, "open", order.Order.Status)
	require.Equal(t, "3000.0", order.Order.Order.Sz)
	require.Equal(t, "0.005", order.Order.Order.LimitPx)
	require.Equal(t, cloid1, order.Order.Order.Cloid)
	j, _ := json.Marshal(order)
	logger.LogInfo(ctx, fmt.Sprintf("initial order: %s", j))

	// modify order by Cloid
	exchangeApi.ModifyOrder(ctx, Address, ModifyOrderRequest{
		OidOrCloid: cloid1,
		Coin:       coin,
		IsBuy:      true,
		Sz:         3005,
		LimitPx:    0.0051,
		OrderType:  OrderType{Limit: &LimitOrderType{Tif: "Alo"}},
		ReduceOnly: false,
		Cloid:      &cloid2,
	})

	// new order was placed
	order = exchangeApi.FindOrder(ctx, Address, cloid2)
	require.Equal(t, "order", order.Status)
	require.Equal(t, "open", order.Order.Status)
	require.Equal(t, "3005.0", order.Order.Order.Sz)
	require.Equal(t, "0.0051", order.Order.Order.LimitPx)
	require.Equal(t, cloid2, order.Order.Order.Cloid)
	j, _ = json.Marshal(order)
	logger.LogInfo(ctx, fmt.Sprintf("modified once order: %s", j))

	// original order is canceled
	originalOrder := exchangeApi.FindOrder(ctx, Address, cloid1)
	require.Equal(t, "order", originalOrder.Status)
	require.Equal(t, "canceled", originalOrder.Order.Status)

	// modify order by Cloid
	exchangeApi.ModifyOrder(ctx, Address, ModifyOrderRequest{
		OidOrCloid: order.Order.Order.Oid,
		Coin:       coin,
		IsBuy:      true,
		Sz:         3007,
		LimitPx:    0.0052,
		OrderType:  OrderType{Limit: &LimitOrderType{Tif: "Alo"}},
		ReduceOnly: false,
		Cloid:      &cloid3,
	})

	order = exchangeApi.FindOrder(ctx, Address, cloid3)
	require.Equal(t, "order", order.Status)
	require.Equal(t, "open", order.Order.Status)
	require.Equal(t, "3007.0", order.Order.Order.Sz)
	require.Equal(t, "0.0052", order.Order.Order.LimitPx)
	require.Equal(t, cloid3, order.Order.Order.Cloid)
	j, _ = json.Marshal(order)
	logger.LogInfo(ctx, fmt.Sprintf("modified twice order: %s", j))
}

func Test_CreateOrder_SizeZero(t *testing.T) {
	ctx := context.Background()
	coin := "KPEPE"

	response := exchangeApi.Order(ctx, Address, OrderRequest{
		Coin:       coin,
		IsBuy:      true,
		Sz:         0,
		LimitPx:    0.005,
		OrderType:  OrderType{Limit: &LimitOrderType{Tif: "Alo"}},
		ReduceOnly: false,
		Cloid:      nil,
	}, GroupingNa)

	require.Equal(t, 1, len(response.Response.Data.Statuses))
	require.Equal(t, "Order has zero size.", *response.Response.Data.Statuses[0].Error)
}

func Test_CreateOrder_OuterError(t *testing.T) {
	ctx := context.Background()

	// send a payload manually, which has an invalid signature
	// this should trigger a specific error, like  "User or API Wallet 0xXXXX does not exist"
	// this error is in the response field, but it's a string
	// this tests the unmarshalPlaceOrderResponse method, which should set ResponseErr accordingly
	payload := ExchangeRequest{
		Action: PlaceOrderAction{
			Type:     "order",
			Grouping: GroupingNa,
			Orders: []OrderWire{
				{
					Asset:      0,
					IsBuy:      true,
					LimitPx:    "0.005",
					SizePx:     "0",
					ReduceOnly: false,
					OrderType:  OrderTypeWire{Limit: &LimitOrderType{Tif: "Alo"}},
				},
			},
		},
		Nonce: GetNonce(),
		Signature: RsvSignature{
			R: "0xa9a7cf2b26c9fa22b8e943f8bec93dd091b10c9d2f32e9bd98b70edaec9b908e",
			S: "0x1e2ba41a0a32e1ac23a9e63ede05afd06994f1e269b381d9edfb13c9dd4485ee",
			V: 27,
		},
	}
	response := baseClient.Post(ctx, "/exchange", payload)

	m, _ := json.Marshal(response)
	placeOrderResponse, err := unmarshalPlaceOrderResponse(m)
	require.NoError(t, err)
	require.Nil(t, placeOrderResponse.Response)
	require.Contains(t, *placeOrderResponse.ResponseErr, "L1 error: User or API Wallet")
}

func Test_Cancel_OuterError(t *testing.T) {
	ctx := context.Background()

	// send a payload manually, which has an invalid signature
	// this should trigger a specific error, like  "User or API Wallet 0xXXXX does not exist"
	// this error is in the response field, but it's a string
	// this tests the unmarshalPlaceOrderResponse method, which should set ResponseErr accordingly
	payload := ExchangeRequest{
		Action: CancelCloidOrderAction{
			Type: "cancelByCloid",
			Cancels: []CancelCloidWire{
				{
					Asset: 0,
					Cloid: "0x9b0044eced0ed61211de2b46d964e874",
				},
			},
		},
		Nonce: GetNonce(),
		Signature: RsvSignature{
			R: "0xa9a7cf2b26c9fa22b8e943f8bec93dd091b10c9d2f32e9bd98b70edaec9b908e",
			S: "0x1e2ba41a0a32e1ac23a9e63ede05afd06994f1e269b381d9edfb13c9dd4485ee",
			V: 27,
		},
	}
	response := baseClient.Post(ctx, "/exchange", payload)

	m, _ := json.Marshal(response)
	placeOrderResponse, err := unmarshalCancelOrderResponse(m)
	require.NoError(t, err)
	require.Nil(t, placeOrderResponse.Response)
	require.Contains(t, *placeOrderResponse.ResponseErr, "L1 error: User or API Wallet")
}

func TestSizeAndPriceToWire(t *testing.T) {
	// Simulate require.Equal, to not add a dependency just for this

	// for ETH
	require.Equal(t, "1234.5", PriceToWire(1234.56, 4))

	// for MEW
	require.Equal(t, "0.00123", PriceToWire(0.00123, 0))
	require.Equal(t, "0.001234", PriceToWire(0.001234, 0))
	require.Equal(t, "0.001234", PriceToWire(0.0012345, 0))

	// for ETH
	require.Equal(t, "16.27", SizeToWire(16.27, 4))
	require.Equal(t, "16.2755", SizeToWire(16.2755, 4))
	require.Equal(t, "16.2755", SizeToWire(16.2755002, 4))

	// for MEW
	require.Equal(t, "2840522", SizeToWire(2840522, 0))
	require.Equal(t, "2840522", SizeToWire(2840522.1, 0))
}
