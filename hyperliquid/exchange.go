package hyperliquid

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/vmihailenco/msgpack/v5"
)

type ExchangeApi interface {
	MarketOpen(context context.Context, req OpenRequest) *PlaceOrderResponse
	MarketClose(context context.Context, req CloseRequest) *PlaceOrderResponse
	Trigger(context context.Context, req TriggerRequest) *PlaceOrderResponse
	Order(context context.Context, address string, req OrderRequest, grouping Grouping) *PlaceOrderResponse
	FindOrder(context context.Context, address string, cloid string) OrderResponse
	CancelOrder(context context.Context, address string, coin string, cloid string) *CancelOrderResponse
	CancelOrderByOid(context context.Context, address string, coin string, oid int64) *CancelOrderResponse
	ModifyOrder(ctx context.Context, address string, request ModifyOrderRequest) *ModifyOrderResponse
	UpdateLeverage(context context.Context, req UpdateLeverageRequest) any
	GetMktPx(context context.Context, coin string) float64
	GetUserFills(context context.Context, address string) []OrderFill
	Withdraw(context context.Context, request WithdrawRequest) *WithdrawResponse
}

type ExchangeImpl struct {
	infoApi    InfoApi
	cli        *API
	meta       map[string]AssetInfo
	keyManager *KeyManager
	logger     Logger
}

func NewExchange(cli *API, manager *KeyManager, logger Logger) ExchangeApi {

	infoApi := NewInfoApi(cli)
	meta := BuildMetaMap(infoApi)

	return &ExchangeImpl{
		infoApi:    infoApi,
		meta:       meta,
		cli:        cli,
		keyManager: manager,
		logger:     logger,
	}
}

func (e *ExchangeImpl) SlippagePrice(ctx context.Context, coin string, isBuy bool, slippage float64, px *float64) float64 {

	if px == nil || *px <= 0.0 {
		px = new(float64)
		parsed := e.GetMktPx(ctx, coin)
		*px = parsed
	}

	return e.CalculateSlippage(ctx, isBuy, px, slippage)

}

func (e *ExchangeImpl) GetMktPx(ctx context.Context, coin string) float64 {
	return e.infoApi.GetMktPx(ctx, coin)
}

