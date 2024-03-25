package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
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
