package api

import (
	"encoding/json"
	"fmt"
	"github.com/Wal-20/cli-chat-app/internal/api/handlers"
	"github.com/Wal-20/cli-chat-app/internal/api/middleware"
	"log"
	"net/http"
)

func NewServer() {

	mux := http.NewServeMux()

	// initialize handler dependencies (services)
	handlers.InitHandlers()

	// Health
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"Status": "Server connection established",
		})
	})

	// Install Client Route
	mux.HandleFunc("/install.sh", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./install.sh")
	})

	// User routes
	mux.HandleFunc("GET /api/users", handlers.GetUsers)
	mux.HandleFunc("POST /api/users", handlers.CreateUser)
	mux.HandleFunc("POST /api/users/login", handlers.Login)
	mux.HandleFunc("POST /api/users/logout", handlers.LogOut)
	mux.HandleFunc("POST /api/users/refresh", handlers.RefreshToken)
	mux.Handle("POST /api/users/update", middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdateUser)))
	mux.Handle("GET /api/users/chatrooms", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetChatroomsByUser)))
	mux.Handle("GET /api/users/notifications", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetNotifications)))

	// Admin routes
	mux.Handle("POST /api/users/chatrooms/{id}/invite/{userId}", middleware.AuthMiddleware(
		middleware.ChatroomMiddleware(
			http.HandlerFunc(handlers.InviteUser),
		),
	))
	mux.Handle("POST /api/users/chatrooms/{id}/kick/{userId}", middleware.AuthMiddleware(
		middleware.ChatroomMiddleware(
			http.HandlerFunc(handlers.KickUser),
		),
	))
	mux.Handle("POST /api/users/chatrooms/{id}/ban/{userId}", middleware.AuthMiddleware(
		middleware.ChatroomMiddleware(
			http.HandlerFunc(handlers.BanUser),
		),
	))

	// Chatroom routes
	mux.HandleFunc("GET /api/chatrooms", handlers.GetChatrooms)
	mux.Handle("GET /api/chatrooms/public", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetPublicChatrooms)))
	mux.Handle("POST /api/chatrooms", middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateChatroom)))
	mux.Handle("DELETE /api/chatrooms/{id}", middleware.AuthMiddleware(
		middleware.ChatroomMiddleware(
			http.HandlerFunc(handlers.DeleteChatroom),
		),
	))
	mux.HandleFunc("GET /api/chatrooms/{id}/users", handlers.GetUsersByChatroom)
	mux.HandleFunc("GET /api/chatrooms/{id}/messages", handlers.GetMessagesByChatroom)
	mux.Handle("POST /api/chatrooms/{id}/join", middleware.AuthMiddleware(http.HandlerFunc(handlers.JoinChatroom)))
	mux.Handle("POST /api/chatrooms/{id}/leave", middleware.AuthMiddleware(
		middleware.ChatroomMiddleware(
			http.HandlerFunc(handlers.LeaveChatroom),
		),
	))

	// Message routes
	mux.Handle("POST /api/chatrooms/{id}/messages",
		middleware.AuthMiddleware(
			middleware.ChatroomMiddleware(
				http.HandlerFunc(handlers.SendMessage),
			),
		),
	)
	mux.Handle("DELETE /api/chatrooms/{id}/messages/{messageId}",
		middleware.AuthMiddleware(
			middleware.ChatroomMiddleware(
				http.HandlerFunc(handlers.DeleteMessage),
			),
		),
	)
	mux.Handle("POST /api/chatrooms/{id}/messages/{messageId}",
		middleware.AuthMiddleware(
			middleware.ChatroomMiddleware(
				http.HandlerFunc(handlers.UpdateMessage),
			),
		),
	)

	// Notification routes
	mux.Handle("DELETE /api/notifications/{id}",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.DeleteNotification),
		),
	)

	// WebSocket routes
	mux.Handle("GET /api/chatrooms/{id}/ws",
		middleware.AuthMiddleware(
			middleware.ChatroomMiddleware(
				http.HandlerFunc(handlers.ChatroomWebSocket),
			),
		),
	)

	// Start the server on port 8080
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", middleware.CheckCORS(mux)))
}
