package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil || id <= 0 {

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]any{"Status": "invalid ID"})
			return
		}

		var user models.User
		result := config.DB.First(&user, id)
		if result.Error != nil {

			w.WriteHeader(http.StatusNotFound)
			encoder.Encode(map[string]any{"Status": "User not found"})
			return
		}

		w.WriteHeader(http.StatusOK)
		encoder.Encode(map[string]any{
			"User": user,
		})

	} else {
		var users []models.User
		result := config.DB.Find(&users)

		if result.Error != nil {
			encoder.Encode(map[string]any{"Message": "Failed to retrieve users"})
			return
		}
		encoder.Encode(map[string]any{"Users": users})
	}
}

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	payload, err := Svcs.Notification.GetUserNotifications(userID)
	if err != nil {
		http.Error(w, "Error retrieving notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func GetChatroomsByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)

	var chatrooms []models.Chatroom
	// Only return rooms the user actually joined
	if err := config.DB.
		Joins("JOIN user_chatrooms ON user_chatrooms.chatroom_id = chatrooms.id").
		Where("user_chatrooms.user_id = ? AND user_chatrooms.is_joined = ?", userID, true).
		Find(&chatrooms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "No chatrooms found for user", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving Chatrooms", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"Chatrooms": chatrooms,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body models.User
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate user inputs
	if body.Name == "" || body.Password == "" {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	accessToken, refreshToken, _, err := Svcs.Auth.Login(body.Name, body.Password)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if err.Error() == "invalid password" {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]any{
		"Status":       "success",
		"AccessToken":  accessToken,
		"RefreshToken": refreshToken,
	})
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	// Stateless: instruct the client to clear its local tokens; drop cached claims if present.
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		utils.AuthCache.Delete(strings.TrimPrefix(authHeader, "Bearer "))
	}
	json.NewEncoder(w).Encode(map[string]any{
		"Status": "Logged out successfully",
	})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := config.DB.Delete(&models.User{}, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user", http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"Status": "User Deleted successfully",
	})
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var body models.User

	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if body.Name == "" || body.Password == "" {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}
	accessToken, refreshToken, user, err := Svcs.Auth.Register(body.Name, body.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]any{
		"Status":       "User created successfully",
		"User":         user,
		"AccessToken":  accessToken,
		"RefreshToken": refreshToken,
	})

}

// RefreshToken accepts a refresh token and returns a new access/refresh token pair.
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	refresh := body["refreshToken"]
	if refresh == "" {
		refresh = body["RefreshToken"]
	}
	if refresh == "" {
		refresh = body["refresh_token"]
	}
	if refresh == "" {
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	claims, err := utils.ValidateJWTToken(refresh)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	userIDFloat, ok := claims["userID"].(float64)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	newAccess, err := utils.GenerateJWTToken(uint(userIDFloat), username)
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}
	newRefresh, err := utils.GenerateRefreshToken(uint(userIDFloat), username)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"Status":       "success",
		"AccessToken":  newAccess,
		"RefreshToken": newRefresh,
	})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user", http.StatusInternalServerError)
		}
		return
	}

	var userUpdate models.User
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&userUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if userUpdate.Name == "" || userUpdate.Password == "" {
		http.Error(w, "Missing attributes", http.StatusBadRequest)
		return
	}
	user.Name = userUpdate.Name

	hashedPassword, err := utils.HashPassword(userUpdate.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	config.DB.Save(&user)

	encoder.Encode(map[string]any{
		"Status": "User Updated successfully",
		"User":   user,
	})
}

