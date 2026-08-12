package main

import (
	"encoding/json"
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Todo Api 1.0.0")
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("GET /todos/stats", getStats)
	//CRUD
	http.HandleFunc("GET /todos", getTodos)
	http.HandleFunc("POST /todos", createTodo)
	http.HandleFunc("DELETE /todos/{id}", deleteTodo)

	log.Printf("Server on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}

func getTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t Todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Printf("Decode error: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	counter++
	t.ID = counter
	t.Created = time.Now()
	todos = append(todos, t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	var t Todo
	idstring := r.PathValue("id")
	id, err := strconv.Atoi(idstring)
	if err != nil {
		log.Printf("Invalid Id: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	for i := 0; i < len(todos); i++ {
		if todos[i].ID == id {
			t = todos[i]
			todos = append(todos[:i], todos[i+1:]...)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}

func getStats(w http.ResponseWriter, r *http.Request) {
	var completed int
	mu.Lock()
	defer mu.Unlock()
	//for i := 0; i < len(todos); i++
	for _, t := range todos {
		if t.Completed == true {
			completed++
		}
	}
	stats := struct {
		Total     int `json:"total"`
		Completed int `json:"completed"`
		Pending   int `json:"pending"`
	}{
		Total:     len(todos),
		Completed: completed,
		Pending:   len(todos) - completed,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
