package hyperliquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type API interface {
	Post(path string, payload any) any
	IsMainnet() bool
}

type APIDefault struct {
	baseUrl    string
	httpClient *http.Client
}

func NewApiDefault(baseUrl string) API {
	httpClient := &http.Client{}
	return &APIDefault{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}

func (a *APIDefault) Post(path string, payload any) any {
	apiUrl := fmt.Sprintf("%s%s", a.baseUrl, path)
	body, _ := json.Marshal(payload)

	bodyReader := bytes.NewReader(body)
	req, _ := http.NewRequest("POST", apiUrl, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	var result map[string]interface{}
	fmt.Printf("Resp status: %s\n", resp.Status)

	errConversion := json.NewDecoder(resp.Body).Decode(&result)
	if errConversion != nil {
		panic("Failed to parse response body")
	}
	return result
}

func (a *APIDefault) IsMainnet() bool {
	return a.baseUrl == MainnetUrl
}
