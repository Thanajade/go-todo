package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/grokify/goauth/authutil/jwtutil"
)

var jwtKey = []byte("super_secret_key")

// User struct for in-memory authentication
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Predefined users
var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

// JWT Claims struct
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Todo struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
}

var todoList = []Todo{}

// Middleware to validate JWT token
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")

		if tokenStr == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		// Extract Bearer from incoming token
		tokenStr = tokenStr[7:]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Login endpoint: Authenticate user and return JWT token
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds User
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	storedPassword, exists := users[creds.Username]
	if !exists || storedPassword != creds.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
}

// To-do CRUD Handlers

// Get all todos
func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoList)
}

// Create a new todo
// Create a new todo
func createTodoHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var newTodo Todo
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate a new UUID for the todo
	newTodo.ID = uuid.New()
	todoList = append(todoList, newTodo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newTodo)

	// delayTime(5)

	log.Printf("POST %s %s %v", r.RequestURI, r.RemoteAddr, time.Since(startTime))
}

// Get a single todo by ID
func getTodoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	for _, todo := range todoList {
		if todo.ID.String() == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(todo)
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}

// Update a todo by ID
func updateTodoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	for i, todo := range todoList {
		if todo.ID.String() == id {
			var updatedTodo Todo
			err := json.NewDecoder(r.Body).Decode(&updatedTodo)
			if err != nil {
				http.Error(w, "Invalid request payload", http.StatusBadRequest)
				return
			}

			todoList[i] = updatedTodo

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(todoList[i])
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}

// Delete a todo by ID
func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	for i, todo := range todoList {
		if todo.ID.String() == id {
			todoList = append(todoList[:i], todoList[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Todo not found", http.StatusNotFound)
}

// Middleware to log each incoming request
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(startTime))
	})
}

func createJWTHS256SignedString(secretKey string, data map[string]any) (string, error) {
	claims := map[string]any{
		"iss": "issuer",
		"exp": time.Now().Add(time.Hour).Unix(),
		"data": map[string]any{
			"id":   "123",
			"name": "JohnDoe",
		},
	}
	return jwtutil.CreateHS256SignedString([]byte(secretKey), claims)
}

// Delay or sleep for t amount of time
func delayTime(delayTime int) {
	t := time.Duration(delayTime) * time.Second
	time.Sleep(t)
	fmt.Fprintf(os.Stdout, "Delayed for %d seconds", delayTime)
}

func main() {
	// Create a log file
	logFile, err := os.OpenFile("requests.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not open log file:", err)
	}
	defer logFile.Close()

	// Set the output of logs to the file
	log.SetOutput(logFile)

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/login", loginHandler).Methods("POST")

	// Authenticated routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware)

	api.HandleFunc("/todos", getTodosHandler).Methods("GET")
	api.HandleFunc("/todos", createTodoHandler).Methods("POST")
	api.HandleFunc("/todos/{id}", getTodoHandler).Methods("GET")
	api.HandleFunc("/todos/{id}", updateTodoHandler).Methods("PUT")
	api.HandleFunc("/todos/{id}", deleteTodoHandler).Methods("DELETE")

	// Apply logging middleware globally
	r.Use(loggingMiddleware)

	// Start server
	log.Println("Server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
