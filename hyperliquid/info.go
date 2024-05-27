package hyperliquid

import (
	"encoding/json"
	"strconv"
	"strings"
)

type InfoApi interface {
	GetUserState(address string) UserState
	GetUserFills(address string) []OrderFill
	GetNonFundingUpdates(address string) []NonFundingUpdate
	GetFundingUpdates(address string) []FundingUpdate
	GetWithdrawals(address string) []Withdrawal
	FindOrder(address string, cloid string) OrderResponse
	FindOpenOrders(address string) []OpenOrder
	GetAllMids() map[string]string
	GetMktPx(coin string) float64
	GetMeta() Meta
}

type InfoApiDefault struct {
	apiClient *API
}

func NewInfoApi(cli *API) InfoApi {
	return &InfoApiDefault{
		apiClient: cli,
	}
}

type UserState struct {
	Withdrawable       string          `json:"withdrawable"`
	AssetPositions     []AssetPosition `json:"assetPositions"`
	CrossMarginSummary MarginSummary   `json:"crossMarginSummary"`
	MarginSummary      MarginSummary   `json:"marginSummary"`
}

type AssetPosition struct {
	Position Position `json:"position"`
	Type     string   `json:"type"`
}

type Position struct {
	Coin           string   `json:"coin"`
	EntryPx        string   `json:"entryPx"`
	Leverage       Leverage `json:"leverage"`
	LiquidationPx  string   `json:"liquidationPx"`
	MarginUsed     string   `json:"marginUsed"`
	PositionValue  string   `json:"positionValue"`
	ReturnOnEquity string   `json:"returnOnEquity"`
	Szi            string   `json:"szi"`
	UnrealizedPnl  string   `json:"unrealizedPnl"`
}

type Leverage struct {
	Type   string  `json:"type"`
	Value  int     `json:"value"`
	RawUsd float64 `json:"rawUsd"`
}

type GetUserStateRequest struct {
	User  string `json:"user"`
	Typez string `json:"type"`
}

type GetInfoRequest struct {
	User  *string `json:"user,omitempty"`
	Typez string  `json:"type"`
	Oid   *string `json:"oid,omitempty"`
}

type MarginSummary struct {
	AccountValue    string `json:"accountValue"`
	TotalMarginUsed string `json:"totalMarginUsed"`
	TotalNtlPos     string `json:"totalNtlPos"`
	TotalRawUsd     string `json:"totalRawUsd"`
}

func (api *InfoApiDefault) GetUserState(address string) UserState {
	request := GetUserStateRequest{
		User:  address,
		Typez: "clearinghouseState",
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result UserState
	_ = json.Unmarshal(parsed, &result)
	return result
}

func (api *InfoApiDefault) GetAllMids() map[string]string {
	request := GetInfoRequest{
		Typez: "allMids",
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result map[string]string
	_ = json.Unmarshal(parsed, &result)

	for k, v := range result {
		result[strings.ToUpper(k)] = v
	}

	return result
}

type Meta struct {
	Universe []Asset `json:"universe"`
}

type Asset struct {
	Name       string `json:"name"`
	SzDecimals int    `json:"szDecimals"`
}

func (api *InfoApiDefault) GetMeta() Meta {
	request := GetInfoRequest{
		Typez: "meta",
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result Meta
	_ = json.Unmarshal(parsed, &result)
	return result
}

func (api *InfoApiDefault) FindOrder(address string, cloid string) OrderResponse {
	request := GetInfoRequest{
		User:  &address,
		Typez: "orderStatus",
		Oid:   &cloid,
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result OrderResponse
	_ = json.Unmarshal(parsed, &result)
	return result
}

func (api *InfoApiDefault) FindOpenOrders(address string) []OpenOrder {
	request := GetInfoRequest{
		User:  &address,
		Typez: "openOrders",
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result []OpenOrder
	_ = json.Unmarshal(parsed, &result)
	return result
}

func (api *InfoApiDefault) GetMktPx(coin string) float64 {
	parsed, _ := strconv.ParseFloat(api.GetAllMids()[coin], 32)
	return parsed
}

func (api *InfoApiDefault) GetUserFills(address string) []OrderFill {
	request := GetInfoRequest{
		User:  &address,
		Typez: "userFills",
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result []OrderFill
	_ = json.Unmarshal(parsed, &result)
	return result
}

func (api *InfoApiDefault) GetNonFundingUpdates(address string) []NonFundingUpdate {
	request := GetInfoRequest{
		User:  &address,
		Typez: "userNonFundingLedgerUpdates",
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result []NonFundingUpdate
	_ = json.Unmarshal(parsed, &result)
	return result
}

func (api *InfoApiDefault) GetFundingUpdates(address string) []FundingUpdate {
	request := GetInfoRequest{
		User:  &address,
		Typez: "userFunding",
	}
	anyResult := (*api.apiClient).Post("/info", request)
	parsed, _ := json.Marshal(anyResult)
	var result []FundingUpdate
	_ = json.Unmarshal(parsed, &result)
	return result
}

func (api *InfoApiDefault) GetWithdrawals(address string) []Withdrawal {
	var ws []Withdrawal
	ups := api.GetNonFundingUpdates(address)
	for _, up := range ups {
		if up.Delta.Type == "withdraw" {
			w := Withdrawal{
				Hash:   up.Hash,
				Amount: up.Delta.Amount,
				Fee:    up.Delta.Fee,
				Nonce:  up.Delta.Nonce,
			}
			ws = append(ws, w)
		}
	}
	return ws
}
