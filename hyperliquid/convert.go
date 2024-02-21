package hyperliquid

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func OrderWiresToOrderAction(orders []OrderWire) PlaceOrderAction {
	return PlaceOrderAction{
		Type:     "order",
		Grouping: "na",
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
		}
	}
	return OrderTypeWire{}
}

func FloatToWire(x float64, szDecimals *int) string {
	// Format the float with custom decimal places, default is 6
	decimals := 6
	if szDecimals != nil {
		decimals = *szDecimals
	}

	rounded := fmt.Sprintf("%.*f", decimals, x)
	for strings.HasSuffix(rounded, "0") {
		rounded = strings.TrimSuffix(rounded, "0")
	}

	rounded = strings.TrimSuffix(rounded, ".")
	return rounded
}