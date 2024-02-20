package cryptoutil

import (
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
)

var (
	errSignature = errors.New("error signing message")
)

type PKeyManager interface {
	PublicECDSA() *ecdsa.PublicKey
	PrivateECDSA() *ecdsa.PrivateKey
	PublicAddress() common.Address
	SignMessage(m []byte) ([]byte, error)
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

func (p Pkey) SignMessage(m []byte) ([]byte, error) {
	sig, err := crypto.Sign(m, p.PrivateECDSA())
	if err != nil {
		log.Printf("error signing message. Reason: %s", err.Error())
		return nil, err
	}

	if sig[64] == 0 || sig[64] == 1 {
		sig[64] += 27
	}

	return sig, nil
}
