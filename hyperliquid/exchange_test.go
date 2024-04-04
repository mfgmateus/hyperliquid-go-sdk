package hyperliquid

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mfgmateus/hyperliquid-go-sdk/cryptoutil"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"
)

const Address = "0x60Cc17b782e9c5f14806663f8F617921275b9720"
const PrivateKey = "35e02d3d3e6f65dcc37886ab779af1c4e01d4b915a06bdacbcdb4da09497996c"

var (
	keyManager  = NewKeyManager(PrivateKey)
	baseClient  = NewApiDefault(TestnetUrl)
	exchangeApi = NewExchange(&baseClient, &keyManager)
)

type SingleKeyManager struct {
	privKey *ecdsa.PrivateKey
}

func (m SingleKeyManager) GetKey(address string) *ecdsa.PrivateKey {
	return m.privKey
}

func NewKeyManager(privKey string) KeyManager {
	manager := cryptoutil.NewPkey(privKey)
	return &SingleKeyManager{privKey: manager.PrivateECDSA()}
}

func TestMarketOpenAndClose(t *testing.T) {

	size := 10.0
	cloid := GetRandomCloid()

	const coin = "ARB"
	req := OpenRequest{
		Address: Address,
		Coin:    coin,
		Sz:      &size,
		Cloid:   &cloid,
	}

	result := exchangeApi.MarketOpen(req)
	m, _ := json.Marshal(result)
	fmt.Printf("Open Result is %s\n", m)

	r2 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)

	cloid = GetRandomCloid()

	closeReq := CloseRequest{
		Address: Address,
		Coin:    coin,
		Cloid:   &cloid,
	}

	//wait for 2 seconds?
	time.Sleep(time.Duration(time.Duration.Seconds(2)))

	//place a take profit order

	result = exchangeApi.MarketClose(closeReq)
	fmt.Printf("%s\n", *result.GetAvgPrice())
	m, _ = json.Marshal(result)

	fmt.Printf("Close Result is %s\n", m)
	r2 = exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)
}

func TestMarketClose(t *testing.T) {

	cloid := GetRandomCloid()

	req := CloseRequest{
		Coin:  "ARB",
		Cloid: &cloid,
	}

	result := exchangeApi.MarketClose(req)
	fmt.Printf("%s\n", *result.GetAvgPrice())
	m, _ := json.Marshal(result)
	fmt.Printf("Result is %s\n", m)

	r2 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s", m)

}

func TestAccountInfo(t *testing.T) {

	infoApi := NewInfoApi(&baseClient)
	state := infoApi.GetUserState(Address)

	m, _ := json.Marshal(state)
	fmt.Printf("Result is %s\n", m)

}

func TestUpdateLeverage(t *testing.T) {

	req := UpdateLeverageRequest{
		Coin:     "ARB",
		Leverage: 5,
		IsCross:  false,
	}

	result := exchangeApi.UpdateLeverage(req)
	m, _ := json.Marshal(result)

	fmt.Printf("Result is %s\n", m)

}

func TestGetUserFills(t *testing.T) {

	fills := exchangeApi.GetUserFills(Address)
	m, _ := json.Marshal(fills)
	fmt.Printf("Result is %s", m)

}

func TestTrigger(t *testing.T) {

	triggerPrice := 2.10
	decimals := 4
	slippage := float64(0)
	price := float64(0)
	cloid := GetRandomCloid()

	req := TriggerRequest{
		Coin:     "ARB",
		Px:       &price,
		Slippage: &slippage,
		Trigger: TriggerOrderType{
			TriggerPx: FloatToWire(triggerPrice, &decimals),
			TpSl:      TriggerTp,
			IsMarket:  true,
		},
		Cloid: &cloid,
	}

	result := exchangeApi.Trigger(req)
	m, _ := json.Marshal(result)
	fmt.Printf("Trigger Result is %s\n", m)

	r2 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s", m)

}

func TestCancel(t *testing.T) {

	//cloid := GetRandomCloid()
	mkPrice := exchangeApi.GetMktPx("ARB")
	mkPrice = mkPrice * 1.05
	var (
		cloid  string
		result *PlaceOrderResponse
		m      any
		order  OrderResponse
	)

	cloid = GetRandomCloid()

	var req = OrderRequest{
		Coin:       "ARB",
		IsBuy:      true,
		LimitPx:    mkPrice,
		Sz:         10,
		OrderType:  OrderType{Limit: &LimitOrderType{Tif: "Gtc"}},
		Cloid:      &cloid,
		ReduceOnly: false,
	}

	result = exchangeApi.Order(Address, req, "na")
	m, _ = json.Marshal(result)
	fmt.Printf("Order Result is %s\n", m)

	triggerPrice := mkPrice * 1.5
	decimals := 4
	slippage := float64(0)
	price := float64(0)
	cloid = GetRandomCloid()

	req2 := TriggerRequest{
		Coin:     "ARB",
		Px:       &price,
		Slippage: &slippage,
		Trigger: TriggerOrderType{
			TriggerPx: FloatToWire(triggerPrice, &decimals),
			TpSl:      TriggerTp,
			IsMarket:  true,
		},
		Cloid: &cloid,
	}

	result = exchangeApi.Trigger(req2)
	m, _ = json.Marshal(result)
	fmt.Printf("Trigger Result is %s\n", m)

	order = exchangeApi.FindOrder(Address, cloid)

	r2 := exchangeApi.CancelOrder(Address, "ARB", cloid)
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)
	fmt.Printf("Result is %s\n", strconv.FormatBool(r2.IsCancelled()))

	r2 = exchangeApi.CancelOrderByOid(Address, "ARB", int(order.Order.Order.Oid))
	m, _ = json.Marshal(r2)
	fmt.Printf("Result is %s\n", m)
	fmt.Printf("Result is %s\n", strconv.FormatBool(r2.IsCancelled()))
	//
	r3 := exchangeApi.FindOrder(Address, cloid)
	m, _ = json.Marshal(r3)
	fmt.Printf("Result is %s\n", m)

}

