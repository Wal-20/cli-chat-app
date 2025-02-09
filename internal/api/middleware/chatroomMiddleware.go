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
		var userChatroom models.UserChatroom
		if err := config.DB.Where("user_id = ? AND chatroom_id = ? AND is_joined = ?", userID, chatroomID, true).First(&userChatroom).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "You are not a member of this chatroom", http.StatusForbidden)
			} else {
				http.Error(w, "Error retrieving user-chatroom association", http.StatusInternalServerError)
			}
			return
		}

		ctx := context.WithValue(r.Context(), "isAdmin", userChatroom.IsAdmin) // store admin status
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// get the authenticated user from the context, check if the user is part of the chatroom, and add the user's admin status
