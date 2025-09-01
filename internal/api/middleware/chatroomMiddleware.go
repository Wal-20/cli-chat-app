package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"
	"net/http"
	"strconv"

	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"gorm.io/gorm"
)

func ChatroomMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from JWT (or request context, assuming auth middleware is already used)
		userID := r.Context().Value("userID").(uint)
		// Get chatroom ID from the request (query params, headers, or body)

		chatroomIDStr := r.PathValue("id")
		cacheKey := fmt.Sprintf("membership:%v:%s", userID, chatroomIDStr)
		var isMember bool

		if isMemberVal, found := utils.MembershipCache.Get(cacheKey); found {
			if isMember, ok := isMemberVal.(bool); ok && isMember {
				next.ServeHTTP(w, r)
				return
			}
		}
	
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
		isMember = true
		
		utils.MembershipCache.Set(cacheKey, isMember, time.Minute * 5)
		ctx := context.WithValue(r.Context(), "isAdmin", userChatroom.IsAdmin) // store admin status
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// get the authenticated user from the context, check if the user is part of the chatroom, and add the user's admin status
