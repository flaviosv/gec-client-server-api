package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"time"
)

type ExchangeRate struct {
	Rate ExchangeRateDetail `json:"USDBRL"`
}

type ExchangeRateDetail struct {
	ID         int `json:"-"`
	Code       string
	Codein     string
	Name       string
	High       float64 `json:",string"`
	Low        float64 `json:",string"`
	VarBid     float64 `json:",string"`
	PctChange  float64 `json:",string"`
	Bid        float64 `json:",string"`
	Ask        float64 `json:",string"`
	Timestamp  int     `json:",string"`
	CreateDate string  `json:"create_date"`
}

type ExchangeRateResponse struct {
	Bid float64
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", getExchangeRate)
	http.ListenAndServe(":8080", mux)
}

func getExchangeRate(w http.ResponseWriter, r *http.Request) {
	rate, err := getDollarExchangeRate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	var exchangeRate ExchangeRate
	err = json.Unmarshal(rate, &exchangeRate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	db, err := sql.Open("sqlite3", "./exchange.db")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = existRateByTimestamp(db, exchangeRate.Rate.Timestamp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	saveExchangeRate(db, exchangeRate)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := ExchangeRateResponse{Bid: exchangeRate.Rate.Bid}
	json.NewEncoder(w).Encode(response)
}

func getDollarExchangeRate() ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

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

func saveExchangeRate(db *sql.DB, rate ExchangeRate) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	stmt, err := db.PrepareContext(ctx, `
			INSERT INTO exchange_rates
			      	(code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date)
			      VALUES
			      	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				
      `)
	if err != nil {
		panic(err)
	}
	defer cancel()

	_, err = stmt.Exec(
		rate.Rate.Code,
		rate.Rate.Codein,
		rate.Rate.Name,
		rate.Rate.High,
		rate.Rate.Low,
		rate.Rate.VarBid,
		rate.Rate.PctChange,
		rate.Rate.Bid,
		rate.Rate.Ask,
		rate.Rate.Timestamp,
		rate.Rate.CreateDate,
	)
	if err != nil {
		return err
	}

	return nil
}

func existRateByTimestamp(db *sql.DB, timestamp int) error {
	stmt, err := db.Prepare("SELECT id FROM exchange_rates WHERE timestamp = ? LIMIT 1")
	if err != nil {
		return err
	}

	var id int
	err = stmt.QueryRow(timestamp).Scan(&id)
	if err != nil {
		return err
	}

	if id >= 1 {
		return nil
	}

	return errors.New("Unexpected error")
}
