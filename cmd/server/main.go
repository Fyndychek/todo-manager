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
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	db *sql.DB
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type contextKey string

const userIDKey contextKey = "userID"

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {

	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-secret-change-me")
	}
}

func generateToken(userID int64) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(jwtSecret)
}

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
	http.HandleFunc("/neiroslop2", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/neiroslop2" {
			http.NotFound(w, r)
			return
		}
		// Получаем текущую рабочую директорию
		wd, _ := os.Getwd()
		filePath := filepath.Join(wd, "cmd", "sandbox", "index3.html")
		http.ServeFile(w, r, filePath)
	}))
	http.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	http.HandleFunc("GET /todos/stats", corsMiddleware(authMiddleware(getStats)))
	//CRUD
	http.HandleFunc("GET /todos", corsMiddleware(authMiddleware(getTodos)))
	http.HandleFunc("POST /todos", corsMiddleware(authMiddleware(createTodo)))
	http.HandleFunc("DELETE /todos/{id}", corsMiddleware(authMiddleware(deleteTodo)))
	http.HandleFunc("PATCH /todos/{id}", corsMiddleware(authMiddleware(updateTodo)))

	//auth
	http.HandleFunc("POST /register", corsMiddleware(registration))
	http.HandleFunc("POST /login", corsMiddleware(login))

	//
	//
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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// Парсим и проверяем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			respondError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Извлекаем user_id из claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respondError(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		userIDFloat, ok := claims["user_id"].(float64) // JWT числа приходят как float64
		if !ok {
			respondError(w, "Invalid user_id in token", http.StatusUnauthorized)
			return
		}
		userID := int64(userIDFloat)

		// Сохраняем user_id в контексте
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func respondError(w http.ResponseWriter, message string, status int) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func getTodos(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	if tfilter, err := storage.GetTodos(filter, sortby, order, db, userID); err != nil {
		log.Printf("Get task error: %v", err)
		respondError(w, "Get task error", http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tfilter)
	}

}

func createTodo(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	if resid, errDB = storage.AddTodo(t, db, userID); errDB != nil {
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

	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	todo, err := storage.UpdateTodo(id, upd, db, userID)
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

	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idstring := r.PathValue("id")
	id, err := strconv.Atoi(idstring)
	if err != nil {
		log.Printf("Invalid Id: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	//вызов слоя бд
	if t, err := storage.DeleteTodo(id, db, userID); err != nil {
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

	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var (
		total     int
		completed int
		pending   int
		err       error
	)

	//вызов слоя бд
	if total, completed, pending, err = storage.GetStats(db, userID); err != nil {
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

func registration(w http.ResponseWriter, r *http.Request) {

	var auth AuthRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		log.Printf("Decode error: %v", err)
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if len(auth.Password) < 4 {
		respondError(w, "Password is too short. It must been longer than 3 symbols", http.StatusBadRequest)
		return
	}
	if len(auth.Username) < 3 {
		respondError(w, "Username is too short. It must been longer than 2 symbols", http.StatusBadRequest)
		return
	}
	user_id, err := storage.CreateUser(auth.Username, auth.Password, db)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			respondError(w, "This name is already taken", http.StatusConflict)
			log.Printf("This name is already taken: %v", err)
			return
		}
		respondError(w, "Create user error", http.StatusInternalServerError)
		log.Printf("Create user error: %v", err)
		return

	}
	log.Printf("User created with ID: %d", user_id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := struct {
		User_id  int64  `json:"user_id"`
		Username string `json:"username"`
	}{
		User_id:  user_id,
		Username: auth.Username,
	}
	json.NewEncoder(w).Encode(resp)
}

func login(w http.ResponseWriter, r *http.Request) {

	var auth AuthRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		log.Printf("Decode error: %v", err)
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if len(auth.Password) < 4 {
		respondError(w, "Password is too short. It must been longer than 3 symbols", http.StatusBadRequest)
		return
	}
	if len(auth.Username) < 3 {
		respondError(w, "Password is too short. It must been longer than 2 symbols", http.StatusBadRequest)
		return
	}

	id, hash, err := storage.GetUserByUsername(auth.Username, db)
	if err == sql.ErrNoRows {
		respondError(w, "Login or password is incorrect", http.StatusUnauthorized)
		return
	}
	if err != nil {
		respondError(w, "Login service is not available", http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(auth.Password))
	if err != nil {
		respondError(w, "Login or password is incorrect", http.StatusUnauthorized)
		return
	}
	token, errorToken := generateToken(id)
	if errorToken != nil {
		respondError(w, "Login service is not available", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