func TestAccountSetup(t *testing.T) {

	GenerateEOA()

}

func GenerateEOA() (string, string) {
	newAccountPk, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyBytes := crypto.FromECDSA(newAccountPk)
	privKey := hexutil.Encode(privateKeyBytes)
	fmt.Println("SAVE BUT DO NOT SHARE THIS (Private Key):", privKey)

	publicKey := newAccountPk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println("Public Key:", hexutil.Encode(publicKeyBytes))

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println("Address:", address)

	return privKey, address
}

func TestAccountTransfer(t *testing.T) {

	client, err := ethclient.Dial("https://public.stackup.sh/api/v1/node/arbitrum-sepolia")
	if err != nil {
		log.Fatal(err)
	}

	privKey, eoaAddress := GenerateEOA()

	contractAddress := "0x279c9462fdba349550b49a23de27dd19d5891baa"

	amountX6 := big.NewInt(5000000)

	//transfer from vault to EOA
	vaultToEoaTx := TransferUSDC(client, PrivateKey, amountX6, eoaAddress)
	_, err = waitTransfer(client, vaultToEoaTx)
	if err != nil {
		log.Fatal(err)
	}
	//17063200000000
	gasPrice := new(big.Int)
	gasPrice.SetString("65563200000000", 10)

	log.Printf("Gas price is %s\n", gasPrice.String())
	vltToEoaGasTx := TransferETH(client, PrivateKey, gasPrice, eoaAddress)
	_, err = waitTransfer(client, vltToEoaGasTx)
	if err != nil {
		log.Fatal(err)
	}

	//transfer from EOA to Hyperliquid Contract
	eoaToHLTx := TransferUSDC(client, privKey, amountX6, contractAddress)
	_, err = waitTransfer(client, eoaToHLTx)
	if err != nil {
		log.Fatal(err)
	}

	//check HL balance
	infoApi := NewInfoApi(&baseClient)
	state := infoApi.GetUserState(strings.Replace(eoaAddress, "0x", "", 1))
	balance := 0.0

	for balance <= 0 {
		state = infoApi.GetUserState(strings.Replace(eoaAddress, "0x", "", 1))
		strBalance := state.Withdrawable
		balance, _ = strconv.ParseFloat(strBalance, 32)
		fmt.Printf("User balance is %s\n", state.Withdrawable)
	}

}

func waitTransfer(client *ethclient.Client, tx string) (*types.Transaction, error) {
	var (
		t         *types.Transaction
		isPending bool
		err       error
	)

	isPending = true

	for isPending {
		t, isPending, err = client.TransactionByHash(context.Background(), common.HexToHash(tx))
		if isPending {
			log.Printf("Transaction %t still pending\n", isPending)
		}

		if err != nil {
			log.Fatalf("Transaction is in error %s\n", err.Error())
			return nil, err
		}
	}

	return t, nil

}

func TransferETH(client *ethclient.Client, privKey string, amountX18 *big.Int, to string) string {

	if strings.HasPrefix(privKey, "0x") {
		privKey = strings.Replace(privKey, "0x", "", 1)
	}

	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := amountX18
	gasLimit := uint64(2100000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(to)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
	return signedTx.Hash().Hex()
}

func TransferUSDC(client *ethclient.Client, privKey string, amountX6 *big.Int, to string) string {

	tokenAddress := common.HexToAddress("0x1870Dc7A474e045026F9ef053d5bB20a250Cc084")
	if strings.HasPrefix(privKey, "0x") {
		privKey = strings.Replace(privKey, "0x", "", 1)
	}

	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(0) // in wei (0 eth)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(to)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	paddedAmount := common.LeftPadBytes(amountX6.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gasLimit) // 23256

	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	res := signedTx.Hash().Hex()
	fmt.Printf("tx sent: %s\n", res)
	return res
}

func TestContractTransfer(t *testing.T) {

	client, err := ethclient.Dial("https://public.stackup.sh/api/v1/node/arbitrum-sepolia")
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("370d4ee26ce197dcee7227a00e721967f5bd250cdbb599315a5de9b6fb592e55")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(0) // in wei (0 eth)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0x279c9462fdba349550b49a23de27dd19d5891baa")
	tokenAddress := common.HexToAddress("0x1870Dc7A474e045026F9ef053d5bB20a250Cc084")

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	amount := new(big.Int)
	amount.SetString("5000000", 10) // sets the value to 1000 tokens, in the token denomination

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gasLimit) // 23256

	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex()) // tx sent: 0xa56316b637a94c4cc0331c73ef26389d6c097506d581073f927275e7a6ece0bc
}
