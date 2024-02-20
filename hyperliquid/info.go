package hyperliquid

import (
	"encoding/json"
	"strconv"
)

type InfoApi interface {
	GetUserState(address string) UserState
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
	User  string `json:"user"`
	Typez string `json:"type"`
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

func (api *InfoApiDefault) GetMktPx(coin string) float64 {
	parsed, _ := strconv.ParseFloat(api.GetAllMids()[coin], 32)
	return parsed
}
