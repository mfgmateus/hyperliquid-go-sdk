package hyperliquid

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type AssetInfo struct {
	SzDecimals int
	AssetId    int
}

type OrderRequest struct {
	Coin       string    `json:"coin"`
	IsBuy      bool      `json:"is_buy"`
	Sz         float64   `json:"sz"`
	LimitPx    float64   `json:"limit_px"`
	OrderType  OrderType `json:"order_type"`
	ReduceOnly bool      `json:"reduce_only"`
	Cloid      *string   `json:"cloid,omitempty"`
}

type ModifyOrderRequest struct {
	OidOrCloid interface{} `json:"oid"`
	Coin       string      `json:"coin"`
	IsBuy      bool        `json:"is_buy"`
	Sz         float64     `json:"sz"`
	LimitPx    float64     `json:"limit_px"`
	OrderType  OrderType   `json:"order_type"`
	ReduceOnly bool        `json:"reduce_only"`
	Cloid      *string     `json:"cloid,omitempty"`
}

type OrderType struct {
	Limit   *LimitOrderType   `json:"limit" json:",omitempty" msgpack:"limit" msgpack:",omitempty"`
	Trigger *TriggerOrderType `json:"trigger" json:",omitempty"  msgpack:"trigger" msgpack:",omitempty"`
}

type LimitOrderType struct {
	Tif string `json:"tif" msgpack:"tif"`
}

type TriggerOrderType struct {
	IsMarket  bool   `json:"isMarket" msgpack:"isMarket"`
	TriggerPx string `json:"triggerPx" msgpack:"triggerPx"`
	TpSl      TpSl   `json:"tpsl" msgpack:"tpsl"`
}

type TpSl string

const TriggerTp TpSl = "tp"
const TriggerSl TpSl = "sl"

type Grouping string

const GroupingNa Grouping = "na"
const GroupingTpSl Grouping = "positionTpsl"

type RsvSignature struct {
	R string `json:"r"`
	S string `json:"s"`
	V byte   `json:"v"`
}

type ExchangeRequest struct {
	Action       any          `json:"action"`
	Nonce        int64        `json:"nonce"`
	Signature    RsvSignature `json:"signature"`
	VaultAddress *string      `json:"vaultAddress"`
}

type Message struct {
	Source       string `json:"source"`
	ConnectionId []byte `json:"connectionId"`
}

type PlaceOrderAction struct {
	Type     string      `msgpack:"type" json:"type"`
	Orders   []OrderWire `msgpack:"orders" json:"orders"`
	Grouping Grouping    `msgpack:"grouping" json:"grouping"`
}

type ModifyOrdersAction struct {
	Type   string            `msgpack:"type" json:"type"`
	Orders []ModifyOrderWire `msgpack:"modifies" json:"modifies"`
}

type CancelOidOrderAction struct {
	Type    string          `msgpack:"type" json:"type"`
	Cancels []CancelOidWire `msgpack:"cancels" json:"cancels"`
}

type CancelCloidOrderAction struct {
	Type    string            `msgpack:"type" json:"type"`
	Cancels []CancelCloidWire `msgpack:"cancels" json:"cancels"`
}

type CancelCloidWire struct {
	Asset int    `msgpack:"asset" json:"asset"`
	Cloid string `msgpack:"cloid" json:"cloid"`
}

type CancelOidWire struct {
	Asset int   `msgpack:"a" json:"a"`
	Oid   int64 `msgpack:"o" json:"o"`
}

type UpdateLeverageAction struct {
	Type     string `msgpack:"type" json:"type"`
	Asset    int    `msgpack:"asset" json:"asset"`
	IsCross  bool   `msgpack:"isCross" json:"isCross"`
	Leverage int    `msgpack:"leverage" json:"leverage"`
}

type WithdrawAction struct {
	Type             string `msgpack:"type" json:"type"`
	HLChain          string `msgpack:"hyperliquidChain" json:"hyperliquidChain"`
	SignatureChainId string `msgpack:"signatureChainId" json:"signatureChainId"`
	Destination      string `msgpack:"destination" json:"destination"`
	Amount           string `msgpack:"amount" json:"amount"`
	Time             int64  `msgpack:"time" json:"time"`
}

type OrderWire struct {
	Asset      int           `msgpack:"a" json:"a"`
	IsBuy      bool          `msgpack:"b" json:"b"`
	LimitPx    string        `msgpack:"p" json:"p"`
	SizePx     string        `msgpack:"s" json:"s"`
	ReduceOnly bool          `msgpack:"r" json:"r"`
	OrderType  OrderTypeWire `msgpack:"t" json:"t"`
	Cloid      *string       `msgpack:"c,omitempty" json:"c,omitempty"`
}

type ModifyOrderWire struct {
	// OidOrCloid is either Cloid (string) or Oid int64
	OidOrCloid interface{} `msgpack:"oid" json:"oid"`
	Order      OrderWire   `msgpack:"order" json:"order"`
}

