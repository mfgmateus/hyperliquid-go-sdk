package hyperliquid

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/STFX-IO/hyperliquid-go-sdk/cryptoutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/vmihailenco/msgpack/v5"
	"math"
	"strconv"
	"time"
)

type ExchangeImpl struct {
	pkeyManager cryptoutil.PKeyManager
	walletAddr  string
	infoApi     InfoApi
	cli         *API
	meta        map[string]AssetInfo
}

func NewExchange(manager cryptoutil.PKeyManager, api InfoApi, meta map[string]AssetInfo, walletAddr string, cli *API) ExchangeImpl {
	return ExchangeImpl{
		pkeyManager: manager,
		infoApi:     api,
		meta:        meta,
		cli:         cli,
		walletAddr:  walletAddr,
	}
}

func (e *ExchangeImpl) GetAddress() string {
	return e.pkeyManager.PublicAddress().String()
}

func (e *ExchangeImpl) SlippagePrice(coin string, isBuy bool, slippage float64, px *float64) float64 {

	if px == nil || *px <= 0.0 {
		px = new(float64)
		parsed := GetMktPx(coin, e)
		*px = parsed
	}

	return CalculateSlippage(isBuy, px, slippage)

}

func GetMktPx(coin string, e *ExchangeImpl) float64 {
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

func (e *ExchangeImpl) MarketOpen(req OpenRequest) any {

	slippage := GetSlippage(req.Slippage)
	finalPx := e.SlippagePrice(req.Coin, req.IsBuy, slippage, req.Px)

	orderType := OrderType{
		Limit: &LimitOrderType{
			Tif: "Ioc",
		},
	}

	return e.Order(req.Coin, req.IsBuy, *req.Sz, finalPx, orderType, false, req.Cloid)

}

func (e *ExchangeImpl) MarketClose(req CloseRequest) any {

	positions := e.infoApi.GetUserState(e.walletAddr).AssetPositions
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

		return e.Order(req.Coin, isBuy, *sz, finalPx, orderType, true, req.Cloid)

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

func (e *ExchangeImpl) Order(coin string, isBuy bool, sz float64,
	px float64, orderType OrderType, reduceOnly bool, cloid *string) any {
	order := OrderRequest{
		Coin:       coin,
		IsBuy:      isBuy,
		Sz:         sz,
		LimitPx:    px,
		OrderType:  orderType,
		ReduceOnly: reduceOnly,
		Cloid:      cloid,
	}

	return e.BulkOrders([]OrderRequest{order})

}

func (e *ExchangeImpl) BulkOrders(requests []OrderRequest) any {
	var wires []OrderWire
	for _, req := range requests {
		wires = append(wires, OrderReqToWire(req, e.meta))
	}

	timestamp := GetNonce()
	action := OrderWiresToOrderAction(wires)

	v, r, s := e.SignL1Action(action, timestamp, (*e.cli).IsMainnet())

	payload := PlaceOrderRequest{
		Action:       action,
		Nonce:        timestamp,
		Signature:    ToTypedSig(r, s, v),
		VaultAddress: nil,
	}

	p, _ := json.Marshal(payload)
	fmt.Printf("CloseRequest body is %s\n", p)

	res := (*e.cli).Post("/exchange", payload)
	return res
}

func GetNonce() int64 {
	return time.Now().UnixMilli()
}

func (e *ExchangeImpl) SignL1Action(action Action, timestamp int64, isMainnet bool) (byte, [32]byte, [32]byte) {
	hash := buildActionHash(action, "", timestamp)
	message := buildMessage(hash.Bytes(), isMainnet)
	return e.SignInner(message)

}

func (e *ExchangeImpl) SignInner(message apitypes.TypedDataMessage) (byte, [32]byte, [32]byte) {

	signer := NewSigner(e.pkeyManager)
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

	v, r, s, err := signer.Sign(req)

	if err != nil {
		fmt.Printf("Error %s", err)
		panic("Failed to sign request")
	}

	return v, r, s

}

func buildActionHash(action Action, vaultAd string, nonce int64) common.Hash {
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
