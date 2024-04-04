package hyperliquid

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/vmihailenco/msgpack/v5"
	"math"
	"strconv"
	"time"
)

type ExchangeApi interface {
	MarketOpen(req OpenRequest) *PlaceOrderResponse
	MarketClose(req CloseRequest) *PlaceOrderResponse
	Trigger(req TriggerRequest) *PlaceOrderResponse
	Order(address string, req OrderRequest, grouping Grouping) *PlaceOrderResponse
	FindOrder(address string, cloid string) OrderResponse
	CancelOrder(address string, coin string, cloid string) CancelOrderResponse
	CancelOrderByOid(address string, coin string, oid int) CancelOrderResponse
	UpdateLeverage(req UpdateLeverageRequest) any
	GetMktPx(coin string) float64
	GetUserFills(address string) []OrderFill
}

type ExchangeImpl struct {
	infoApi    InfoApi
	cli        *API
	meta       map[string]AssetInfo
	keyManager *KeyManager
	secret     string
}

func NewExchange(cli *API, manager *KeyManager, secret string) ExchangeApi {

	infoApi := NewInfoApi(cli)
	meta := BuildMetaMap(infoApi)

	return &ExchangeImpl{
		infoApi:    infoApi,
		meta:       meta,
		cli:        cli,
		keyManager: manager,
		secret:     secret,
	}
}

func (e *ExchangeImpl) SlippagePrice(coin string, isBuy bool, slippage float64, px *float64) float64 {

	if px == nil || *px <= 0.0 {
		px = new(float64)
		parsed := e.GetMktPx(coin)
		*px = parsed
	}

	return CalculateSlippage(isBuy, px, slippage)

}

func (e *ExchangeImpl) GetMktPx(coin string) float64 {
	return e.infoApi.GetMktPx(coin)
}

func CalculateSlippage(isBuy bool, px *float64, slippage float64) float64 {

	if isBuy {
		*px = *px * (1 + slippage)
	} else {
		*px = *px * (1 - slippage)
	}

	// Format the float with a precision of 5 significant figures
	pxStr := strconv.FormatFloat(*px, 'g', 5, 64)

	// Convert the formatted string to a float
	pxFloat, err := strconv.ParseFloat(pxStr, 64)
	if err != nil {
		fmt.Println("Error parsing float:", err)
		panic("Failed to parse")
	}

	// Round the float to 6 decimal places
	return pxFloat
}

func IsBuy(szi float64) bool {
	if szi < 0 {
		return true
	} else {
		return false
	}
}

func (e *ExchangeImpl) MarketOpen(req OpenRequest) *PlaceOrderResponse {

	slippage := GetSlippage(req.Slippage)
	finalPx := e.SlippagePrice(req.Coin, req.IsBuy, slippage, req.Px)

	orderType := OrderType{
		Limit: &LimitOrderType{
			Tif: "Ioc",
		},
	}

	orderReq := OrderRequest{
		Coin:       req.Coin,
		IsBuy:      req.IsBuy,
		Sz:         *req.Sz,
		LimitPx:    finalPx,
		OrderType:  orderType,
		ReduceOnly: false,
		Cloid:      req.Cloid,
	}

	return e.Order(req.Address, orderReq, GroupingNa)

}

func (e *ExchangeImpl) MarketClose(req CloseRequest) *PlaceOrderResponse {

	positions := e.infoApi.GetUserState(req.Address).AssetPositions
	slippage := GetSlippage(req.Slippage)

	for _, position := range positions {

		item := position.Position

		if req.Coin != item.Coin {
			continue
		}

		szi, _ := strconv.ParseFloat(item.Szi, 64)
		sz := req.Sz

		if sz == nil || *sz <= 0.0 {
			sz = new(float64)
			*sz = math.Abs(szi)
		}

		isBuy := IsBuy(szi)

		finalPx := e.SlippagePrice(req.Coin, isBuy, slippage, req.Px)

		orderType := OrderType{
			Limit: &LimitOrderType{
				Tif: "Ioc",
			},
		}

		orderReq := OrderRequest{
			Coin:       req.Coin,
			IsBuy:      isBuy,
			Sz:         *sz,
			LimitPx:    finalPx,
			OrderType:  orderType,
			ReduceOnly: true,
			Cloid:      req.Cloid,
		}

		return e.Order(req.Address, orderReq, GroupingNa)

	}

	return nil
}

func (e *ExchangeImpl) Trigger(req TriggerRequest) *PlaceOrderResponse {

	slippage := GetSlippage(req.Slippage)
	positions := e.infoApi.GetUserState(req.Address).AssetPositions

	for _, position := range positions {

		item := position.Position

		if req.Coin != item.Coin {
			continue
		}

		szi, _ := strconv.ParseFloat(item.Szi, 64)

		sz := req.Sz
		if sz == nil || *sz <= 0.0 {
			sz = new(float64)
			*sz = math.Abs(szi)
		}
		isBuy := IsBuy(szi)
		finalPx := e.SlippagePrice(req.Coin, isBuy, slippage, req.Px)

		orderType := OrderType{
			Trigger: &req.Trigger,
		}

		orderReq := OrderRequest{
			Coin:       req.Coin,
			IsBuy:      isBuy,
			Sz:         0,
			LimitPx:    finalPx,
			OrderType:  orderType,
			ReduceOnly: true,
			Cloid:      req.Cloid,
		}

		return e.Order(req.Address, orderReq, GroupingTpSl)

	}

	return nil
}