func (e *ExchangeImpl) CalculateSlippage(ctx context.Context, isBuy bool, px *float64, slippage float64) float64 {

	if isBuy {
		*px = *px * (1 + slippage)
	} else {
		*px = *px * (1 - slippage)
	}

	// Format the float with a precision of 6 significant figures
	pxStr := strconv.FormatFloat(*px, 'g', 5, 64)

	// Convert the formatted string to a float
	pxFloat, err := strconv.ParseFloat(pxStr, 64)
	if err != nil {
		e.logger.LogErr(ctx, "Error parsing float", err)
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

func (e *ExchangeImpl) MarketOpen(ctx context.Context, req OpenRequest) *PlaceOrderResponse {

	slippage := GetSlippage(req.Slippage)
	finalPx := e.SlippagePrice(ctx, req.Coin, req.IsBuy, slippage, req.Px)

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

	return e.Order(ctx, req.Address, orderReq, GroupingNa)

}

func (e *ExchangeImpl) MarketClose(ctx context.Context, req CloseRequest) *PlaceOrderResponse {

	positions := e.infoApi.GetUserState(ctx, req.Address).AssetPositions
	slippage := GetSlippage(req.Slippage)

	for _, position := range positions {

		item := position.Position

		if strings.ToUpper(req.Coin) != strings.ToUpper(item.Coin) {
			continue
		}

		szi, _ := strconv.ParseFloat(item.Szi, 64)
		sz := req.Sz

		if sz == nil || *sz <= 0.0 {
			sz = new(float64)
			*sz = math.Abs(szi)
		}

		isBuy := IsBuy(szi)

		finalPx := e.SlippagePrice(ctx, req.Coin, isBuy, slippage, req.Px)

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

		return e.Order(ctx, req.Address, orderReq, GroupingNa)

	}

	err := fmt.Sprintf("No position found for asset %s", req.Coin)
	return buildFailedResponse(err)
}

func buildFailedResponse(err string) *PlaceOrderResponse {
	return &PlaceOrderResponse{
		Status:      "err",
		ResponseErr: &err,
	}
}

func (e *ExchangeImpl) Trigger(ctx context.Context, req TriggerRequest) *PlaceOrderResponse {

	slippage := GetSlippage(req.Slippage)
	positions := e.infoApi.GetUserState(ctx, req.Address).AssetPositions

	for _, position := range positions {

		item := position.Position

		if strings.ToUpper(req.Coin) != strings.ToUpper(item.Coin) {
			continue
		}

		szi, _ := strconv.ParseFloat(item.Szi, 64)

		sz := req.Sz
		if sz == nil || *sz <= 0.0 {
			sz = new(float64)
			*sz = math.Abs(szi)
		}
		isBuy := IsBuy(szi)
		finalPx := e.SlippagePrice(ctx, req.Coin, isBuy, slippage, req.Px)

		orderType := OrderType{
			Trigger: &req.Trigger,
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

		return e.Order(ctx, req.Address, orderReq, GroupingTpSl)

	}

	err := fmt.Sprintf("No position found for asset %s", req.Coin)
	return buildFailedResponse(err)
}

func GetSlippage(sl *float64) float64 {
	slippage := DefaultSlippage

	if sl != nil {
		slippage = *sl
	}
	return slippage
}

func (e *ExchangeImpl) Order(context context.Context, address string, req OrderRequest, grouping Grouping) *PlaceOrderResponse {
	return e.BulkOrders(context, address, []OrderRequest{req}, grouping)

}

func (e *ExchangeImpl) BulkOrders(ctx context.Context, address string, requests []OrderRequest, grouping Grouping) *PlaceOrderResponse {
	var wires []OrderWire
	for _, req := range requests {
		wires = append(wires, OrderReqToWire(req, e.meta))
	}

	timestamp := GetNonce()
	action := OrderWiresToOrderAction(wires, grouping)

	v, r, s := e.SignL1Action(ctx, address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post(ctx, "/exchange", payload)
	m, _ := json.Marshal(res)

	e.logger.LogInfo(ctx, fmt.Sprintf("Response is %s", res))

	response, err := unmarshalPlaceOrderResponse(m)
	if err != nil {
		e.logger.LogErr(ctx, "failed to unmarshalPlaceOrderResponse", err)
		return nil
	}

	return response
}

func (e *ExchangeImpl) ModifyOrder(ctx context.Context, address string, request ModifyOrderRequest) *ModifyOrderResponse {
	return e.BulkModify(ctx, address, []ModifyOrderRequest{request})
}

func (e *ExchangeImpl) BulkModify(ctx context.Context, address string, requests []ModifyOrderRequest) *ModifyOrderResponse {
	var wires []ModifyOrderWire
	for _, req := range requests {
		wires = append(wires, ModifyOrderReqToWire(req, e.meta))
	}

	timestamp := GetNonce()
	action := ModifyOrderWiresToModifyOrderAction(wires)

	v, r, s := e.SignL1Action(ctx, address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}
	res := (*e.cli).Post(ctx, "/exchange", payload)
	m, _ := json.Marshal(res)

	e.logger.LogInfo(ctx, fmt.Sprintf("response is %s", res))

	response, err := unmarshalModifyOrderResponse(m)
	if err != nil {
		e.logger.LogErr(ctx, "failed to unmarshalModifyOrderResponse", err)
		return nil
	}

	return response
}

func (e *ExchangeImpl) CancelOrder(ctx context.Context, address string, coin string, cloid string) *CancelOrderResponse {
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

	v, r, s := e.SignL1Action(ctx, address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}
	res := (*e.cli).Post(ctx, "/exchange", payload)
	m, _ := json.Marshal(res)

	e.logger.LogInfo(ctx, fmt.Sprintf("response is %s", res))

	response, err := unmarshalCancelOrderResponse(m)
	if err != nil {
		e.logger.LogErr(ctx, "failed to unmarshalModifyOrderResponse", err)
		return nil
	}

	return response
}

func (e *ExchangeImpl) CancelOrderByOid(ctx context.Context, address string, coin string, oid int64) *CancelOrderResponse {
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

	v, r, s := e.SignL1Action(ctx, address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post(ctx, "/exchange", payload)
	m, _ := json.Marshal(res)

	e.logger.LogInfo(ctx, fmt.Sprintf("response is %s", res))

	response, err := unmarshalCancelOrderResponse(m)
	if err != nil {
		e.logger.LogErr(ctx, "failed to unmarshalModifyOrderResponse", err)
		return nil
	}

	return response
}

func (e *ExchangeImpl) UpdateLeverage(context context.Context, request UpdateLeverageRequest) any {

	timestamp := GetNonce()

	action := UpdateLeverageAction{
		Type:     "updateLeverage",
		Asset:    e.meta[request.Coin].AssetId,
		IsCross:  request.IsCross,
		Leverage: request.Leverage,
	}

	v, r, s := e.SignL1Action(context, request.Address, action, timestamp, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post(context, "/exchange", payload)
	return res
}

func (e *ExchangeImpl) FindOrder(context context.Context, address string, cloid string) OrderResponse {
	return e.infoApi.FindOrder(context, address, cloid)
}

func (e *ExchangeImpl) GetUserFills(context context.Context, address string) []OrderFill {
	return e.infoApi.GetUserFills(context, address)

}

func (e *ExchangeImpl) Withdraw(context context.Context, request WithdrawRequest) *WithdrawResponse {

	timestamp := GetNonce()
	chain := "Testnet"
	chainId := "0x66eee"

	if (*e.cli).IsMainnet() {
		chain = "Mainnet"
		chainId = "0xa4b1"
	}

	amount := ConvertTo2Decimals(request.Amount)
	szDecimals := 2

	action := WithdrawAction{
		Type:             "withdraw3",
		HLChain:          chain,
		SignatureChainId: chainId,
		Amount:           SizeToWire(amount, szDecimals),
		Destination:      request.Destination,
		Time:             timestamp,
	}

	v, r, s := e.SignWithdrawAction(context, request.Address, action, (*e.cli).IsMainnet())

	payload := ExchangeRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	res := (*e.cli).Post(context, "/exchange", payload)
	e.logger.LogInfo(context, fmt.Sprintf("Response is %s", res))
	m, _ := json.Marshal(res)
	response := &WithdrawResponse{}
	_ = json.Unmarshal(m, &response)
	response.Nonce = timestamp

	return response
}

var lastNonce *int64
var lastNonceMu sync.Mutex

// GetNonce is thread safe and makes sure that all nonces are increasing, even if called in the same millisecond
func GetNonce() int64 {
	lastNonceMu.Lock()
	defer lastNonceMu.Unlock()

	nonce := time.Now().UnixMilli()
	if lastNonce == nil {
		lastNonce = &nonce
		return nonce
	} else if *lastNonce >= nonce {
		*lastNonce += 1
		return *lastNonce
	} else {
		lastNonce = &nonce
		return nonce
	}
}

func (e *ExchangeImpl) SignL1Action(ctx context.Context, address string, action any, timestamp int64, isMainnet bool) (byte, [32]byte, [32]byte) {
	hash := e.buildActionHash(ctx, action, "", timestamp)
	message := buildMessage(hash.Bytes(), isMainnet)
	return e.SignInner(ctx, address, message, isMainnet)
}

func (e *ExchangeImpl) SignInner(ctx context.Context, address string, message apitypes.TypedDataMessage, isMainNet bool) (byte, [32]byte, [32]byte) {

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
		DTypeMsg:  message,
		IsMainNet: isMainNet,
	}

	v, r, s, err := signer.Sign(address, req)

	if err != nil {
		e.logger.LogErr(ctx, "Failed to sign request", err)
		panic("Failed to sign request")
	}

	return v, r, s

}

func (e *ExchangeImpl) SignWithdrawAction(ctx context.Context, address string, action WithdrawAction, mainnet bool) (byte, [32]byte, [32]byte) {

	message := apitypes.TypedDataMessage{
		"hyperliquidChain": action.HLChain,
		"destination":      action.Destination,
		"amount":           action.Amount,
		"time":             strconv.FormatInt(action.Time, 10),
	}

	signer := NewSigner(e.keyManager)
	req := SigRequest{
		PrimaryType: "HyperliquidTransaction:Withdraw",
		DType: []apitypes.Type{
			{
				Name: "hyperliquidChain",
				Type: "string",
			},
			{
				Name: "destination",
				Type: "string",
			},
			{
				Name: "amount",
				Type: "string",
			},
			{
				Name: "time",
				Type: "uint64",
			},
		},
		DTypeMsg:  message,
		IsMainNet: mainnet,
	}

	v, r, s, err := signer.Sign(address, req)

	if err != nil {
		e.logger.LogErr(ctx, "Failed to sign request", err)
		panic("Failed to sign request")
	}

	return v, r, s

}

func (e *ExchangeImpl) buildActionHash(ctx context.Context, action any, vaultAd string, nonce int64) common.Hash {
	var (
		data []byte
	)

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseCompactInts(true)
	err := enc.Encode(action)
	if err != nil {
		e.logger.LogErr(ctx, "Failed to pack the data", err)
		panic(fmt.Sprintf("Failed to pack the data %s", err))
	}
	data = buf.Bytes()

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
