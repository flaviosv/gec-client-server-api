package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	Filename = "cotacao.txt"
)

type ExchangeRate struct {
	Bid float64
}

func main() {
	rate, err := getExchangeRate()
	if err != nil {
		panic(err)
	}

	var exchangeRate ExchangeRate
	err = json.Unmarshal(rate, &exchangeRate)
	if err != nil {
		panic(err)
	}

	writeExchangeOnFile(exchangeRate)
}

func getExchangeRate() ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func writeExchangeOnFile(exchange ExchangeRate) {
	_, err := os.ReadFile(Filename)

	if err != nil {
		os.WriteFile(Filename, []byte(""), 0666)
	}

	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("Dolar: %f\n", exchange.Bid))
}
