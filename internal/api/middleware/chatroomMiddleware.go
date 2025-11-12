package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"gorm.io/gorm"
)

// membershipInfo is cached to avoid frequent DB lookups and to carry admin flag.
type membershipInfo struct {
	IsMember bool
	IsAdmin  bool
}

func ChatroomMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from JWT (or request context, assuming auth middleware is already used)
		userID := r.Context().Value("userID").(uint)
		// Get chatroom ID from the request (query params, headers, or body)

		chatroomIDStr := r.PathValue("id")
		cacheKey := fmt.Sprintf("membership:%v:%s", userID, chatroomIDStr)

		if cached, found := utils.MembershipCache.Get(cacheKey); found {
			if info, ok := cached.(membershipInfo); ok && info.IsMember {
				// Ensure admin status is propagated to downstream handlers
				ctx := context.WithValue(r.Context(), "isAdmin", info.IsAdmin)
				next.ServeHTTP(w, r.WithContext(ctx))
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
		if err := config.DB.Where("user_id = ? AND chatroom_id = ? AND is_joined = ? AND is_banned = ?", userID, chatroomID, true, false).First(&userChatroom).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "You are not a member of this chatroom", http.StatusForbidden)
			} else {
				http.Error(w, "Error retrieving user-chatroom association", http.StatusInternalServerError)
			}
			return
		}
		// Cache both membership and admin flag (owner has admin capabilities)
		adminOrOwner := userChatroom.IsAdmin || userChatroom.IsOwner
		utils.MembershipCache.Set(cacheKey, membershipInfo{IsMember: true, IsAdmin: adminOrOwner}, time.Minute*5)
		ctx := context.WithValue(r.Context(), "isAdmin", adminOrOwner) // store admin status
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// get the authenticated user from the context, check if the user is part of the chatroom, and add the user's admin status
