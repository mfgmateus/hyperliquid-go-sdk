package hyperliquid

import (
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const MainnetUrl = "https://api.hyperliquid.xyz"

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
