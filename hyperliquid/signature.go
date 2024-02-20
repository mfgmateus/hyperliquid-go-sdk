package hyperliquid

import (
	"github.com/STFX-IO/hyperliquid-go-sdk/cryptoutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const ChainId = 1337
const VerifyingContract = "0x0000000000000000000000000000000000000000"

type Signer struct {
	pkey cryptoutil.PKeyManager
}

func NewSigner(pkey cryptoutil.PKeyManager) Signer {
	return Signer{
		pkey: pkey,
	}
}

func (signer Signer) Sign(req SigRequest) (byte, [32]byte, [32]byte, error) {
	var (
		err error
	)

	types := GetContractTypes(req)
	domain := GetDomain()
	typedData := apitypes.TypedData{
		Types:       types,
		PrimaryType: req.PrimaryType,
		Domain:      domain,
		Message:     req.DTypeMsg,
	}

	bytes, _, err := apitypes.TypedDataAndHash(typedData)

	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	sig, err := crypto.Sign(bytes, signer.pkey.PrivateECDSA())

	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	return SigToVRS(sig)
}
