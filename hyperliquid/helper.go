package hyperliquid

import (
	"crypto/rand"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"log"
)

const MainnetUrl = "https://api.hyperliquid.xyz"
const TestnetUrl = "https://api.hyperliquid-testnet.xyz"
const DefaultSlippage = 0.05

func SigToVRS(sig []byte) (byte, [32]byte, [32]byte, error) {
	var v byte
	var r [32]byte
	var s [32]byte

	v = sig[64] + 27
	copy(r[:], sig[:32])
	copy(s[:], sig[32:64])

	return v, r, s, nil
}

func GetContractTypes(req SigRequest) apitypes.Types {
	types := apitypes.Types{
		req.PrimaryType: req.DType,
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
	}
	return types
}

func GetDomain() apitypes.TypedDataDomain {
	return apitypes.TypedDataDomain{
		Name:              "Exchange",
		Version:           "1",
		ChainId:           math.NewHexOrDecimal256(ChainId),
		VerifyingContract: VerifyingContract,
	}
}

func GetRandomCloid() string {
	buf := make([]byte, 16)
	// then we can call rand.Read.
	_, err := rand.Read(buf)
	if err != nil {
		log.Fatalf("error while generating random string: %s", err)
	}

	return hexutil.Encode(buf)
}

func BuildMetaMap(info InfoApi) map[string]AssetInfo {
	metaMap := make(map[string]AssetInfo)
	for index, asset := range info.GetMeta().Universe {
		metaMap[asset.Name] = AssetInfo{
			SzDecimals: asset.SzDecimals,
			AssetId:    index,
		}
	}
	return metaMap
}
