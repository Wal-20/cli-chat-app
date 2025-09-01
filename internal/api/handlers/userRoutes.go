package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


func GetUsers(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil || id <= 0 {

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(gin.H{"Status": "invalid ID"})
			return
		}

		var user models.User
		result := config.DB.First(&user, id)
		if result.Error != nil {

			w.WriteHeader(http.StatusNotFound)
			encoder.Encode(gin.H{"Status": "User not found"})
			return
		}

		w.WriteHeader(http.StatusOK)
		encoder.Encode(gin.H{
			"User": user,
		})

	} else {
		var users []models.User
		result := config.DB.Find(&users)

		if result.Error != nil {
			encoder.Encode(gin.H{"Message": "Failed to retrieve users"})
			return
		}

		encoder.Encode(gin.H{"Users": users})

	}
}

func GetChatroomsByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)

	var user models.User
	err := config.DB.Preload("Chatrooms").First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "No chatrooms found for user", http.StatusNotFound)		
		} else {
			http.Error(w, "Error retrieving Chatrooms", http.StatusInternalServerError)		
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{} {
		"User": user,
		"Chatrooms": user.Chatrooms,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	// Decode the request body into the user struct
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate user inputs
	if user.Name == "" || user.Password == "" {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// Fetch the user from the database
	var userDB models.User
	result := config.DB.Where("name = ?", user.Name).First(&userDB)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Check the password hash
	if !utils.CheckPasswordHash(user.Password, userDB.Password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Load token pair from storage
	tokenPair, err := utils.LoadTokenPair()
	if err == nil && tokenPair.AccessToken != "" {

		claims, err := utils.ValidateJWTToken(tokenPair.AccessToken)
		if err == nil && claims["userID"] == userDB.ID {
			// Check if the token belongs to the current user
			encoder.Encode(map[string] any {
				"Status":      "user already logged in",
				"AccessToken": tokenPair.AccessToken,
			})
			return
		} else if err != nil {
			tokenPair.AccessToken = ""
			tokenPair.RefreshToken = ""
		}
	}

	accessToken, err := utils.GenerateJWTToken(userDB.ID, userDB.Name)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	var refreshToken string

	if tokenPair.RefreshToken == "" {
		refreshToken, err := utils.GenerateRefreshToken(userDB.ID, userDB.Name )
		if err != nil {
			http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
			return
		}
		tokenPair.RefreshToken = refreshToken
	}

	tokenPair = utils.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: tokenPair.RefreshToken,
	}
	if err := utils.SaveTokenPair(tokenPair); err != nil {
		http.Error(w, "Error saving credentials", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	userDB.LastLogin = &now
	if err := config.DB.Save(&userDB).Error; err != nil {
		log.Printf("Failed to update LastLogin: %v", err)
		http.Error(w, "Failed to update LastLogin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]interface{}{
		"Status":       "success",
		"AccessToken":        accessToken,
		"RefreshToken": refreshToken,
	})
}


func LogOut(w http.ResponseWriter, r *http.Request) {
	
	tokenPair, err := utils.LoadTokenPair()
	if err != nil {
		http.Error(w, "Error loading token pair", http.StatusInternalServerError)
		return
	}

	tokenPair.AccessToken = ""
	tokenPair.RefreshToken = ""

	utils.AuthCache.Delete(tokenPair.AccessToken)
	utils.AuthCache.Delete(tokenPair.RefreshToken)

	if err := utils.SaveTokenPair(tokenPair); err != nil {
		http.Error(w, "Error saving token pair", http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)

	encoder.Encode(map[string]interface{}{
		"Status": "Logged out succcessfully",
	})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {

	tokenPair, err := utils.LoadTokenPair()
	if err != nil {
		http.Error(w, "Error loading token pair", http.StatusInternalServerError)
		return
	}

	tokenPair.AccessToken = ""
	tokenPair.RefreshToken = ""

	if err := utils.SaveTokenPair(tokenPair); err != nil {
		http.Error(w, "Error saving token pair", http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value("userId")

	if err := config.DB.Delete(&models.User{}, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user", http.StatusInternalServerError)
		}
		return
	}

	encoder := json.NewEncoder(w)

	encoder.Encode(map[string]interface{}{
		"Status": "User Deleted successfully",
	})

}


func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User

	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if user.Name == "" || user.Password == "" {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusBadRequest)
		return
	}

	user.Password = hashedPassword

	now := time.Now()
	user.LastLogin = &now

	result := config.DB.Create(&user)

	if result.Error != nil {
		http.Error(w, "Error creating user", http.StatusBadRequest)
		return
	}

	// Generate both access and refresh tokens
	accessToken, err := utils.GenerateJWTToken(user.ID, user.Name)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Name)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	// Save tokens to local tokenPair
	tokenPair := utils.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if err := utils.SaveTokenPair(tokenPair); err != nil {
		http.Error(w, "Error saving credentials", http.StatusInternalServerError)
		return
	}

	// Respond back with the user data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]interface{}{
		"Status": "User created successfully",
		"User":   user,
		"AccessToken": accessToken,
		"RefreshToken": refreshToken,
	})

}


func UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(w,"Unauthorized", http.StatusUnauthorized)
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

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	config.DB.Save(&user)

	encoder.Encode(map[string]interface{} {
		"Status": "User Updated successfully",
		"User": user,
	})
}

// admin actions
func InviteUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		http.Error(w, "No valid user ID provided", http.StatusBadRequest)
		return
	}

	chatroomId := r.PathValue("id")
	if chatroomId == "" {
		http.Error(w, "No valid chatroom ID provided", http.StatusBadRequest)
		return
	}

	userIdNum, err := strconv.ParseUint(userId, 10, 32) // 10 = base 10, 32-bit conversion
	if err != nil {
		http.Error(w, "Invalid user ID, cannot convert to number", http.StatusBadRequest)
		return
	}

	chatroomIdNum, err := strconv.ParseUint(userId, 10, 32) // 10 = base 10, 32-bit conversion
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
	if err := config.DB.Where(r.Context().Value("userID")).First(&admin).Error; err != nil {
		http.Error(w, "Error fetching admin", http.StatusInternalServerError)
		return
	}

	var userChatroom models.UserChatroom
	if err := config.DB.Where("user_id = ? AND chatroom_id = ?", userId, chatroomId).First(&userChatroom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			exp_time := time.Now().Add(7 * 24 * time.Hour) // invite expires in 7 days
			newUserChatroom := models.UserChatroom {
				UserID:       uint(userIdNum),
				ChatroomID:   uint(chatroomIdNum),
				LastJoinTime: nil, 
				IsInvited: true,
				IsJoined: false,
				InviteExpires: &exp_time,
			}		
			if err := config.DB.Create(&newUserChatroom).Error; err != nil {
				http.Error(w, "Error adding user to chatroom", http.StatusInternalServerError)
				return
			}

		} else {
			http.Error(w, "Error retrieving user-chatroom association", http.StatusInternalServerError)
		}
		return
	}

	if userChatroom.IsBanned {
		http.Error(w, "user is banned from this chatroom", http.StatusBadRequest)
		return
	}

	if userChatroom.IsJoined {
		http.Error(w, "user is already part of this chatroom", http.StatusBadRequest)
		return
	}

	exp_time := time.Now().Add(7 * 24 * time.Hour)
	userChatroom.IsInvited = true
	userChatroom.InviteExpires = &exp_time 

	notification := models.Notification {
		UserId:    userChatroom.UserID, 
		Content:   fmt.Sprintf("You have been invited to a chatroom by user %s", admin.Name), 
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Set TTL to 7 days
	}
	if err := config.DB.Save(&notification).Error; err != nil {
		http.Error(w, "Error saving notification", http.StatusInternalServerError)
		return
	}

	if err := config.DB.Save(&userChatroom).Error; err != nil {
		http.Error(w, "Error saving user-chatroom association", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{} {
		"Status": "User Invited to the chatroom",
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

	json.NewEncoder(w).Encode(map[string]interface{} {
		"Status": "User kicked",
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

	json.NewEncoder(w).Encode(map[string]interface{} {
		"Status": "User banned mn gher matrood!",
		"User status": userChatroom,
	})

}