// admin actions
func InviteUser(w http.ResponseWriter, r *http.Request) {
	userIdent := r.PathValue("userId")
	if userIdent == "" {
		http.Error(w, "No valid user identifier provided", http.StatusBadRequest)
		return
	}

	chatroomId := r.PathValue("id")
	if chatroomId == "" {
		http.Error(w, "No valid chatroom ID provided", http.StatusBadRequest)
		return
	}

	// Resolve user by numeric ID or by username
	var userIdNum uint64
	if idNum, err := strconv.ParseUint(userIdent, 10, 32); err == nil {
		userIdNum = idNum
	} else {
		var u models.User
		if err := config.DB.Where("name = ?", userIdent).First(&u).Error; err != nil {
			http.Error(w, "Invalid user identifier: not found", http.StatusNotFound)
			return
		}
		userIdNum = uint64(u.ID)
	}

	chatroomIdNum, err := strconv.ParseUint(chatroomId, 10, 32)
	if err != nil {
		http.Error(w, "Invalid chatroom ID, cannot convert to number", http.StatusBadRequest)
		return
	}

	isAdmin := r.Context().Value("isAdmin").(bool)
	if !isAdmin {
		http.Error(w, "You are not an admin", http.StatusUnauthorized)
		return
	}

	var admin models.User
	if err := config.DB.First(&admin, r.Context().Value("userID")).Error; err != nil {
		http.Error(w, "Error fetching admin", http.StatusInternalServerError)
		return
	}

	var targetUser models.User
	if err := config.DB.First(&targetUser, uint(userIdNum)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	var userChatroom models.UserChatroom
	err = config.DB.Where("user_id = ? AND chatroom_id = ?", userIdNum, chatroomIdNum).First(&userChatroom).Error
	expiration := time.Now().Add(7 * 24 * time.Hour) // 7 days

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userChatroom = models.UserChatroom{
				UserID:        targetUser.ID,
				Name:          targetUser.Name,
				ChatroomID:    uint(chatroomIdNum),
				LastJoinTime:  nil,
				IsInvited:     true,
				IsJoined:      false,
				InviteExpires: &expiration,
			}
			if err := config.DB.
				Create(&userChatroom).Error; err != nil {
				http.Error(w, "Error creating user-chatroom association", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Error retrieving user-chatroom association", http.StatusInternalServerError)
			return
		}
	} else {
		if userChatroom.IsBanned {
			http.Error(w, "user is banned from this chatroom", http.StatusBadRequest)
			return
		}

		if userChatroom.IsJoined {
			http.Error(w, "user is already part of this chatroom", http.StatusBadRequest)
			return
		}

		userChatroom.Name = targetUser.Name
		userChatroom.IsInvited = true
		userChatroom.IsJoined = false
		userChatroom.InviteExpires = &expiration

		if err := config.DB.Save(&userChatroom).Error; err != nil {
			http.Error(w, "Error saving user-chatroom association", http.StatusInternalServerError)
			return
		}
	}

	var chatroom models.Chatroom
	if err := config.DB.Select("title").Where("id = ?", chatroomIdNum).First(&chatroom).Error; err != nil {
		http.Error(w, "Error fetching chatroom info", http.StatusInternalServerError)
		return
	}

	notification := models.Notification{
		UserId:     userChatroom.UserID,
		ChatroomId: uint(chatroomIdNum),
		Type:       "invite",
		SenderId:   admin.ID,
		Content:    fmt.Sprintf("You have been invited to join '%s' by %s", chatroom.Title, admin.Name),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ExpiresAt:  expiration,
	}
	if err := config.DB.Create(&notification).Error; err != nil {
		http.Error(w, "Error saving notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"Status":      "User invited to the chatroom",
		"User status": userChatroom,
	})
}

func KickUser(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value("isAdmin").(bool)
	userId := r.PathValue("userId")
	chatroomId := r.PathValue("id")

	if userId == "" {
		http.Error(w, "No valid user ID provided", http.StatusBadRequest)
		return
	}

	if chatroomId == "" {
		http.Error(w, "No valid chatroom ID provided", http.StatusBadRequest)
		return
	}

	if !isAdmin {
		http.Error(w, "You are not an admin", http.StatusUnauthorized)
		return
	}

	var userChatroom models.UserChatroom
	if err := config.DB.Where("user_id = ? AND chatroom_id = ?", userId, chatroomId).First(&userChatroom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "user-chatroom association not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving user-chatroom association", http.StatusInternalServerError)
		}
		return
	}

	if !userChatroom.IsJoined {
		http.Error(w, "User already not part of this chatroom", http.StatusBadRequest)
		return
	}
	userChatroom.IsJoined = false

	if err := config.DB.Save(&userChatroom).Error; err != nil {
		http.Error(w, "Error saving user-chatroom association", http.StatusInternalServerError)
		return
	}

	_ = config.DB.Create(&models.Notification{
		UserId:     userChatroom.UserID,
		ChatroomId: userChatroom.ChatroomID,
		Type:       "kick",
		SenderId:   r.Context().Value("userID").(uint),
		Content:    fmt.Sprintf("You have been kicked from chatroom %s", chatroomId),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}).Error

	json.NewEncoder(w).Encode(map[string]any{
		"Status":      "User kicked",
		"User status": userChatroom,
	})
}

func BanUser(w http.ResponseWriter, r *http.Request) {

	isAdmin := r.Context().Value("isAdmin").(bool)
	userId := r.PathValue("userId")
	chatroomId := r.PathValue("id")

	if userId == "" {
		http.Error(w, "No valid user ID provided", http.StatusBadRequest)
		return
	}

	if chatroomId == "" {
		http.Error(w, "No valid chatroom ID provided", http.StatusBadRequest)
		return
	}

	if !isAdmin {
		http.Error(w, "You are not an admin", http.StatusUnauthorized)
		return
	}

	var userChatroom models.UserChatroom
	if err := config.DB.Where("user_id = ? AND chatroom_id = ?", userId, chatroomId).First(&userChatroom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "user-chatroom association not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving user-chatroom association", http.StatusInternalServerError)
		}
		return
	}

	if userChatroom.IsBanned {
		http.Error(w, "user already banned", http.StatusBadRequest)
		return
	}

	userChatroom.IsBanned = true
	userChatroom.IsInvited = false
	userChatroom.IsJoined = false

	if err := config.DB.Save(&userChatroom).Error; err != nil {
		http.Error(w, "Error saving user-chatroom association", http.StatusInternalServerError)
		return
	}

	_ = config.DB.Create(&models.Notification{
		UserId:     userChatroom.UserID,
		ChatroomId: userChatroom.ChatroomID,
		Type:       "ban",
		SenderId:   r.Context().Value("userID").(uint),
		Content:    fmt.Sprintf("You have been banned from chatroom %s", chatroomId),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}).Error

	json.NewEncoder(w).Encode(map[string]any{
		"Status":      "User banned mn gher matrood!",
		"User status": userChatroom,
	})

}
