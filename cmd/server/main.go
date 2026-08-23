package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/Fyndychek/todo-manager/internal/handlers"
	"github.com/Fyndychek/todo-manager/internal/storage"
	_ "github.com/mattn/go-sqlite3"
)

// структура задачи
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Created   time.Time `json:"created_at"`
}

type SortGet struct {
	Filter *string `json:"filter"`
	Sort   *string `json:"sort"`
	Order  *string `json:"order"`
}

var (
	db *sql.DB
)

func main() {

	var err error
	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		dbFile = "todos.db"
	}
	db, err = storage.NewDatabase(dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	http.HandleFunc("/neiroslop", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/neiroslop" {
			http.NotFound(w, r)
			return
		}
		// Получаем текущую рабочую директорию
		wd, _ := os.Getwd()
		filePath := filepath.Join(wd, "cmd", "sandbox", "index2.html")
		http.ServeFile(w, r, filePath)
	}))
	http.HandleFunc("/handmade", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/handmade" {
			http.NotFound(w, r)
			return
		}
		// Получаем текущую рабочую директорию
		wd, _ := os.Getwd()
		filePath := filepath.Join(wd, "cmd", "sandbox", "index.html")
		http.ServeFile(w, r, filePath)
	}))
	http.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	http.HandleFunc("GET /todos/stats", corsMiddleware(getStats))

	//CRUD
	http.HandleFunc("GET /todos", corsMiddleware(getTodos))
	http.HandleFunc("POST /todos", corsMiddleware(createTodo))
	http.HandleFunc("DELETE /todos/{id}", corsMiddleware(deleteTodo))
	http.HandleFunc("PATCH /todos/{id}", corsMiddleware(updateTodo))

	//запуск
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: nil,
	}

	go func() {
		log.Printf("Server starting on :%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	//прием управляющих сигналов
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited gracefully")

	if err := db.Close(); err != nil {
		log.Printf("Error closing DB: %v", err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Добавляем заголовки CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Если это preflight OPTIONS-запрос, отвечаем сразу, не передавая дальше
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Иначе передаём управление основному обработчику
		next(w, r)
	}
}

func respondError(w http.ResponseWriter, message string, status int) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func getTodos(w http.ResponseWriter, r *http.Request) {

	filter := r.URL.Query().Get("filter")
	sortby := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	if filter == "" {
		filter = "all"
	}
	if sortby == "" {
		sortby = "created"
	}
	if order == "" {
		order = "asc"
	}

	if len(filter) > 255 {
		log.Printf("too long request")
		respondError(w, "too long request(max 255)", http.StatusBadRequest)
		return
	}

	if len(sortby) > 255 {
		log.Printf("too long request")
		respondError(w, "too long request(max 255)", http.StatusBadRequest)
		return
	}

	//вызов слоя бд
	if tfilter, err := storage.GetTodos(filter, sortby, order, db); err != nil {
		log.Printf("Get task error: %v", err)
		respondError(w, "Get task error", http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tfilter)
	}

}

func createTodo(w http.ResponseWriter, r *http.Request) {

	var (
		t     storage.DBtodo
		resid int64
		errDB error
	)
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Printf("Decode error: %v", err)
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if len(t.Title) > 255 {
		log.Printf("too long request")
		respondError(w, "too long request(max 255)", http.StatusBadRequest)
		return
	}
	//проблема с чтением тела
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		log.Printf("Empty request")
		respondError(w, "Empty request", http.StatusBadRequest)
		return
	}

	//вызов слоя бд
	t.Created = time.Now()
	if resid, errDB = storage.AddTodo(t, db); errDB != nil {
		log.Printf("Add task error")
		respondError(w, "Add task error", http.StatusInternalServerError)
		return

	}
	t.ID = resid
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {

	var upd storage.UpdateTodoRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		log.Printf("Decode error: %v", err)
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	//извлечение id
	idstring := r.PathValue("id")
	id, err := strconv.Atoi(idstring)
	if err != nil {
		log.Printf("Invalid Id: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	//проверка контента тайтла задачи
	if upd.Title != nil {
		*upd.Title = strings.TrimSpace(*upd.Title)
		if *upd.Title == "" {
			log.Printf("Empty request")
			respondError(w, "Empty request", http.StatusBadRequest)
			return
		}
		if len(*upd.Title) > 255 {
			log.Printf("too long request")
			respondError(w, "too long request(max 255)", http.StatusBadRequest)
			return
		}
	}

	//вызов слоя бд
	todo, err := storage.UpdateTodo(id, upd, db)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, "Not Found", http.StatusNotFound)
			log.Printf("Not Found ID")
			return
		}
		respondError(w, "Update error", http.StatusInternalServerError)
		log.Printf("Update error: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todo)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {

	idstring := r.PathValue("id")
	id, err := strconv.Atoi(idstring)
	if err != nil {
		log.Printf("Invalid Id: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	//вызов слоя бд
	if t, err := storage.DeleteTodo(id, db); err != nil {
		if err == sql.ErrNoRows {
			respondError(w, "Not Found", http.StatusNotFound)
			log.Printf("Not Found ID")
			return
		} else {
			respondError(w, "Delete error", http.StatusInternalServerError)
			log.Printf("Delete error: %v", err)
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(t)
		return

	}
}

func getStats(w http.ResponseWriter, r *http.Request) {

	var (
		total     int
		completed int
		pending   int
		err       error
	)

	//вызов слоя бд
	if total, completed, pending, err = storage.GetStats(db); err != nil {
		respondError(w, "Get stats error", http.StatusInternalServerError)
		log.Printf("Get stats error: %v", err)
		return
	}

	stats := struct {
		Total     int `json:"total"`
		Completed int `json:"completed"`
		Pending   int `json:"pending"`
	}{
		Total:     total,
		Completed: completed,
		Pending:   pending,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
