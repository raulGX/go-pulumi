package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type valueResponse struct {
	Value float32 `json:"value"`
}
type averageResponse struct {
	Average float32 `json:"average"`
}
type bitcoinResponse struct {
	USD float32 `json:"USD"`
}

func handleCurrentValue(t Tracker) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			value, err := t.Value()
			if err != nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(valueResponse{value})

		},
	)
}

func handleAverage(t Tracker) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			average := t.Average()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(averageResponse{average})
		},
	)
}

func main() {
	mux := http.NewServeMux()
	t := newRingBuffer(10 * 60 / 10)
	setValue := func() {
		resp, err := http.Get("https://min-api.cryptocompare.com/data/price?fsym=BTC&tsyms=USD")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var data bitcoinResponse
		err = json.Unmarshal(body, &data)
		if err != nil {
			panic(err)
		}
		t.Update(data.USD)
	}

	setValue()

	done := make(chan bool)
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				go func() {
					setValue()
				}()
			}
		}
	}()

	mux.Handle("/value", handleCurrentValue(t))
	mux.Handle("/average", handleAverage(t))
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("127.0.0.1", getenv("PORT", "8080")),
		Handler: mux,
	}

	log.Printf("listening on %s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	}

}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
