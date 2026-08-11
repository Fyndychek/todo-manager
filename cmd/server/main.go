package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Created   time.Time `json:"created_at"`
}

var (
	todos   = []Todo{}
	mu      sync.Mutex
	counter = 0
)

func main() {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	http.HandleFunc("/8", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Todo Api 1.0.0")
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}
