package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type API interface {
	Post(context context.Context, path string, payload any) any
	IsMainnet() bool
}

type APIDefault struct {
	baseUrl    string
	httpClient *http.Client
	logger     Logger
}

func NewApiDefault(baseUrl string, logger Logger) API {
	httpClient := &http.Client{}
	return &APIDefault{
		baseUrl:    baseUrl,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (a *APIDefault) Post(ctx context.Context, path string, payload any) any {
	apiUrl := fmt.Sprintf("%s%s", a.baseUrl, path)
	body, _ := json.Marshal(payload)

	a.logger.LogInfo(ctx, fmt.Sprintf("Request body is %s", body))

	bodyReader := bytes.NewReader(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", apiUrl, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	var result any
	a.logger.LogInfo(ctx, fmt.Sprintf("Resp status: %s", resp.Status))

	errConversion := json.NewDecoder(resp.Body).Decode(&result)
	if errConversion != nil {
		a.logger.LogErr(ctx, "Failed to parse response body", errConversion)
		panic("Failed to parse response body")
	}
	return result
}

func (a *APIDefault) IsMainnet() bool {
	return a.baseUrl == MainnetUrl
}
