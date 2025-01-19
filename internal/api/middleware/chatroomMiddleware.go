package middleware

import (
	"net/http"
	"context"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/config"
	"gorm.io/gorm"
	"errors"
	"strconv"
)

func ChatroomMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from JWT (or request context, assuming auth middleware is already used)
		userID := r.Context().Value("userID").(uint)

		// Get chatroom ID from the request (query params, headers, or body)
		chatroomIDStr := r.PathValue("id")
		if chatroomIDStr == "" {
			http.Error(w, "Chatroom ID is required", http.StatusBadRequest)
			return
		}

		chatroomID, err := strconv.Atoi(chatroomIDStr)
		if err != nil {
			http.Error(w, "Invalid chatroom ID format", http.StatusBadRequest)
			return
		}

		// Verify user membership in the chatroom
		var chatroom models.Chatroom
		if err := config.DB.Preload("Users").First(&chatroom, chatroomID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Chatroom not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error retrieving chatroom", http.StatusInternalServerError)
			}
			return
		}

		// Check if the user is a member of the chatroom
		isMember := false
		for _, user := range chatroom.Users {
			if user.ID == userID {
				isMember = true
				break
			}
		}

		if !isMember {
			http.Error(w, "You are not a member of this chatroom", http.StatusForbidden)
			return
		}

		// Store chatroom ID in context
		ctx := context.WithValue(r.Context(), "chatroomID", chatroomID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// all this middleware does is get the authenticated user from the context, checks if the user is part of the chatroom, and adds that chatroom to the context
