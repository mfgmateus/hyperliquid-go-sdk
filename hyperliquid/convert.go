package hyperliquid

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math"
	"math/big"
	"strconv"
	"strings"
)

func ToTypedSig(r [32]byte, s [32]byte, v byte) RsvSignature {
	return RsvSignature{
		R: hexutil.Encode(r[:]),
		S: hexutil.Encode(s[:]),
		V: v,
	}
}

func ArrayAppend(data []byte, toAppend []byte) []byte {
	return append(data, toAppend...)
}

func HexToBytes(addr string) []byte {
	if strings.HasPrefix(addr, "0x") {
		fAddr := strings.Replace(addr, "0x", "", 1)
		b, _ := hex.DecodeString(fAddr)
		return b
	} else {
		b, _ := hex.DecodeString(addr)
		return b
	}
}

func OrderWiresToOrderAction(orders []OrderWire, grouping Grouping) PlaceOrderAction {
	return PlaceOrderAction{
		Type:     "order",
		Grouping: grouping,
		Orders:   orders,
	}
}

func OrderReqToWire(req OrderRequest, meta map[string]AssetInfo) OrderWire {
	info := meta[req.Coin]
	return OrderWire{
		Asset:      info.AssetId,
		IsBuy:      req.IsBuy,
		LimitPx:    FloatToWire(req.LimitPx, nil),
		SizePx:     FloatToWire(req.Sz, &info.SzDecimals),
		ReduceOnly: req.ReduceOnly,
		OrderType:  OrderTypeToWire(req.OrderType),
		Cloid:      req.Cloid,
	}
}

func OrderTypeToWire(orderType OrderType) OrderTypeWire {
	if orderType.Limit != nil {
		return OrderTypeWire{
			Limit: &LimitOrderType{
				Tif: orderType.Limit.Tif,
			},
			Trigger: nil,
		}
	} else if orderType.Trigger != nil {
		return OrderTypeWire{
			Trigger: &TriggerOrderType{
				TpSl:      orderType.Trigger.TpSl,
				TriggerPx: orderType.Trigger.TriggerPx,
				IsMarket:  orderType.Trigger.IsMarket,
			},
			Limit: nil,
		}
	}
	return OrderTypeWire{}
}

func FloatToWire(x float64, szDecimals *int) string {
	// Format the float with custom decimal places, default is 6
	//hyperliquid only allows at most 6 digits.

	bigf := big.NewFloat(x)
	var maxDecSz uint
	if szDecimals != nil {
		maxDecSz = uint(*szDecimals)
	} else {
		intPart, _ := bigf.Int64()
		intSize := len(strconv.FormatInt(intPart, 10))
		if intSize >= 6 {
			maxDecSz = 0
		} else {
			maxDecSz = uint(6 - intSize)
		}
	}

	x, _ = bigf.Float64()
	mult := math.Pow(10.0, float64(maxDecSz))

	rounded := math.Floor(x*mult) / mult
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

func FloatToWire2(x float64, maxFigure int) string {
	// Format the float with custom decimal places, default is 6
	//hyperliquid only allows at most 6 digits.

	bigf := big.NewFloat(x)

	maxDecSz := 0

	intPart, _ := bigf.Int64()
	intSize := len(strconv.FormatInt(intPart, 10))
	if intSize >= maxFigure {
		maxDecSz = 0
	} else {
		maxDecSz = maxFigure - intSize
	}

	x, _ = bigf.Float64()
	mult := math.Pow(10.0, float64(maxDecSz))

	rounded := math.Floor(x*mult) / mult
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

func ConvertTo2Decimals(x float64) float64 {
	return math.Floor(x*100) / 100
}
