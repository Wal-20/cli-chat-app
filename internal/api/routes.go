package api

import (
	"fmt"
	"net/http"
	"strconv"
	"github.com/Wal-20/cli-chat-app/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"log"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var users []User

func getUserByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Path[len("/test/users/"):]

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID must be a number", http.StatusBadRequest)
		return
	}

	if id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if id > len(users) {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	user := users[id - 1]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gin.H{"User": user})
}

// GetUsers handler to fetch all users
func getUsers(w http.ResponseWriter, r *http.Request) {
	// Check if there are users
	if len(users) == 0 {
		http.Error(w, "No users found", http.StatusNotFound)
		return
	}

	// Return all users as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gin.H{"Users": users})
}

// AddUser handler to add a new user
func addUser(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into the user struct
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.ID <= 0 {
		user.ID = len(users) + 1
	}

	if user.Name == "" {
		http.Error(w, "User name is required", http.StatusBadRequest)
		return
	}

	users = append(users, user)

	// Return success message and the created user as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gin.H{
		"Status": "User created successfully",
		"User":   user,
	})
}

func NewServer() {
	// Create a router
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/users", handlers.GetUsers)
	mux.HandleFunc("POST /api/users", handlers.CreateUser)

	// test routes 
	mux.HandleFunc("GET /test/users/{id}", getUserByID)
	mux.HandleFunc("GET /test/users", getUsers)
	mux.HandleFunc("POST /test/users", addUser)

	// Start the server on port 8080
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