func GetSlippage(sl *float64) float64 {
	slippage := DefaultSlippage

	if sl != nil {
		slippage = *sl
	}
	return slippage
}

func (e *ExchangeImpl) Order(address string, req OrderRequest, grouping Grouping) *PlaceOrderResponse {
	return e.BulkOrders(address, []OrderRequest{req}, grouping)

}

func (e *ExchangeImpl) BulkOrders(address string, requests []OrderRequest, grouping Grouping) *PlaceOrderResponse {
	var wires []OrderWire
	for _, req := range requests {
		wires = append(wires, OrderReqToWire(req, e.meta))
	}

	timestamp := GetNonce()
	action := OrderWiresToOrderAction(wires, grouping)

	v, r, s := e.SignL1Action(address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post("/exchange", payload)
	m, _ := json.Marshal(res)

	fmt.Printf("Response is %s\n", m)

	response := &PlaceOrderResponse{}

	_ = json.Unmarshal(m, &response)

	return response
}

func (e *ExchangeImpl) CancelOrder(address string, coin string, cloid string) CancelOrderResponse {
	info := e.meta[coin]
	timestamp := GetNonce()
	action := CancelCloidOrderAction{
		Type: "cancelByCloid",
		Cancels: []CancelCloidWire{
			{
				Asset: info.AssetId,
				Cloid: cloid,
			},
		},
	}

	v, r, s := e.SignL1Action(address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post("/exchange", payload)
	m, _ := json.Marshal(res)

	response := &CancelOrderResponse{}

	_ = json.Unmarshal(m, &response)

	return *response
}

func (e *ExchangeImpl) CancelOrderByOid(address string, coin string, oid int) CancelOrderResponse {
	info := e.meta[coin]
	timestamp := GetNonce()
	action := CancelOidOrderAction{
		Type: "cancel",
		Cancels: []CancelOidWire{
			{
				Asset: info.AssetId,
				Oid:   oid,
			},
		},
	}

	v, r, s := e.SignL1Action(address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post("/exchange", payload)
	fmt.Printf("Response for cancel is %s\n", res)
	m, _ := json.Marshal(res)

	response := &CancelOrderResponse{}

	_ = json.Unmarshal(m, &response)

	return *response
}

func (e *ExchangeImpl) UpdateLeverage(request UpdateLeverageRequest) any {

	timestamp := GetNonce()

	action := UpdateLeverageAction{
		Type:     "updateLeverage",
		Asset:    e.meta[request.Coin].AssetId,
		IsCross:  request.IsCross,
		Leverage: request.Leverage,
	}

	v, r, s := e.SignL1Action(request.Address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post("/exchange", payload)
	return res
}

func (e *ExchangeImpl) FindOrder(address string, cloid string) OrderResponse {
	return e.infoApi.FindOrder(address, cloid)
}

func (e *ExchangeImpl) GetUserFills(address string) []OrderFill {
	return e.infoApi.GetUserFills(address)

}

func GetNonce() int64 {
	return time.Now().UnixMilli()
}

func (e *ExchangeImpl) SignL1Action(address string, action any, timestamp int64, isMainnet bool) (byte, [32]byte, [32]byte) {
	hash := buildActionHash(action, "", timestamp)
	message := buildMessage(hash.Bytes(), isMainnet)
	return e.SignInner(address, message)

}

func (e *ExchangeImpl) SignInner(address string, message apitypes.TypedDataMessage) (byte, [32]byte, [32]byte) {

	signer := NewSigner(e.keyManager)
	req := SigRequest{
		PrimaryType: "Agent",
		DType: []apitypes.Type{
			{
				Name: "source",
				Type: "string",
			},
			{
				Name: "connectionId",
				Type: "bytes32",
			},
		},
		DTypeMsg: message,
	}

	v, r, s, err := signer.Sign(address, req)

	if err != nil {
		fmt.Printf("Error %s", err)
		panic("Failed to sign request")
	}

	return v, r, s

}

func buildActionHash(action any, vaultAd string, nonce int64) common.Hash {
	var (
		data []byte
	)

	data, err := msgpack.Marshal(action)
	if err != nil {
		panic(fmt.Sprintf("Failed to pack the data %s", err))
	}

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, uint64(nonce))
	data = ArrayAppend(data, nonceBytes)

	if vaultAd == "" {
		data = ArrayAppend(data, []byte("\x00"))
	} else {
		data = ArrayAppend(data, []byte("\x01"))
		data = ArrayAppend(data, HexToBytes(vaultAd))
	}

	result := crypto.Keccak256Hash(data)
	return result
}

func buildMessage(hash []byte, isMain bool) apitypes.TypedDataMessage {
	source := GetNetSource(isMain)
	return apitypes.TypedDataMessage{
		"source":       source,
		"connectionId": hash,
	}
}

func GetNetSource(isMain bool) string {
	if isMain {
		return "a"
	} else {
		return "b"
	}
}
