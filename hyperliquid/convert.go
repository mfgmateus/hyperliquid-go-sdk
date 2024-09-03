package hyperliquid

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
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

func ModifyOrderWiresToModifyOrderAction(orders []ModifyOrderWire) ModifyOrdersAction {
	return ModifyOrdersAction{
		Type:   "batchModify",
		Orders: orders,
	}
}

func OrderReqToWire(req OrderRequest, meta map[string]AssetInfo) OrderWire {
	info, ok := meta[req.Coin]
	if !ok {
		panic(fmt.Sprintf("coin (%v) is not defined in meta table", req.Coin))
	}
	return OrderWire{
		Asset:      info.AssetId,
		IsBuy:      req.IsBuy,
		LimitPx:    PriceToWire(req.LimitPx, info.SzDecimals),
		SizePx:     SizeToWire(req.Sz, info.SzDecimals),
		ReduceOnly: req.ReduceOnly,
		OrderType:  OrderTypeToWire(req.OrderType),
		Cloid:      req.Cloid,
	}
}

func ModifyOrderReqToWire(req ModifyOrderRequest, meta map[string]AssetInfo) ModifyOrderWire {
	info, ok := meta[req.Coin]
	if !ok {
		panic(fmt.Sprintf("coin (%v) is not defined in meta table", req.Coin))
	}
	return ModifyOrderWire{
		OidOrCloid: req.OidOrCloid,
		Order: OrderWire{
			Asset:      info.AssetId,
			IsBuy:      req.IsBuy,
			LimitPx:    PriceToWire(req.LimitPx, info.SzDecimals),
			SizePx:     SizeToWire(req.Sz, info.SzDecimals),
			ReduceOnly: req.ReduceOnly,
			OrderType:  OrderTypeToWire(req.OrderType),
			Cloid:      req.Cloid,
		},
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

func int64ToFixedSize(x int64, decimals int) string {
	// this is required as providing trailing 0s, even if we provide the correct number of decimals, would
	// result in an invalid hash. The hashing process uses msgpack to create the hash, which uses the string
	// representation of the number (returned here) so this needs to match HyperLiquid way of doing it.
	for x%10 == 0 && decimals > 0 {
		decimals -= 1
		x /= 10
	}

	str := strconv.FormatInt(x, 10)

	if decimals == 0 {
		return str
	}

	if len(str) <= decimals {
		// number is < 0
		neededZeros := decimals - len(str)
		str = "0." + strings.Repeat("0", neededZeros) + str
	} else {
		str = str[:len(str)-decimals] + "." + str[len(str)-decimals:]
	}
	return str
}

func PriceToWire(x float64, szDecimals int) string {
	// Prices can have up to 5 significant figures, but no more than MAX_DECIMALS - szDecimals decimal places
	// where MAX_DECIMALS is 6 for perps and 8 for spot.
	// This only treats perps case

	bigf := big.NewFloat(x)

	maxDecSz := 0
	intPart, _ := bigf.Int64()

	// all significant figures are after the decimal point
	if intPart == 0 {
		// default limit is MAX_DECIMALS - szDecimals
		maxDecSz = 6 - szDecimals

		if x >= 0.1 && maxDecSz < 5 {
			// limit to 5 significant figures, because first digit after the decimal point is != 0
			maxDecSz = 5
		}
	} else {
		intSize := len(strconv.FormatInt(intPart, 10))
		if intSize > 5 {
			panic("price too big. at most 5 decimals are allowed")
		} else {
			maxDecSz = 5 - intSize
		}
	}

	exp := math.Pow(10.0, float64(maxDecSz))
	return int64ToFixedSize(int64(x*exp), maxDecSz)
}

func SizeToWire(x float64, szDecimals int) string {
	exp := math.Pow(10.0, float64(szDecimals))
	return int64ToFixedSize(int64(x*exp), szDecimals)
}

func ConvertTo2Decimals(x float64) float64 {
	return math.Floor(x*100) / 100
}
