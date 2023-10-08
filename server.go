package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type ExchangeRate struct {
	Rate ExchangeRateDetail `json:"USDBRL"`
}

type ExchangeRateDetail struct {
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

func main() {
	getDollarExchangeRate()
}

func getDollarExchangeRate() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Println(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(body))

	var exchangeRate ExchangeRate
	err = json.Unmarshal(body, &exchangeRate)
	if err != nil {
		log.Println(err)
	}

	log.Println(exchangeRate)
}