type OrderTypeWire struct {
	Limit   *LimitOrderType   `json:"limit,omitempty" msgpack:"limit,omitempty"`
	Trigger *TriggerOrderType `json:"trigger,omitempty" msgpack:"trigger,omitempty"`
}

type SigRequest struct {
	PrimaryType string
	DType       []apitypes.Type
	DTypeMsg    map[string]interface{}
	IsMainNet   bool
}

func (req SigRequest) GetChainId() *math.HexOrDecimal256 {
	if req.PrimaryType == "HyperliquidTransaction:Withdraw" {
		if req.IsMainNet {
			return math.NewHexOrDecimal256(int64(42161))
		} else {
			return math.NewHexOrDecimal256(int64(421614))
		}
	} else {
		return math.NewHexOrDecimal256(int64(1337))
	}
}

type CloseRequest struct {
	Address  string
	Coin     string
	Px       *float64
	Sz       *float64
	Slippage *float64
	Cloid    *string
}

type TriggerRequest struct {
	Address  string
	Coin     string
	Px       *float64
	Sz       *float64
	Slippage *float64
	Cloid    *string
	Trigger  TriggerOrderType
}

type OpenRequest struct {
	Address  string
	Coin     string
	IsBuy    bool
	Px       *float64
	Sz       *float64
	Slippage *float64
	Cloid    *string
}

type WithdrawRequest struct {
	Address     string
	Destination string
	Amount      float64
}

type WithdrawWire struct {
	Destination string `json:"destination"`
	Amount      string `json:"amount"`
	HLChain     string `json:"hyperliquidChain"`
	Time        int64  `json:"time"`
}

type UpdateLeverageRequest struct {
	Address  string
	Coin     string
	IsCross  bool
	Leverage int
}

// TODO: most probably this should be used by other errors as well
func unmarshalInnerResponse(data []byte, responseTarget any) (status string, responseErr *string, err error) {
	// try to unmarshal in generic type, to get the Response field
	var temp struct {
		Status   string          `json:"status"`
		Response json.RawMessage `json:"response"`
	}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		return "", nil, err
	}

	// check if response field is string
	var str string
	err = json.Unmarshal(temp.Response, &str)
	if err == nil {
		return temp.Status, &str, nil
	}

	// check if response field is responseTarget
	err = json.Unmarshal(temp.Response, responseTarget)
	if err != nil {
		return temp.Status, nil, fmt.Errorf("unable to unmarshal response: %w", err)
	} else {
		return temp.Status, nil, nil
	}
}

func unmarshalPlaceOrderResponse(data []byte) (response *PlaceOrderResponse, err error) {
	response = &PlaceOrderResponse{
		Response: new(InnerResponse),
	}
	response.Status, response.ResponseErr, err = unmarshalInnerResponse(data, response.Response)
	if err != nil {
		return nil, err
	}
	if response.ResponseErr != nil {
		response.Response = nil
	}
	return response, nil
}

func unmarshalModifyOrderResponse(data []byte) (response *ModifyOrderResponse, err error) {
	response = &ModifyOrderResponse{
		Response: new(InnerResponse),
	}
	response.Status, response.ResponseErr, err = unmarshalInnerResponse(data, response.Response)
	if err != nil {
		return nil, err
	}
	if response.ResponseErr != nil {
		response.Response = nil
	}
	return response, nil
}

func unmarshalCancelOrderResponse(data []byte) (response *CancelOrderResponse, err error) {
	response = &CancelOrderResponse{
		Response: new(InnerCancelResponse),
	}
	response.Status, response.ResponseErr, err = unmarshalInnerResponse(data, response.Response)
	if err != nil {
		return nil, err
	}
	if response.ResponseErr != nil {
		response.Response = nil
	}
	return response, nil
}

type PlaceOrderResponse struct {
	Status      string         `json:"status"`
	ResponseErr *string        // json["response"] is either ResponseErr or Response
	Response    *InnerResponse // json["response"] is either ResponseErr or Response
}

type ModifyOrderResponse struct {
	Status      string         `json:"status"`
	ResponseErr *string        // json["response"] is either ResponseErr or Response
	Response    *InnerResponse // json["response"] is either ResponseErr or Response
}

type WithdrawResponse struct {
	Status string `json:"status"`
	Nonce  int64
}

type CancelOrderResponse struct {
	Status      string               `json:"status"`
	ResponseErr *string              // json["response"] is either ResponseErr or Response
	Response    *InnerCancelResponse // json["response"] is either ResponseErr or Response
}

type OrderStatus string

const (
	OrderStatusFilled OrderStatus = "FILLED"
	OrderStatusOpen   OrderStatus = "OPEN"
	OrderStatusFailed OrderStatus = "FAILED"
)

