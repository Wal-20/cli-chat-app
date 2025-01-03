package api

import (
	"fmt"
	"net/http"
	"github.com/Wal-20/cli-chat-app/internal/api/handlers"
	"github.com/Wal-20/cli-chat-app/internal/api/middleware"
	"log"
)


func NewServer() {
	// Create a router
	mux := http.NewServeMux()

	// User routes
	mux.HandleFunc("GET /api/users", handlers.GetUsers)
	mux.HandleFunc("POST /api/users", handlers.CreateUser)
	mux.HandleFunc("POST /api/users/login", handlers.Login)
    mux.Handle("POST /api/users/logout", middleware.AuthMiddleware(http.HandlerFunc(handlers.LogOut)))
    mux.Handle("POST /api/users/update", middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdateUser)))

/* 	mux.HandleFunc("GET /api/chatrooms", handlers.GetChatrooms)
	mux.HandleFunc("POST /api/chatrooms", handlers.CreateChatroom)
*/

	// Start the server on port 8080
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

