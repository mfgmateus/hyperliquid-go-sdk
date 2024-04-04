package cryptoutil

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type PKeyManager interface {
	PublicECDSA() *ecdsa.PublicKey
	PrivateECDSA() *ecdsa.PrivateKey
	PublicAddress() common.Address
}

type Pkey struct {
	privKey *ecdsa.PrivateKey
	pubKey  *ecdsa.PublicKey
}

func NewPkey(pkey string) PKeyManager {
	privKey, err := crypto.HexToECDSA(pkey)
	if err != nil {
		panic("unable to load private key.")
	}

	pubKey, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		panic("unable to extract public key from private.")
	}

	return Pkey{privKey: privKey, pubKey: pubKey}
}

func (p Pkey) PublicECDSA() *ecdsa.PublicKey {
	return p.pubKey
}

func (p Pkey) PrivateECDSA() *ecdsa.PrivateKey {
	return p.privKey
}

func (p Pkey) PublicAddress() common.Address {
	return crypto.PubkeyToAddress(*p.pubKey)
}
