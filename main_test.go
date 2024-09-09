package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestLoginHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/login", loginHandler).Methods("POST")

	creds := User{Username: "user1", Password: "password1"}
	body, _ := json.Marshal(creds)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]string
	json.NewDecoder(rr.Body).Decode(&response)
	assert.NotEmpty(t, response["token"])
}

func TestGetTodosHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/todos", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/todos", getTodosHandler).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCreateTodoHandler(t *testing.T) {
	newTodo := Todo{
		Title:       "Test Todo",
		Description: "Test Description",
		DueDate:     time.Now().Add(24 * time.Hour),
	}
	body, _ := json.Marshal(newTodo)

	req, _ := http.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/todos", createTodoHandler).Methods("POST")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var createdTodo Todo
	json.NewDecoder(rr.Body).Decode(&createdTodo)
	assert.Equal(t, newTodo.Title, createdTodo.Title)
	assert.Equal(t, newTodo.Description, createdTodo.Description)
}

func TestGetTodoHandler(t *testing.T) {
	todo := Todo{
		ID:          uuid.New(),
		Title:       "Test Todo",
		Description: "Test Description",
		DueDate:     time.Now().Add(24 * time.Hour),
	}
	todoList = append(todoList, todo)

	req, _ := http.NewRequest("GET", "/api/todos/"+todo.ID.String(), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/todos/{id}", getTodoHandler).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var fetchedTodo Todo
	json.NewDecoder(rr.Body).Decode(&fetchedTodo)
	assert.Equal(t, todo.ID, fetchedTodo.ID)
}

func TestUpdateTodoHandler(t *testing.T) {
	todo := Todo{
		ID:          uuid.New(),
		Title:       "Test Todo",
		Description: "Test Description",
		DueDate:     time.Now().Add(24 * time.Hour),
	}
	todoList = append(todoList, todo)

	updatedTodo := Todo{
		Title:       "Updated Todo",
		Description: "Updated Description",
		DueDate:     time.Now().Add(48 * time.Hour),
	}
	body, _ := json.Marshal(updatedTodo)

	req, _ := http.NewRequest("PUT", "/api/todos/"+todo.ID.String(), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/todos/{id}", updateTodoHandler).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var fetchedTodo Todo
	json.NewDecoder(rr.Body).Decode(&fetchedTodo)
	assert.Equal(t, updatedTodo.Title, fetchedTodo.Title)
	assert.Equal(t, updatedTodo.Description, fetchedTodo.Description)
}

func TestDeleteTodoHandler(t *testing.T) {
	todo := Todo{
		ID:          uuid.New(),
		Title:       "Test Todo",
		Description: "Test Description",
		DueDate:     time.Now().Add(24 * time.Hour),
	}
	todoList = append(todoList, todo)

	req, _ := http.NewRequest("DELETE", "/api/todos/"+todo.ID.String(), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/todos/{id}", deleteTodoHandler).Methods("DELETE")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestAuthMiddleware(t *testing.T) {
	claims := &Claims{
		Username: "user1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(jwtKey)

	req, _ := http.NewRequest("GET", "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.Handle("/api/todos", authMiddleware(http.HandlerFunc(getTodosHandler))).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLoggingMiddleware(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/todos", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/todos", getTodosHandler).Methods("GET")
	router.Use(loggingMiddleware)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