func (r PlaceOrderResponse) GetAvgPrice() *string {
	for _, status := range r.Response.Data.Statuses {
		if status.Filled != nil {
			return &status.Filled.AvgPx
		} else {
			return nil
		}
	}
	return nil
}

func (r PlaceOrderResponse) GetStatus() OrderStatus {
	if r.Status != "ok" {
		return OrderStatusFailed
	}
	for _, status := range r.Response.Data.Statuses {
		if status.Error != nil {
			return OrderStatusFailed
		}
		if status.Filled != nil {
			return OrderStatusFilled
		}
		if status.Resting != nil {
			return OrderStatusOpen
		}
	}
	return OrderStatusOpen
}

type InnerResponse struct {
	Type string       `json:"type"`
	Data DataResponse `json:"data"`
}

type InnerCancelResponse struct {
	Data CancelDataResponse `json:"data"`
}

type DataResponse struct {
	Statuses []StatusResponse `json:"statuses"`
}

type CancelDataResponse struct {
	Statuses []string `json:"statuses"`
}

func (r CancelOrderResponse) IsCancelled() bool {
	if r.Status != "ok" {
		return false
	}
	for _, status := range r.Response.Data.Statuses {
		if status == "success" {
			return true
		}
		return false
	}
	return false
}

type StatusResponse struct {
	Resting *RestingStatus `json:"resting"`
	Filled  *FilledStatus  `json:"filled"`
	Error   *string        `json:"error"`
}

type RestingStatus struct {
	OrderId int64  `json:"oid"`
	Cloid   string `json:"cloid"`
}

type FilledStatus struct {
	OrderId int64  `json:"oid"`
	Cloid   string `json:"cloid"`
	AvgPx   string `json:"avgPx"`
	TotalSz string `json:"totalSz"`
}

type OpenOrder struct {
	Coin      string `json:"coin"`
	LimitPx   string `json:"limitPx"`
	Oid       int64  `json:"oid"`
	Side      string `json:"side"`
	Sz        string `json:"sz"`
	Timestamp int64  `json:"timestamp"`
}

type OrderResponse struct {
	Order struct {
		Order struct {
			Children         []any  `json:"children"`
			Cloid            string `json:"cloid"`
			Coin             string `json:"coin"`
			IsPositionTpsl   bool   `json:"isPositionTpsl"`
			IsTrigger        bool   `json:"isTrigger"`
			LimitPx          string `json:"limitPx"`
			Oid              int64  `json:"oid"`
			OrderType        string `json:"orderType"`
			OrigSz           string `json:"origSz"`
			ReduceOnly       bool   `json:"reduceOnly"`
			Side             string `json:"side"`
			Sz               string `json:"sz"`
			Tif              string `json:"tif"`
			Timestamp        int64  `json:"timestamp"`
			TriggerCondition string `json:"triggerCondition"`
			TriggerPx        string `json:"triggerPx"`
		} `json:"order"`
		Status          string `json:"status"`
		StatusTimestamp int64  `json:"statusTimestamp"`
	} `json:"order"`
	Status string `json:"status"`
}

type OrderFill struct {
	Cloid         string       `json:"cloid"`
	ClosedPnl     string       `json:"closedPnl"`
	Coin          string       `json:"coin"`
	Crossed       bool         `json:"crossed"`
	Dir           string       `json:"dir"`
	Fee           string       `json:"fee"`
	FeeToken      string       `json:"feeToken"`
	Hash          string       `json:"hash"`
	Oid           int          `json:"oid"`
	Px            string       `json:"px"`
	Side          string       `json:"side"`
	StartPosition string       `json:"startPosition"`
	Sz            string       `json:"sz"`
	Tid           int64        `json:"tid"`
	Time          int64        `json:"time"`
	Liquidation   *Liquidation `json:"liquidation"`
}

type Liquidation struct {
	User      string `json:"liquidatedUser"`
	MarkPrice string `json:"markPx"`
	Method    string `json:"method"`
}

type KeyManager interface {
	GetKey(address string) *ecdsa.PrivateKey
}

type NonFundingUpdate struct {
	Hash  string          `json:"hash"`
	Time  int64           `json:"time"`
	Delta NonFundingDelta `json:"delta"`
}

type FundingUpdate struct {
	Hash  string       `json:"hash"`
	Time  int64        `json:"time"`
	Delta FundingDelta `json:"delta"`
}

type NonFundingDelta struct {
	Type   string  `json:"type"`
	Amount string  `json:"usdc"`
	Fee    *string `json:"fee"`
	Nonce  *int64  `json:"nonce"`
}

type FundingDelta struct {
	Asset       string `json:"coin"`
	FundingRate string `json:"fundingRate"`
	Size        string `json:"szi"`
	UsdcAmount  string `json:"usdc"`
}

type Withdrawal struct {
	Hash   string  `json:"hash"`
	Amount string  `json:"usdc"`
	Fee    *string `json:"fee"`
	Nonce  *int64  `json:"nonce"`
}
